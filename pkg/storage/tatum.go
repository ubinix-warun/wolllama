package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

const (
	tatumMaxChunkSize  = 45 * 1024 * 1024 // 45 MiB — safe margin under Tatum's 50 MiB limit
	tatumDefaultAPIURL = "https://api.tatum.io"
	tatumPollInterval  = 5 * time.Second
	tatumPollTimeout   = 10 * time.Minute
)

// TatumProvider uploads blobs to Walrus through Tatum's managed storage gateway.
type TatumProvider struct {
	apiKey string
	apiURL string
	client *http.Client
}

// NewTatumProvider creates a TatumProvider from config.
func NewTatumProvider(cfg Config) (*TatumProvider, error) {
	if cfg.TatumAPIKey == "" {
		return nil, fmt.Errorf("Tatum API key is required (set tatum_api_key in config or WOLLLAMA_TATUM_API_KEY env var)")
	}
	apiURL := cfg.TatumAPIURL
	if apiURL == "" {
		apiURL = tatumDefaultAPIURL
	}
	return &TatumProvider{
		apiKey: cfg.TatumAPIKey,
		apiURL: apiURL,
		client: &http.Client{Timeout: 10 * time.Minute},
	}, nil
}

func (t *TatumProvider) Name() string       { return "tatum" }
func (t *TatumProvider) MaxChunkSize() int64 { return tatumMaxChunkSize }

func (t *TatumProvider) Upload(data []byte) (string, error) {
	if int64(len(data)) > tatumMaxChunkSize {
		return "", fmt.Errorf("blob size %d exceeds Tatum max %d bytes", len(data), tatumMaxChunkSize)
	}

	// Build multipart form
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", "blob")
	if err != nil {
		return "", fmt.Errorf("create form: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return "", fmt.Errorf("write form: %w", err)
	}
	w.Close()

	// Upload
	req, err := http.NewRequest("POST", t.apiURL+"/v4/data/storage/upload", &buf)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("x-api-key", t.apiKey)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("tatum upload: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return "", fmt.Errorf("tatum upload failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		JobID  string `json:"jobId"`
		BlobID string `json:"blobId"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse tatum response: %w (body: %s)", err, string(body))
	}

	if result.BlobID == "" {
		return "", fmt.Errorf("tatum response missing blobId: %s", string(body))
	}

	// Poll until certified — the Walrus object ID may differ from Tatum's blobId.
	// After certification, Tatum returns the actual Walrus object ID.
	walrusID, err := t.pollUntilCertified(result.JobID)
	if err != nil {
		return "", fmt.Errorf("tatum certification: %w", err)
	}

	return walrusID, nil
}

// pollUntilCertified polls the job status until certified and returns the Walrus object ID.
func (t *TatumProvider) pollUntilCertified(jobID string) (string, error) {
	deadline := time.Now().Add(tatumPollTimeout)
	attempts := 0

	for time.Now().Before(deadline) {
		time.Sleep(tatumPollInterval)
		attempts++

		if attempts == 1 {
			fmt.Fprintf(os.Stderr, "  waiting for Walrus certification")
		} else {
			fmt.Fprintf(os.Stderr, ".")
		}

		req, err := http.NewRequest("GET", t.apiURL+"/v4/data/storage/upload/"+jobID, nil)
		if err != nil {
			continue
		}
		req.Header.Set("x-api-key", t.apiKey)

		resp, err := t.client.Do(req)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != 200 && resp.StatusCode != 201 {
			continue
		}

		var status struct {
			Status        string `json:"status"`
			BlobID        string `json:"blobId"`
			QuiltPatchID  string `json:"quiltPatchId"`
			SuiObjectID   string `json:"suiObjectId"`
		}
		if err := json.Unmarshal(body, &status); err != nil {
			continue
		}

		if status.Status == "DONE" || status.Status == "CERTIFIED" || status.Status == "COMPLETED" {
			fmt.Fprintf(os.Stderr, " certified\n")
			// Use quiltPatchId for downloads — it resolves to the raw content via
			// /v1/blobs/by-quilt-patch-id/{quiltPatchId} on the Walrus aggregator.
			if status.QuiltPatchID != "" {
				return status.QuiltPatchID, nil
			}
			return status.BlobID, nil
		}
	}

	fmt.Fprintf(os.Stderr, " timed out\n")
	return "", fmt.Errorf("certification timed out after %v", tatumPollTimeout)
}
