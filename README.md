# Wolllama — Decentralized Model Registry for Ollama

Your models. Your storage. No limits.

Wolllama is a decentralized model registry that lets you push and pull Ollama models
to [Walrus](https://walrus.site) decentralized storage. No central registry. No gatekeepers.

```bash
# Push a model from Ollama to Walrus
wolllama push llama3.2:3b-q4_K_M

# Pull it anywhere
wolllama pull O1ABCdef...xyz
```

## Features

- **Decentralized storage** — Models live on Walrus, not a central server
- **Sidecar to Ollama** — Reads/writes `~/.ollama/models/` directly, no daemon needed
- **Multi-provider** — Walrus native, Tatum gateway, IPFS + S3 (planned)
- **Chunked large files** — Auto-split blobs > 500 MB for Walrus compatibility
- **SHA256 verification** — Every pulled blob is checksum-verified
- **Web registry** — Browse, discover, and submit models at the Wolllama site
- **Open source** — MIT licensed

## Quick Start

### Install

```bash
# Download binary
curl -fsSL https://github.com/wolllama/wolllama/releases/latest/download/wolllama-darwin-arm64.tar.gz | tar xz

# Or with Go
go install github.com/wolllama/cli@latest
```

### Push a model

```bash
# List available models (from Ollama)
wolllama list

# Push to Walrus
wolllama push llama3.2:latest

# Push via Tatum gateway (no Sui key needed)
wolllama push llama3.2:latest --provider tatum --tatum-api-key YOUR_KEY
```

### Pull a model

```bash
wolllama pull <manifest-obj-id>

# Pull to custom Ollama path
wolllama pull <manifest-obj-id> --ollama-path /tmp/fake-ollama

# Restart Ollama to use the model
ollama serve
ollama run llama3.2:latest
```

### Configuration

```bash
# Switch to Walrus mainnet (for production / Tatum)
wolllama config set walrus_network mainnet

# Set Tatum as default provider
wolllama config set provider tatum
wolllama config set tatum-api-key YOUR_KEY

# View all settings
wolllama config show
```

### Start the API + Site

```bash
# No auth mode
WOLLLAMA_AUTH_MODE=open ./wolllama-api

# With Walrus mainnet (for Tatum models)
WOLLLAMA_AUTH_MODE=open WOLLLAMA_WALRUS_NETWORK=mainnet ./wolllama-api
```

Open `http://localhost:8080` — browse models, submit new ones, view blob details.

## Storage Providers

| Provider | Push | Pull | Chunk Size | Best For |
|----------|------|------|------------|----------|
| **Walrus** | Direct publisher | Aggregator | 256 MB | Full control, no third-party |
| **Tatum** | Managed gateway | Quilt-patch | 45 MB | No Sui key management |
| **IPFS** | Pinata (planned) | Pinata | 100 MB | IPFS ecosystem |
| **S3** | AWS SDK (planned) | AWS SDK | 5 GB | Enterprise/private cloud |

## Project Structure

```
wolllama/
├── cli/              # Go CLI (cobra commands)
├── api/              # Go API server (SQLite + embedded SPA)
├── site/             # React/Vite/Tailwind SPA
├── pkg/
│   ├── manifest/     # Wolllama manifest schema
│   ├── walrus/       # Walrus HTTP client wrapper
│   └── storage/      # Multi-provider abstraction
├── docs/             # Architecture + development docs
├── Taskfile.yml      # Build/dev orchestration
└── go.work           # Go workspace
```

## Links

- [Architecture](docs/architecture.md)
- [Development Guide](docs/development.md)
- [Roadmap](docs/roadmap.md)
