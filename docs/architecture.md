# Wolllama — Architecture Specification v1

> Decentralized model registry for Ollama. Push and pull models via Walrus storage.
> No central registry. No gatekeepers.

---

## 1. Product Vision

| Version | Mode | Description |
|---------|------|-------------|
| **V1** | Direct (B-mode) | Push/pull straight to Walrus, fully decentralized, no auth. "Freedom from Ollama registry." |
| **V2** | Registry (A-mode) | Private model registry with GitHub OAuth, public/private models, SealSDK encryption. "Docker Hub for models." |

**Target audience:** AI developers comfortable with Ollama CLI who want independence from the Ollama registry.

**Sidecar relationship:** Wolllama requires Ollama installed. It reads/writes `~/.ollama/models/` directly — no daemon interaction. After pull, user restarts Ollama to pick up new models.

---

## 2. System Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│ wolllama CLI │     │ wolllama API │     │ wolllama Site│
│   (Go)       │     │   (Go)       │     │ (React/Vite) │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │ Store/Read         │ Fetch manifest     │ (embedded in API)
       ▼                    ▼                    ▼
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Walrus     │     │   SQLite     │     │  embed.FS    │
│ (Publisher + │     │  (models,    │     │  (SPA served │
│  Aggregator) │     │   users)     │     │  by API)     │
└──────────────┘     └──────────────┘     └──────────────┘
```

**Key principle:** CLI never talks to the API for push/pull. Walrus is the source of truth for model data. The site + API provide discovery, browsing, and (in v2) authenticated registry operations.

---

## 3. Monorepo Structure

```
wolllama/
├── cli/                    # wolllama CLI (Go module)
│   ├── main.go
│   ├── cmd/                # cobra commands: push, pull, show, list, config
│   ├── internal/
│   │   └── ollama/         # Ollama store interaction (filesystem reads/writes)
│   └── go.mod
├── api/                    # wolllama API server (Go module)
│   ├── main.go
│   ├── handler/            # HTTP handlers
│   ├── db/                 # SQLite queries, migrations
│   ├── auth/               # GitHub OAuth
│   └── go.mod
├── site/                   # wolllama frontend (React/TypeScript/Vite)
│   ├── src/
│   │   ├── pages/          # Landing, Models, ModelDetail, Profile, Submit
│   │   ├── components/     # Shared UI components
│   │   └── lib/            # API client, auth helpers
│   ├── package.json
│   └── vite.config.ts
├── pkg/                    # Shared Go libraries (go.work covers cli/ + api/)
│   ├── manifest/           # WolllamaManifest struct, JSON marshal/unmarshal, validate
│   │   ├── manifest.go
│   │   └── manifest_test.go
│   └── walrus/             # Thin wrapper around walrus-go: client init, convenience helpers
│       └── client.go
├── scripts/                # Build/dev/release scripts
├── Taskfile.yml            # Top-level task orchestration
├── go.work                 # Go workspace: cli, api, pkg
└── README.md
```

**Tooling:** Taskfile (https://taskfile.dev) for orchestration. `go.work` for shared Go types between CLI and API. No monorepo framework (Turborepo/Nx) — three packages, not thirty.

---

## 4. Wolllama Manifest (Data Model)

The wolllama manifest is a JSON blob stored on Walrus. It maps Ollama model blobs to their Walrus object IDs.

```json
{
  "wolllamaVersion": 1,
  "name": "llama3.2:3b-q4_K_M",
  "ollamaManifest": {
    "schemaVersion": 2,
    "config": {
      "mediaType": "application/vnd.ollama.image.model",
      "digest": "sha256:abc123...",
      "size": 1024
    },
    "layers": [
      { "mediaType": "application/vnd.ollama.image.license", "digest": "sha256:def456...", "size": 44 },
      { "mediaType": "application/vnd.ollama.image.model", "digest": "sha256:ghi789...", "size": 3456789012 }
    ]
  },
  "blobs": {
    "sha256:abc123...": "walrus_objid_AAA",
    "sha256:def456...": "walrus_objid_BBB",
    "sha256:ghi789...": "walrus_objid_CCC"
  },
  "createdAt": "2025-01-15T10:30:00Z"
}
```

**The manifest's own Walrus object ID** is the unique identifier for the published model. It's what users share and what `wolllama pull` accepts.

---

## 5. CLI Specification

### 5.1 Commands

| Command | Description |
|---------|-------------|
| `wolllama push <model:tag>` | Upload model from local Ollama store → Walrus. Print manifest object ID. |
| `wolllama pull <obj-id>` | Fetch manifest from Walrus → reconstruct in Ollama models directory. |
| `wolllama show <obj-id>` | Fetch manifest from Walrus, print human-friendly model summary. |
| `wolllama list` | List locally synced models from `~/.wolllama/manifests/` cache. |
| `wolllama config` | Get/set Walrus publisher/aggregator URLs, epochs. |

### 5.2 Push Flow

```
wolllama push llama3.2:3b-q4_K_M
  │
  ├─ 1. Parse ~/.ollama/models/manifests/registry.ollama.ai/library/llama3.2/3b-q4_K_M
  │     → Extract config digest + layer digests
  │
  ├─ 2. For each blob in {config} ∪ {layers}:
  │     └─ Read blob from ~/.ollama/models/blobs/sha256-<digest>
  │     └─ POST to Walrus Publisher via StoreFromReader
  │     └─ Collect walrus object ID
  │     └─ [sequential, with per-blob progress bar]
  │
  ├─ 3. Build wolllama manifest JSON (embedding original Ollama manifest)
  │
  ├─ 4. Store wolllama manifest on Walrus → get manifest object ID
  │
  └─ 5. Print success summary:
        ✓ llama3.2:3b-q4_K_M pushed to Walrus

          Model:    llama3.2:3b-q4_K_M
          Blobs:    4 uploaded (0 deduplicated)
          Size:     4.7 GB
          Epochs:   10
          Manifest: O1ABC...xyz

          Share: wolllama pull O1ABC...xyz
          List:   https://wolllama.dev/models  (submit to appear here)
