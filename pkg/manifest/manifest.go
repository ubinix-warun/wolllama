// Package manifest defines the Wolllama manifest schema — the JSON index
// that maps an Ollama model's blobs to their Walrus object IDs.
// The manifest itself is stored on Walrus; its own object ID is the
// unique identifier for a published model.
package manifest

import (
	"encoding/json"
	"fmt"
	"time"
)

// WolllamaVersion is the current manifest schema version.
const WolllamaVersion = 1

// WolllamaManifest is the top-level manifest stored on Walrus.
type WolllamaManifest struct {
	WolllamaVersion int              `json:"wolllamaVersion"`
	Name            string           `json:"name"`            // "model:tag", e.g. "llama3.2:3b-q4_K_M"
	OllamaManifest  json.RawMessage  `json:"ollamaManifest"`  // Original Ollama manifest JSON (opaque)
	Blobs           map[string]string `json:"blobs"`          // sha256 digest → walrus object ID
	CreatedAt       time.Time        `json:"createdAt"`
}

// OllamaManifest is the parsed structure of Ollama's internal manifest JSON.
// We embed the raw JSON for fidelity, but parse enough to extract metadata.
type OllamaManifest struct {
	SchemaVersion int              `json:"schemaVersion"`
	Config        *OllamaBlobRef   `json:"config,omitempty"`
	Layers        []OllamaBlobRef  `json:"layers,omitempty"`
}

// OllamaBlobRef references a blob in the Ollama manifest.
type OllamaBlobRef struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}

// ModelSummary is a human-friendly summary extracted from a manifest.
type ModelSummary struct {
	Name       string
	Tag        string
	BlobCount  int
	TotalSize  int64
	CreatedAt  time.Time
	Blobs      []BlobInfo
}

// BlobInfo describes a single blob in the manifest.
type BlobInfo struct {
	Digest       string
	MediaType    string
	Size         int64
	WalrusObjID  string
}

// Parse extracts metadata from the raw manifest without requiring the caller
// to understand the internal Ollama manifest structure.
func (m *WolllamaManifest) Parse() (*ModelSummary, error) {
	var om OllamaManifest
	if err := json.Unmarshal(m.OllamaManifest, &om); err != nil {
		return nil, fmt.Errorf("parse ollama manifest: %w", err)
	}

	summary := &ModelSummary{
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
		Blobs:     make([]BlobInfo, 0, len(m.Blobs)),
	}

	// Extract tag from name ("model:tag" → "tag")
	if idx := lastIndexByte(m.Name, ':'); idx >= 0 {
		summary.Tag = m.Name[idx+1:]
	}

	var totalSize int64

	// Config blob
	if om.Config != nil {
		objID := m.Blobs[om.Config.Digest]
		summary.Blobs = append(summary.Blobs, BlobInfo{
			Digest:      om.Config.Digest,
			MediaType:   om.Config.MediaType,
			Size:        om.Config.Size,
			WalrusObjID: objID,
		})
		totalSize += om.Config.Size
	}

	// Layer blobs
	for _, layer := range om.Layers {
		objID := m.Blobs[layer.Digest]
		summary.Blobs = append(summary.Blobs, BlobInfo{
			Digest:      layer.Digest,
			MediaType:   layer.MediaType,
			Size:        layer.Size,
			WalrusObjID: objID,
		})
		totalSize += layer.Size
	}

	summary.BlobCount = len(summary.Blobs)
	summary.TotalSize = totalSize

	return summary, nil
}

// Validate checks that the manifest has all required fields and the blob
// map covers every blob referenced in the Ollama manifest.
func (m *WolllamaManifest) Validate() error {
	if m.WolllamaVersion != WolllamaVersion {
		return fmt.Errorf("unsupported manifest version %d (expected %d)", m.WolllamaVersion, WolllamaVersion)
	}
	if m.Name == "" {
		return fmt.Errorf("manifest name is empty")
	}
	if len(m.Blobs) == 0 {
		return fmt.Errorf("manifest has no blobs")
	}

	var om OllamaManifest
	if err := json.Unmarshal(m.OllamaManifest, &om); err != nil {
		return fmt.Errorf("ollama manifest is not valid JSON: %w", err)
	}

	// Every blob referenced in the Ollama manifest must have a Walrus mapping
	if om.Config != nil {
		if _, ok := m.Blobs[om.Config.Digest]; !ok {
			return fmt.Errorf("config blob %s not in blobs map", om.Config.Digest)
		}
	}
	for _, layer := range om.Layers {
		if _, ok := m.Blobs[layer.Digest]; !ok {
			return fmt.Errorf("layer blob %s not in blobs map", layer.Digest)
		}
	}

	return nil
}

// New creates a WolllamaManifest with defaults.
func New(name string, ollamaManifest json.RawMessage, blobs map[string]string) *WolllamaManifest {
	return &WolllamaManifest{
		WolllamaVersion: WolllamaVersion,
		Name:            name,
		OllamaManifest:  ollamaManifest,
		Blobs:           blobs,
		CreatedAt:       time.Now().UTC(),
	}
}

func lastIndexByte(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}
