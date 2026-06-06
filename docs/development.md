# Wolllama Development Guide

## Prerequisites

- **Go 1.23+** — managed via [gvm](https://github.com/moovweb/gvm)
- **Node.js 22+** — managed via [nvm](https://github.com/nvm-sh/nvm)
- **Task** — `go install github.com/go-task/task/v3/cmd/task@latest` or `brew install go-task`

## Getting Started

```bash
git clone https://github.com/wolllama/wolllama.git
cd wolllama

# Install dependencies
task install

# Build everything (site + API + CLI)
task build

# Verify
./cli/wolllama --help
./api/wolllama-api --help
```

## Project Layout

```
wolllama/
├── cli/                        # CLI binary (Go)
│   ├── main.go                 # Entry point
│   ├── cmd/                    # Cobra commands
│   │   ├── root.go             # Root command + global flags
│   │   ├── config.go           # config show/set + walrusURLs()
│   │   ├── push.go             # Push flow (storage provider)
│   │   ├── pull.go             # Pull flow + SHA256 verification
│   │   ├── show.go             # Manifest display
│   │   └── list.go             # Model listing
│   └── internal/ollama/        # Ollama disk store R/W
│       └── store.go            # Manifest + blob filesystem ops
│
├── api/                        # API server (Go)
│   ├── main.go                 # Entry point + routing + SPA embed
│   ├── handler/handler.go      # HTTP handlers
│   ├── db/db.go                # SQLite: users, models, CRUD
│   ├── auth/github.go          # GitHub OAuth client
│   └── site/dist/              # Built SPA (copied during build)
│
├── site/                       # React SPA
│   ├── src/
│   │   ├── App.tsx             # Router
│   │   ├── lib/api.ts          # API client + types
│   │   ├── components/         # Layout, ModelCard
│   │   └── pages/              # Landing, Models, ModelDetail, Profile, Submit
│   └── vite.config.ts          # Vite + Tailwind + API proxy
│
├── pkg/                        # Shared Go libraries
│   ├── manifest/               # Wolllama manifest schema + tests
│   │   ├── manifest.go
│   │   └── manifest_test.go
│   ├── walrus/                 # Walrus HTTP client wrapper
│   │   └── client.go
│   └── storage/                # Multi-provider abstraction
│       ├── provider.go         # Interface + factory
│       ├── walrus.go           # WalrusProvider
│       ├── tatum.go            # TatumProvider
│       ├── ipfs.go             # IPFS stub
│       └── s3.go               # S3 stub
│
├── docs/                       # Documentation
│   ├── architecture.md
│   ├── development.md          # This file
│   └── roadmap.md
│
├── go.work                     # Go workspace (cli, api, pkg)
├── Taskfile.yml                # Build/dev tasks
├── .goreleaser.yml             # CLI release config
└── .github/workflows/ci.yml    # CI pipeline
```

## Common Tasks

```bash
# Build
task build              # Full build: site + API + CLI
task build-cli          # CLI only
task build-api          # API with embedded site
task build-site         # Site only

# Development (hot reload)
task dev                # Start API + Vite dev server
task dev-api            # API only (port 8080)
task dev-site           # Vite dev server (port 5173, proxies /api)

# Testing
task test               # All tests
task test-go            # Go tests
task test-site          # Frontend tests

# Linting
task lint               # All linters
task lint-go            # golangci-lint
task lint-site          # ESLint

# Code quality
task fmt                # Format all code
task tidy               # Tidy Go modules

# Clean
task clean              # Remove build artifacts
```

## Development Workflow

### CLI Development

```bash
# Build and test a command
task build-cli
./cli/wolllama list
./cli/wolllama push llama3.2:latest
./cli/wolllama config show

# Test with custom Ollama path
mkdir -p /tmp/fake-ollama/models/{blobs,manifests}
./cli/wolllama pull <obj-id> --ollama-path /tmp/fake-ollama
```

### API + Site Development

```bash
# Terminal 1: API
task dev-api

# Terminal 2: Site (hot reload)
task dev-site

# Open http://localhost:5173
# API calls proxy to localhost:8080 automatically
```

For full production build (API with embedded site):

```bash
task build-api
WOLLLAMA_AUTH_MODE=open ./api/wolllama-api
# Open http://localhost:8080
```

### Adding a New Storage Provider

1. Create `pkg/storage/<name>.go` implementing the `Provider` interface
2. Add the provider case to `provider.go`'s `New()` factory
3. Add config keys to `cli/cmd/config.go` and `cli/cmd/root.go`
4. Add env var support if needed
5. Update `README.md` provider table

```go
// pkg/storage/example.go
type ExampleProvider struct { ... }

func NewExampleProvider(cfg Config) (*ExampleProvider, error) { ... }
func (e *ExampleProvider) Name() string       { return "example" }
func (e *ExampleProvider) MaxChunkSize() int64 { return 100 * 1024 * 1024 }
func (e *ExampleProvider) Upload(data []byte) (string, error) { ... }
```

### Running with Tatum

```bash
# CLI
wolllama push <model> --provider tatum --tatum-api-key YOUR_KEY

# API (serves Tatum models)
WOLLLAMA_AUTH_MODE=open WOLLLAMA_WALRUS_NETWORK=mainnet ./api/wolllama-api
```

### Running API with Auth

```bash
# Open mode (no auth)
WOLLLAMA_AUTH_MODE=open ./api/wolllama-api

# Token mode
WOLLLAMA_AUTH_MODE=token WOLLLAMA_API_TOKEN=my-secret ./api/wolllama-api
# Submit with: curl -H "Authorization: Bearer my-secret" -X POST ...

# GitHub OAuth
GITHUB_CLIENT_ID=xxx GITHUB_CLIENT_SECRET=xxx ./api/wolllama-api
```

## Configuration Reference

### CLI (`~/.wolllama/config.yaml`)

| Key | Default | Description |
|-----|---------|-------------|
| `provider` | `walrus` | Storage backend: walrus, tatum, ipfs, s3 |
| `walrus_network` | `testnet` | Walrus network: testnet, mainnet |
| `publisher_url` | auto | Override Walrus publisher URL |
| `aggregator_url` | auto | Override Walrus aggregator URL |
| `epochs` | `10` | Storage epochs |
| `tatum_api_key` | — | Tatum API key |
| `tatum_api_url` | `https://api.tatum.io` | Tatum API base URL |

### API (Environment Variables)

| Variable | Default | Description |
|----------|---------|-------------|
| `WOLLLAMA_AUTH_MODE` | `github` | Auth mode: open, token, github |
| `WOLLLAMA_API_TOKEN` | — | Bearer token for token mode |
| `WOLLLAMA_WALRUS_NETWORK` | `testnet` | Walrus network |
| `WOLLLAMA_AGGREGATOR_URL` | auto | Override aggregator URL |
| `GITHUB_CLIENT_ID` | — | GitHub OAuth client ID |
| `GITHUB_CLIENT_SECRET` | — | GitHub OAuth client secret |
| `PORT` | `8080` | Listen port |