```

### 5.3 Pull Flow

```
wolllama pull O1ABC...xyz
  │
  ├─ 1. GET wolllama manifest from Walrus Aggregator (by object ID)
  │
  ├─ 2. For each blob in manifest.blobs:
  │     └─ GET blob from Walrus Aggregator (by walrus object ID)
  │     └─ Write to ~/.ollama/models/blobs/sha256-<digest>
  │     └─ [sequential, with per-blob progress bar]
  │
  ├─ 3. Write original Ollama manifest to:
  │     ~/.ollama/models/manifests/registry.ollama.ai/library/<name>/<tag>
  │
  ├─ 4. Cache wolllama manifest in ~/.wolllama/manifests/<obj-id>.json (for `wolllama list`)
  │
  └─ 5. Print success + restart notice
```

**Post-pull notice:**
```
✓ llama3.2:3b-q4_K_M pulled from Walrus

  Restart Ollama to use: ollama serve
  Then run: ollama run llama3.2:3b-q4_K_M
```

### 5.4 Show Output

```
$ wolllama show O1ABC...xyz

  Model:    llama3.2:3b-q4_K_M
  Blobs:    4
  Size:     4.7 GB
  Created:  2025-01-15T10:30:00Z
  Epochs:   ~10 remaining

  Blobs:
    sha256:abc123... (config)    1.0 KB  → walrus_objid_AAA
    sha256:def456... (license)     44 B  → walrus_objid_BBB
    sha256:ghi789... (model)     4.7 GB  → walrus_objid_CCC

  View on site: https://wolllama.dev/models/<submitted-id>
