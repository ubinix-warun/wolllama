// Package ollama reads and writes Ollama's local model store on disk.
// It reads the manifest registry and blob files directly — no daemon required.
package ollama

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// DefaultOllamaHome is the standard Ollama data directory.
	DefaultOllamaHome = "~/.ollama"

	// Ollama manifest registry path relative to OLLAMA_HOME.
	manifestsRelPath = "models/manifests/registry.ollama.ai/library"
	blobsRelPath     = "models/blobs"
)

// Store provides read/write access to an Ollama model store on disk.
type Store struct {
	home string // expanded path to ~/.ollama
}

// NewStore creates a Store. If home is empty, DefaultOllamaHome is used.
func NewStore(home string) (*Store, error) {
	if home == "" {
		home = DefaultOllamaHome
	}
	expanded, err := expandHome(home)
	if err != nil {
		return nil, fmt.Errorf("expand ollama home: %w", err)
	}
	return &Store{home: expanded}, nil
}

// Home returns the expanded Ollama home directory.
func (s *Store) Home() string { return s.home }

// ManifestsDir returns the directory containing Ollama manifest JSON files.
func (s *Store) ManifestsDir() string {
	return filepath.Join(s.home, manifestsRelPath)
}

// BlobsDir returns the directory containing sha256-addressed blob files.
func (s *Store) BlobsDir() string {
	return filepath.Join(s.home, blobsRelPath)
}

// ManifestEntry is the JSON structure Ollama uses to map a model:tag to
// its config and layers.
type ManifestEntry struct {
	SchemaVersion int            `json:"schemaVersion"`
	Config        *BlobRef       `json:"config,omitempty"`
	Layers        []BlobRef      `json:"layers,omitempty"`
	Raw           json.RawMessage `json:"-"` // original bytes
}

// BlobRef references a blob by digest, media type, and size.
type BlobRef struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}

// ReadManifest reads the Ollama manifest for a given model:tag.
// modelTag is like "llama3.2:3b-q4_K_M" or "llama3.2:latest".
func (s *Store) ReadManifest(modelTag string) (*ManifestEntry, error) {
	// Split "model:tag" into path components
	model, tag, err := splitModelTag(modelTag)
	if err != nil {
		return nil, err
	}

	manifestPath := filepath.Join(s.ManifestsDir(), model, tag)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("model %q not found in Ollama store (looked at %s)", modelTag, manifestPath)
		}
		return nil, fmt.Errorf("read manifest %s: %w", manifestPath, err)
	}

	var entry ManifestEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("parse manifest %s: %w", manifestPath, err)
	}
	entry.Raw = data

	return &entry, nil
}

// ReadBlob reads a blob file from the Ollama blobs directory.
func (s *Store) ReadBlob(digest string) ([]byte, error) {
	// Digest format: "sha256:abc123..."
	blobPath := filepath.Join(s.BlobsDir(), blobFilename(digest))
	data, err := os.ReadFile(blobPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("blob %s not found at %s", digest, blobPath)
		}
		return nil, fmt.Errorf("read blob %s: %w", digest, err)
	}
	return data, nil
}

// BlobPath returns the filesystem path for a given blob digest.
func (s *Store) BlobPath(digest string) string {
	return filepath.Join(s.BlobsDir(), blobFilename(digest))
}

// WriteManifest writes an Ollama manifest entry to the store.
func (s *Store) WriteManifest(modelTag string, entry *ManifestEntry) error {
	model, tag, err := splitModelTag(modelTag)
	if err != nil {
		return err
	}

	manifestPath := filepath.Join(s.ManifestsDir(), model, tag)
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		return fmt.Errorf("create manifest dir: %w", err)
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		return fmt.Errorf("write manifest %s: %w", manifestPath, err)
	}
	return nil
}

// WriteBlob writes raw bytes to the blobs directory.
func (s *Store) WriteBlob(digest string, data []byte) error {
	blobPath := filepath.Join(s.BlobsDir(), blobFilename(digest))
	if err := os.MkdirAll(filepath.Dir(blobPath), 0o755); err != nil {
		return fmt.Errorf("create blobs dir: %w", err)
	}
	if err := os.WriteFile(blobPath, data, 0o644); err != nil {
		return fmt.Errorf("write blob %s: %w", digest, err)
	}
	return nil
}

// BlobFilename converts a digest like "sha256:abc123..." to the filename
// Ollama uses: "sha256-abc123..."
func blobFilename(digest string) string {
	if len(digest) > 7 && digest[:7] == "sha256:" {
		return "sha256-" + digest[7:]
	}
	return digest
}

// splitModelTag splits "llama3.2:3b-q4_K_M" into ("llama3.2", "3b-q4_K_M").
func splitModelTag(modelTag string) (model, tag string, err error) {
	idx := lastIndexByte(modelTag, ':')
	if idx < 0 {
		return "", "", fmt.Errorf("model tag %q must include a tag (e.g. 'model:latest')", modelTag)
	}
	return modelTag[:idx], modelTag[idx+1:], nil
}

// expandHome replaces a leading "~" with the user's home directory.
func expandHome(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	if len(path) == 1 {
		return home, nil
	}
	return filepath.Join(home, path[2:]), nil // skip "~/"
}

// ListEntry represents a single model:tag found in the Ollama manifest registry.
type ListEntry struct {
	Model     string
	Tag       string
	ModelTag  string    // "model:tag"
	ID        string    // first 12 hex chars of manifest sha256
	Size      int64     // total size of all blobs in bytes
	BlobCount int
	Modified  time.Time // manifest file modification time
}

// ListModels walks the Ollama manifest registry and returns all available models.
func (s *Store) ListModels() ([]ListEntry, error) {
	manifestsDir := s.ManifestsDir()

	var entries []ListEntry

	// Walk registry.ollama.ai/library/<model>/<tag>
	modelDirs, err := os.ReadDir(manifestsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No models yet
		}
		return nil, fmt.Errorf("read manifests dir %s: %w", manifestsDir, err)
	}

	for _, modelDir := range modelDirs {
		if !modelDir.IsDir() {
			continue
		}
		model := modelDir.Name()

		tagFiles, err := os.ReadDir(filepath.Join(manifestsDir, model))
		if err != nil {
			continue // skip unreadable model dirs
		}

		for _, tagFile := range tagFiles {
			if tagFile.IsDir() {
				continue
			}
			tag := tagFile.Name()
			manifestPath := filepath.Join(manifestsDir, model, tag)

			entry, err := s.ReadManifest(model + ":" + tag)
			if err != nil {
				continue // skip unreadable manifests
			}

			// Compute ID: first 12 hex chars of sha256 of raw manifest bytes
			h := sha256.Sum256(entry.Raw)
			id := hex.EncodeToString(h[:])[:12]

			// Get modification time from the manifest file
			info, err := os.Stat(manifestPath)
			var modTime time.Time
			if err == nil {
				modTime = info.ModTime()
			}

			var totalSize int64
			blobCount := 0
			if entry.Config != nil {
				totalSize += entry.Config.Size
				blobCount++
			}
			for _, layer := range entry.Layers {
				totalSize += layer.Size
				blobCount++
			}

			entries = append(entries, ListEntry{
				Model:     model,
				Tag:       tag,
				ModelTag:  model + ":" + tag,
				ID:        id,
				Size:      totalSize,
				BlobCount: blobCount,
				Modified:  modTime,
			})
		}
	}

	return entries, nil
}

func lastIndexByte(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}
