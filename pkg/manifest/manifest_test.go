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

	blobs := map[string]BlobRef{
		"sha256:abc123": {Single: "walrus_objid_AAA"},
		"sha256:def456": {Single: "walrus_objid_BBB"},
	}

	m := New("llama3.2:3b-q4_K_M", ollamaJSON, blobs)

	if m.WolllamaVersion != 2 {
		t.Errorf("expected version 2, got %d", m.WolllamaVersion)
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

	if m2.Blobs["sha256:abc123"].Single != "walrus_objid_AAA" {
		t.Errorf("blob ref mismatch")
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

	blobs := map[string]BlobRef{
		"sha256:abc123": {Single: "walrus_objid_AAA"},
		"sha256:def456": {Single: "walrus_objid_BBB"},
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
	if len(summary.Blobs[0].WalrusIDs) != 1 || summary.Blobs[0].WalrusIDs[0] != "walrus_objid_AAA" {
		t.Errorf("walrus IDs mismatch")
	}
}

func TestChunkedBlobRef(t *testing.T) {
	ollamaJSON := json.RawMessage(`{
		"schemaVersion": 2,
		"config": { "mediaType": "x", "digest": "sha256:a", "size": 1 },
		"layers": [
			{ "mediaType": "x", "digest": "sha256:b", "size": 2000000000 }
		]
	}`)

	blobs := map[string]BlobRef{
		"sha256:a": {Single: "objid_A"},
		"sha256:b": {Chunks: []string{"chunk_1", "chunk_2", "chunk_3"}},
	}

	m := New("model:tag", ollamaJSON, blobs)

	summary, err := m.Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if summary.BlobCount != 2 {
		t.Errorf("blob count: got %d", summary.BlobCount)
	}

	// Config blob: single
	if len(summary.Blobs[0].WalrusIDs) != 1 || summary.Blobs[0].WalrusIDs[0] != "objid_A" {
		t.Errorf("config blob IDs mismatch: %v", summary.Blobs[0].WalrusIDs)
	}

	// Layer blob: chunked
	if len(summary.Blobs[1].WalrusIDs) != 3 {
		t.Errorf("layer blob should have 3 chunk IDs, got %d", len(summary.Blobs[1].WalrusIDs))
	}

	// JSON round-trip with chunks
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal chunked: %v", err)
	}

	var m2 WolllamaManifest
	if err := json.Unmarshal(data, &m2); err != nil {
		t.Fatalf("unmarshal chunked: %v", err)
	}

	if !m2.Blobs["sha256:b"].IsChunked() {
		t.Error("expected chunked blob ref after round-trip")
	}
	if len(m2.Blobs["sha256:b"].Chunks) != 3 {
		t.Errorf("expected 3 chunks after round-trip, got %d", len(m2.Blobs["sha256:b"].Chunks))
	}
}

func TestValidate_AllBlobsMapped(t *testing.T) {
	ollamaJSON := json.RawMessage(`{
		"schemaVersion": 2,
		"config": { "mediaType": "x", "digest": "sha256:a", "size": 1 }
	}`)

	m := New("model", ollamaJSON, map[string]BlobRef{})
	m.CreatedAt = time.Now()

	if err := m.Validate(); err == nil {
		t.Error("expected validation error for missing blob mapping")
	}
}

func TestValidate_MissingName(t *testing.T) {
	m := New("", json.RawMessage(`{}`), map[string]BlobRef{"x": {Single: "y"}})
	m.CreatedAt = time.Now()

	if err := m.Validate(); err == nil {
		t.Error("expected validation error for empty name")
	}
}