```

### 5.5 Configuration

Default Walrus testnet endpoints shipped in the binary. Overridable:

```yaml
# ~/.wolllama/config.yaml
publisher_url: "https://publisher.walrus-testnet.walrus.space"
aggregator_url: "https://aggregator.walrus-testnet.walrus.space"
epochs: 10
```

Env vars override: `WOLLLAMA_PUBLISHER_URL`, `WOLLLAMA_AGGREGATOR_URL`, `WOLLLAMA_EPOCHS`.

Flag for custom Ollama path: `--ollama-path` / `WOLLLAMA_OLLAMA_HOME`.

### 5.6 Error Handling

- Actionable error messages with suggestions (no interactive prompts)
- Push failure: print which blobs succeeded with their Walrus object IDs
- Pull failure: print which blobs downloaded, suggest retry
- Structured logging via `log/slog`; verbose output with `--verbose`
- Format version check: fail clearly if Ollama manifest schema is unsupported

### 5.7 Distribution

- Week 1: `go install github.com/<org>/wolllama/cli@latest` + GitHub Releases (GoReleaser, cross-compile linux/macOS amd64+arm64)
- Week 2-3: Homebrew tap formula

---

## 6. API Specification

### 6.1 Endpoints

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| `GET` | `/api/models` | No | List/search models (paginated) |
| `GET` | `/api/models/:id` | No | Single model detail + publisher README |
| `POST` | `/api/models` | Yes | Submit model manifest (body: `{ manifestObjId, displayName, descriptionMd? }`) |
| `GET` | `/api/auth/login` | No | Redirect to GitHub OAuth |
| `GET` | `/api/auth/callback` | No | GitHub OAuth callback |
| `GET` | `/api/auth/me` | Yes | Current user info |
| `GET` | `/api/users/:id/models` | No | Models published by a specific user |

### 6.2 POST /api/models — Submission Flow

```
1. User authenticated (GitHub OAuth session cookie)
2. Accept { manifestObjId, displayName, descriptionMd }
3. Fetch wolllama manifest from Walrus by manifestObjId (sync validation)
   ├─ Fail → 400: "Manifest not found on Walrus"
   ├─ Invalid schema → 400: "Invalid wolllama manifest"
   └─ Success → continue
4. Extract model name + blob metadata from manifest
5. Insert into SQLite: users.id, displayName, manifestObjId, descriptionMd, size, blobCount, created_at
6. Return 201 → redirect to /models/<id>
```

### 6.3 Database Schema (SQLite)

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    github_id INTEGER UNIQUE NOT NULL,
    username TEXT NOT NULL,
    avatar_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE models (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    submitter_id INTEGER NOT NULL REFERENCES users(id),
    manifest_obj_id TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    description_md TEXT,
    original_name TEXT,       -- from embedded Ollama manifest
    tag TEXT,                 -- from embedded Ollama manifest
    total_size INTEGER,       -- bytes, computed from blob Head calls
    blob_count INTEGER,
    walrus_blob_id TEXT,      -- the manifest's own Walrus object ID (same as manifest_obj_id)
    available BOOLEAN DEFAULT 1, -- health check status
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_models_display_name ON models(display_name);
CREATE INDEX idx_models_submitter ON models(submitter_id);
CREATE INDEX idx_models_available ON models(available);
```

### 6.4 Model Health Check

Daily background goroutine: `Head` each model's `manifest_obj_id` on Walrus. If manifest is gone → set `available = 0`. Site shows "unavailable" badge on expired models.

### 6.5 Authentication

GitHub OAuth via `golang.org/x/oauth2`. Session cookies (encrypted). No password management. Wallet-based auth (Sui) deferred to v2.

---

## 7. Site Specification

### 7.1 Routes

| Route | Auth | Content |
|-------|------|---------|
| `/` | No | Landing page (pitch deck) |
| `/models` | No | Public model listing (search, paginate, browse) |
| `/models/:id` | No | Model detail page (metadata card + publisher README) |
| `/profile` | Yes | User's submitted models |
| `/submit` | Yes | Form: `manifestObjId` (required), `displayName` (prefilled from manifest, editable), `descriptionMd` (optional) |
| `/login` | No | Redirect to `/api/auth/login` |

### 7.2 Landing Page (`/`)

Design concept: *"Your models. Your storage. No limits."*

1. **Hero**: Terminal mockup showing `wolllama push llama3.2` + subtle animation. Tagline + subtext.
2. **How It Works**: 3-column layout — Push (CLI → Walrus) → Share (object ID) → Pull (Walrus → Ollama)
3. **Featured Models**: Grid of 6 cards pulled from `GET /api/models?limit=6`
4. **CTA**: Install snippet + GitHub link + Twitter link

### 7.3 Model List Page (`/models`)

- Search bar (by display name)
- Grid of model cards: name, publisher avatar, size, blob count, availability badge
- Pagination (cursor or offset)

### 7.4 Model Detail Page (`/models/:id`)

