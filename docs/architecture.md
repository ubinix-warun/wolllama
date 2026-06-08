# Wolllama Architecture

## System Overview

```
                    ┌─────────────────┐
                    │   wolllama CLI   │
                    │   (Go, cobra)    │
                    └────┬───┬────┬────┘
                         │   │    │
              ┌──────────┘   │    └──────────┐
              ▼              ▼               ▼
        ┌──────────┐  ┌──────────┐   ┌──────────┐
        │  Ollama  │  │  Walrus  │   │  Tatum   │
        │  Store   │  │ Publisher│   │ Gateway  │
        │  (disk)  │  │ / Aggreg.│   │  (HTTP)  │
        └──────────┘  └──────────┘   └──────────┘
                             ▲
        ┌────────────────────┘
        │
  ┌─────┴──────┐     ┌──────────┐     ┌──────────┐
  │ wolllama   │────▶│  SQLite  │◀────│ wolllama │
  │ API (Go)   │     │  (models │     │ Site     │
  │            │     │   users) │     │ (React)  │
  └────────────┘     └──────────┘     └──────────┘
```

## Component Details

### CLI (`cli/`)

The wolllama CLI is a Go binary that acts as a sidecar to Ollama. It reads and writes
directly to `~/.ollama/models/` — no Ollama daemon required for push/pull.

**Commands:**

| Command | Description |
|---------|-------------|
| `push <model:tag>` | Read from Ollama store → upload blobs to provider → store wolllama manifest |
| `pull <obj-id>` | Fetch wolllama manifest → download blobs → write to Ollama store |
| `show <obj-id>` | Display model summary from manifest |
| `list` | List models in Ollama store + cached wolllama manifests |
| `config [show|set]` | Manage Walrus endpoints, provider, epochs, network |

**Push flow:**
1. Parse `~/.ollama/models/manifests/registry.ollama.ai/library/<model>/<tag>` — JSON with config + layers
2. For each blob (config + layers): read from `~/.ollama/models/blobs/sha256-<digest>`
3. If blob > provider MaxChunkSize: split into chunks, upload each
4. Upload via selected storage provider (Walrus publisher or Tatum gateway)
5. Build wolllama manifest JSON (blob digest → Walrus/Tatum object IDs)
6. Upload manifest to provider → print manifest object ID

**Pull flow:**
1. Fetch wolllama manifest via provider (with quilt-patch fallback for Tatum)
2. For each blob reference: download (quilt-patch first if Tatum, then regular)
3. If chunked: download all chunks, reassemble
4. SHA256-verify each blob against its digest key
5. Write blobs to `~/.ollama/models/blobs/`, write Ollama manifest

### API (`api/`)

Go HTTP server with embedded React SPA. SQLite database for model registry.

**Endpoints:**

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| `GET` | `/api/models` | No | Paginated model list with search |
| `GET` | `/api/models/featured` | No | List up to 5 featured models |
| `GET` | `/api/models/:id` | No | Model detail with blob info |
| `PUT` | `/api/models/:id/featured` | Yes | Toggle featured flag (owner only) |
| `POST` | `/api/models` | Yes | Submit model (sync Walrus validation) |
| `GET` | `/api/manifest/preview` | No | Pre-fetch manifest info for submit form |
| `GET` | `/api/blobs/:obj_id` | No | Proxy raw blob content from Walrus |
| `GET/POST` | `/api/auth/sui/nonce` | No | Generate/verify Sui wallet nonce |
| `POST` | `/api/auth/sui/verify` | No | Verify Sui wallet signature |
| `GET` | `/api/auth/login` | No | GitHub OAuth redirect |
| `GET` | `/api/auth/callback` | No | OAuth callback |
| `GET` | `/api/auth/me` | Yes | Current user info |
| `GET` | `/api/users/:id/models` | No | Models by user |
| `GET` | `/api/config` | No | Frontend config (network, featured owner status) |

**Auth modes:** `open` (no auth), `sui` (wallet signature), `token` (bearer token), `github` (OAuth).

### Site (`site/`)

React/TypeScript/Vite SPA with Tailwind CSS. Embedded into the API binary via `go:embed`.

**Pages:**

| Route | Description |
|-------|-------------|
| `/` | Landing page — hero, how-it-works, CTA |
| `/models` | Model listing with search |
| `/models/:id` | Model detail — metadata, pull command, blob details, config/license/params content |
| `/submit` | Submit form with live manifest preview |
| `/profile` | User profile + published models |

### Manifest Schema (`pkg/manifest/`)

Wolllama manifest v2 — stored on Walrus, its object ID is the model's unique identifier.

```json
{
  "wolllamaVersion": 2,
  "provider": "walrus",
  "name": "llama3.2:3b-q4_K_M",
  "ollamaManifest": { "schemaVersion": 2, "config": {...}, "layers": [...] },
  "blobs": {
    "sha256:abc...": {"single": "walrus_objid_AAA"},
    "sha256:def...": {"chunks": ["c1", "c2", "c3"]}
  },
  "createdAt": "2025-06-05T10:30:00Z"
}
```

### Storage Providers (`pkg/storage/`)

```go
type Provider interface {
    Upload(data []byte) (blobID string, err error)
    MaxChunkSize() int64
    Name() string
}
```

| Provider | `MaxChunkSize()` | Upload Target |
|----------|-----------------|---------------|
| Walrus | 256 MB | Walrus publisher (`PUT /v1/blobs`) |
| Tatum | 45 MB | Tatum gateway (`POST /v4/data/storage/upload`) |
| IPFS | 100 MB (stub) | Pinata |
| S3 | 5 GB (stub) | AWS S3 |

### Walrus Client (`pkg/walrus/`)

Wraps `github.com/namihq/walrus-go` with fallback download logic:

- `ReadBlob(id)` — standard `/v1/blobs/{id}`
- `ReadBlobByQuiltPatchID(id)` — `/v1/blobs/by-quilt-patch-id/{id}` (Tatum)
- `ReadBlobWithFallback(id)` — tries standard, falls back to quilt-patch
- `HeadBlob(id)` — metadata without download
- `StoreBlob(data, epochs)` / `StoreBlobFromReader(reader, epochs)` — upload

## Data Flow

### Push (Walrus provider)
```
Ollama disk → CLI reads blobs → pkg/walrus.StoreBlob() → Walrus Publisher
                                                           ↓
                                          wolllama manifest → StoreBlob() → object ID
```

### Push (Tatum provider)
```
Ollama disk → CLI reads blobs → pkg/storage/tatum.Upload() → Tatum API
                            multipart POST /v4/data/storage/upload
                                                           ↓
                                          poll GET /job/{id} → status: CERTIFIED
                                                           ↓
                                          extract quiltPatchId → return to CLI
```

### Pull (any provider)
```
manifest object ID → ReadBlobWithFallback() → quilt-patch or standard endpoint
                                              ↓
wolllama manifest parsed → for each blob ref:
  if chunked: download chunks → reassemble → SHA256 verify
  if single:  download blob → SHA256 verify
                                              ↓
write to ~/.ollama/models/blobs/ + manifest registry
```

### Site submission
```
User pastes obj ID → GET /api/manifest/preview?obj_id=...
                  → API fetches manifest via ReadBlobWithFallback()
                  → returns {name, tag, blob_count, total_size}
                  → User confirms → POST /api/models
                  → API validates manifest → stores in SQLite
```
