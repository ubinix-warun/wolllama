// Package walrus provides a thin convenience wrapper around
// github.com/namihq/walrus-go for wolllama-specific operations.
package walrus

import (
	"fmt"
	"io"
	"net/http"

	walrusgo "github.com/namihq/walrus-go"
)

// DefaultChunkSize is the max single-request download size for Walrus aggregators.
// The public aggregator caps at 500 MB; we use 256 MB chunks for safety.
const DefaultChunkSize = 256 * 1024 * 1024 // 256 MB

// Client wraps the walrus-go SDK client with wolllama-specific defaults.
type Client struct {
	inner          *walrusgo.Client
	aggregatorURLs []string
	publisherURLs  []string
	httpClient     *http.Client
}

// Config holds Walrus connection parameters.
type Config struct {
	PublisherURLs  []string
	AggregatorURLs []string
}

// NewClient creates a Client with the given config.
func NewClient(cfg Config) *Client {
	opts := []walrusgo.ClientOption{}
	if len(cfg.PublisherURLs) > 0 {
		opts = append(opts, walrusgo.WithPublisherURLs(cfg.PublisherURLs))
	}
	if len(cfg.AggregatorURLs) > 0 {
		opts = append(opts, walrusgo.WithAggregatorURLs(cfg.AggregatorURLs))
	}
	return &Client{
		inner:          walrusgo.NewClient(opts...),
		aggregatorURLs: cfg.AggregatorURLs,
		publisherURLs:  cfg.PublisherURLs,
		httpClient:     &http.Client{},
	}
}

// StoreBlob uploads data and returns the blob's object ID.
func (c *Client) StoreBlob(data []byte, epochs int) (string, error) {
	resp, err := c.inner.Store(data, &walrusgo.StoreOptions{Epochs: epochs})
	if err != nil {
		return "", fmt.Errorf("store blob: %w", err)
	}
	if resp.NewlyCreated != nil {
		return resp.NewlyCreated.BlobObject.BlobID, nil
	}
	if resp.AlreadyCertified != nil {
		return resp.AlreadyCertified.BlobID, nil
	}
	return "", fmt.Errorf("unexpected store response: no blob ID")
}

// StoreBlobFromReader uploads data from a reader and returns the blob's object ID.
func (c *Client) StoreBlobFromReader(r io.Reader, epochs int) (string, error) {
	resp, err := c.inner.StoreFromReader(r, &walrusgo.StoreOptions{Epochs: epochs})
	if err != nil {
		return "", fmt.Errorf("store blob: %w", err)
	}
	if resp.NewlyCreated != nil {
		return resp.NewlyCreated.BlobObject.BlobID, nil
	}
	if resp.AlreadyCertified != nil {
		return resp.AlreadyCertified.BlobID, nil
	}
	return "", fmt.Errorf("unexpected store response: no blob ID")
}

// StoreFile uploads a file and returns its blob object ID.
func (c *Client) StoreFile(path string, epochs int) (string, error) {
	resp, err := c.inner.StoreFile(path, &walrusgo.StoreOptions{Epochs: epochs})
	if err != nil {
		return "", fmt.Errorf("store file: %w", err)
	}
	if resp.NewlyCreated != nil {
		return resp.NewlyCreated.BlobObject.BlobID, nil
	}
	if resp.AlreadyCertified != nil {
		return resp.AlreadyCertified.BlobID, nil
	}
	return "", fmt.Errorf("unexpected store response: no blob ID")
}

