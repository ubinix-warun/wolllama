package manifest

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewManifest_RoundTrip(t *testing.T) {
	ollamaJSON := json.RawMessage(`{
		"schemaVersion": 2,
		"config": {
			"mediaType": "application/vnd.ollama.image.model",
			"digest": "sha256:abc123",
			"size": 1024
		},
		"layers": [
			{
				"mediaType": "application/vnd.ollama.image.model",
				"digest": "sha256:def456",
				"size": 4000000000
			}
		]
	}`)

	blobs := map[string]string{
		"sha256:abc123": "walrus_objid_AAA",
		"sha256:def456": "walrus_objid_BBB",
	}

	m := New("llama3.2:3b-q4_K_M", ollamaJSON, blobs)

	if m.WolllamaVersion != 1 {
		t.Errorf("expected version 1, got %d", m.WolllamaVersion)
	}
	if m.Name != "llama3.2:3b-q4_K_M" {
		t.Errorf("expected name 'llama3.2:3b-q4_K_M', got %q", m.Name)
	}
	if len(m.Blobs) != 2 {
		t.Errorf("expected 2 blobs, got %d", len(m.Blobs))
	}

	// JSON round-trip
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m2 WolllamaManifest
	if err := json.Unmarshal(data, &m2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if m2.Name != m.Name {
		t.Errorf("name mismatch: %q vs %q", m2.Name, m.Name)
	}
}

func TestParse_ExtractsSummary(t *testing.T) {
	ollamaJSON := json.RawMessage(`{
		"schemaVersion": 2,
		"config": {
			"mediaType": "application/vnd.ollama.image.model",
			"digest": "sha256:abc123",
			"size": 1024
		},
		"layers": [
			{
				"mediaType": "application/vnd.ollama.image.model",
				"digest": "sha256:def456",
				"size": 4000000000
			}
		]
	}`)

	blobs := map[string]string{
		"sha256:abc123": "walrus_objid_AAA",
		"sha256:def456": "walrus_objid_BBB",
	}

	m := New("llama3.2:3b-q4_K_M", ollamaJSON, blobs)

	summary, err := m.Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if summary.Name != "llama3.2:3b-q4_K_M" {
		t.Errorf("name: got %q", summary.Name)
	}
	if summary.Tag != "3b-q4_K_M" {
		t.Errorf("tag: got %q", summary.Tag)
	}
	if summary.BlobCount != 2 {
		t.Errorf("blob count: got %d", summary.BlobCount)
	}
	if summary.TotalSize != 4000001024 {
		t.Errorf("total size: got %d", summary.TotalSize)
	}
}

func TestValidate_AllBlobsMapped(t *testing.T) {
	ollamaJSON := json.RawMessage(`{
		"schemaVersion": 2,
		"config": { "mediaType": "x", "digest": "sha256:a", "size": 1 }
	}`)

	// Missing config blob mapping
	m := New("model", ollamaJSON, map[string]string{})
	m.CreatedAt = time.Now()

	if err := m.Validate(); err == nil {
		t.Error("expected validation error for missing blob mapping")
	}
}

func TestValidate_MissingName(t *testing.T) {
	m := New("", json.RawMessage(`{}`), map[string]string{"x": "y"})
	m.CreatedAt = time.Now()

	if err := m.Validate(); err == nil {
		t.Error("expected validation error for empty name")
	}
}