- **Metadata card**: display name, publisher (GitHub avatar + username), size, blob count, date added, availability badge
- **Pull command**: `wolllama pull <manifest-obj-id>` with copy button
- **Description**: Publisher's rendered markdown
- **Blob list**: collapsible table of sha256 → walrus object ID mappings

### 7.5 Submission Form (`/submit`)

```
┌─────────────────────────────────────────────┐
│  Submit a Model                             │
│                                             │
│  Manifest Object ID *                       │
│  ┌─────────────────────────────────────────┐│
│  │ O1ABC...xyz                             ││
│  └─────────────────────────────────────────┘│
│  (Paste the ID from 'wolllama push' output) │
│                                             │
│  Display Name *               [llama3.2   ]│
│  (Prefilled from manifest, editable)        │
│                                             │
│  Description (supports Markdown)            │
│  ┌─────────────────────────────────────────┐│
│  │                                         ││
│  └─────────────────────────────────────────┘│
│                                             │
│  [Submit Model]                             │
└─────────────────────────────────────────────┘
```

### 7.6 Tech Stack

- React 18+ with TypeScript
- Vite (build tool)
- React Router v6 (client-side routing)
- Plain CSS modules or Tailwind (TBD)
- Vitest + React Testing Library (component tests)

---

## 8. Deployment

### 8.1 Single Binary Pattern

The Go API binary embeds the built React SPA via `embed.FS`:

```go
//go:embed site/dist/*
var siteFS embed.FS

// In main.go:
mux.Handle("/", http.FileServer(http.FS(siteFS)))
mux.Handle("/api/", apiHandler)
```

**Build pipeline:**
```bash
task build
# → cd site && npm run build
# → cd api && go build -o wolllama-api .
# Single binary: wolllama-api
```

Deploy target: fly.io, Railway, or a $6 VPS. One binary, one process, no CORS.

### 8.2 CI/CD (GitHub Actions)

**On PR:**
- Lint Go (`golangci-lint`)
- Test Go (`go test ./...`)
- Lint TypeScript (`eslint`)
- Test TypeScript (`vitest`)
- Build site (`npm run build`)
- Build API (`go build`)

**On tag push (`v*`):**
- GoReleaser: cross-compile CLI binaries + checksums → GitHub Release
- Build API binary with embedded site
- (Optional) Deploy API binary

---

## 9. Testing Strategy

| Layer | Type | Tool | Scope |
|-------|------|------|-------|
| CLI + API | Unit | `go test` | Mocks for Walrus/Ollama interfaces |
| CLI + API | Integration | `go test -tags=integration` | Real Walrus testnet (manual, not in CI) |
| Site | Component | Vitest + RTL | Component rendering, form validation |
| Site | E2E | Playwright (v2) | Full flows |

---

## 10. Dependencies

| Component | Library | Purpose |
|-----------|---------|---------|
| CLI + API | `github.com/namihq/walrus-go` | Walrus publish/aggregate HTTP client |
| CLI + API | `github.com/spf13/cobra` | CLI command framework |
| CLI + API | `github.com/spf13/viper` | Config file + env var management |
| CLI + API | `golang.org/x/exp/slog` or `log/slog` | Structured logging |
| API | `github.com/mattn/go-sqlite3` or `modernc.org/sqlite` | SQLite driver |
| API | `golang.org/x/oauth2` | GitHub OAuth |
| API | `github.com/gorilla/sessions` or similar | Session management |
| Site | `react-router-dom` | Client-side routing |
| Site | `react-markdown` | Render publisher markdown |

---

## 11. What's Deferred to V2+

| Feature | Target |
|---------|--------|
| Private registry mode (API auth on push/pull) | V2 |
| SealSDK encryption for private model storage | V2 |
| `wolllama publish` CLI command (submit from CLI) | V2 |
| `wolllama signin` CLI command | V2 |
| Tatum storage gateway mode | V2 |
| Wallet-based auth (Sui) | V2 |
| Model card extraction from Ollama config (params, license) | V2 |
| Concurrent blob upload | V1.1 |
| `AlreadyCertified` dedup detection on retry | V1.1 |
| Homebrew tap | Week 2-3 |
| Playwright E2E tests | V2 |
| PostgreSQL (when SQLite limits are hit) | V2 |