// ReadBlobByQuiltID downloads a blob using the Walrus quilt-id endpoint.
// Tatum-uploaded blobs are accessible at /v1/blobs/by-quilt-id/{blobId}/blob
// even though the regular /v1/blobs/{blobId} returns a binary wrapper.
func (c *Client) ReadBlobByQuiltID(blobID string) ([]byte, error) {
	if len(c.aggregatorURLs) == 0 {
		return nil, fmt.Errorf("no aggregator URLs configured")
	}

	url := fmt.Sprintf("%s/v1/blobs/by-quilt-id/%s/blob", c.aggregatorURLs[0], blobID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("quilt-id request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("quilt-id download failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// ReadBlobByQuiltPatchID downloads a blob using the Walrus quilt-patch endpoint.
// This is used for blobs uploaded through Tatum's gateway, which wraps content
// in a binary format on the regular /v1/blobs/{id} endpoint. The quilt-patch
// endpoint returns the raw unwrapped content.
func (c *Client) ReadBlobByQuiltPatchID(quiltPatchID string) ([]byte, error) {
	if len(c.aggregatorURLs) == 0 {
		return nil, fmt.Errorf("no aggregator URLs configured")
	}

	url := fmt.Sprintf("%s/v1/blobs/by-quilt-patch-id/%s", c.aggregatorURLs[0], quiltPatchID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("quilt-patch request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("quilt-patch download failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// ReadBlobWithFallback tries multiple endpoints to handle both native Walrus
// blobs and Tatum-uploaded blobs. Fallback order:
//   1. Regular /v1/blobs/{id}
//   2. Quilt-patch /v1/blobs/by-quilt-patch-id/{id}  (Tatum quiltPatchIds with suffix)
//   3. Quilt-id /v1/blobs/by-quilt-id/{id}/blob       (Tatum plain blobIds)
func (c *Client) ReadBlobWithFallback(blobID string) ([]byte, error) {
	data, err := c.ReadBlob(blobID)
	if err != nil || isTatumWrapper(data) {
		// Try quilt-patch (full quiltPatchId with suffix like ...BAQAEAA)
		data2, err2 := c.ReadBlobByQuiltPatchID(blobID)
		if err2 == nil && !isTatumWrapper(data2) {
			return data2, nil
		}
		// Try quilt-id (plain blobId — works for Tatum blobs without suffix)
		data3, err3 := c.ReadBlobByQuiltID(blobID)
		if err3 == nil && !isTatumWrapper(data3) {
			return data3, nil
		}
		// Return original error if all fallbacks failed
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func isTatumWrapper(data []byte) bool {
	return len(data) > 0 && data[0] == 0x01
}

// ReadBlob downloads a blob by its object ID. For blobs under 500 MB,
// this uses a single request. For larger blobs, use ReadLargeBlob.
func (c *Client) ReadBlob(blobID string) ([]byte, error) {
	data, err := c.inner.Read(blobID, nil)
	if err != nil {
		return nil, fmt.Errorf("read blob %s: %w", blobID, err)
	}
	return data, nil
}

// ReadLargeBlob downloads a blob that exceeds the aggregator's 500 MB single-request
// limit. It tries the publisher endpoint first (which serves both reads and writes),
// then falls back to chunked Range requests on the aggregator.
func (c *Client) ReadLargeBlob(blobID string, totalSize int64) ([]byte, error) {
	// Collect all URLs to try: publisher first (no size limit for uploads, may serve reads),
	// then aggregator fallback with chunked Range requests.
	var urls []string
	urls = append(urls, c.publisherURLs...)
	urls = append(urls, c.aggregatorURLs...)

	if len(urls) == 0 {
		return nil, fmt.Errorf("no publisher or aggregator URLs configured")
	}

	// Try a single full download from each URL first (publisher may allow it)
	for _, baseURL := range urls {
		data, err := c.downloadFull(baseURL, blobID)
		if err == nil {
			return data, nil
		}
	}

	// Fallback: chunked Range requests on aggregator
	if len(c.aggregatorURLs) > 0 {
		result := make([]byte, totalSize)
		chunkSize := int64(DefaultChunkSize)

		for offset := int64(0); offset < totalSize; offset += chunkSize {
			end := offset + chunkSize - 1
			if end >= totalSize {
				end = totalSize - 1
			}

			chunk, err := c.downloadRange(c.aggregatorURLs[0], blobID, offset, end)
			if err != nil {
				return nil, fmt.Errorf("download range %d-%d: %w", offset, end, err)
			}

			copy(result[offset:], chunk)
		}
		return result, nil
	}

	return nil, fmt.Errorf("failed to download large blob %s from any endpoint", blobID)
}

// downloadFull fetches the complete blob from a specific base URL.
func (c *Client) downloadFull(baseURL, blobID string) ([]byte, error) {
	url := fmt.Sprintf("%s/v1/blobs/%s", baseURL, blobID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// downloadRange fetches a byte range from a specific base URL.
func (c *Client) downloadRange(baseURL, blobID string, start, end int64) ([]byte, error) {
	url := fmt.Sprintf("%s/v1/blobs/%s", baseURL, blobID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	return data, nil
}

// ReadBlobToFile downloads a blob and writes it to a local file.
func (c *Client) ReadBlobToFile(blobID, filePath string) error {
	if err := c.inner.ReadToFile(blobID, filePath, nil); err != nil {
		return fmt.Errorf("read blob %s to file %s: %w", blobID, filePath, err)
	}
	return nil
}

// HeadBlob returns metadata for a blob without downloading its content.
func (c *Client) HeadBlob(blobID string) (*walrusgo.BlobMetadata, error) {
	meta, err := c.inner.Head(blobID)
	if err != nil {
		return nil, fmt.Errorf("head blob %s: %w", blobID, err)
	}
	return meta, nil
}
