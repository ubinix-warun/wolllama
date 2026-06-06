# Wolllama Roadmap

## v1.0 — Direct Walrus ✅

First release: CLI sidecar, API server, web registry.

- `wolllama` CLI: `push`, `pull`, `show`, `list`, `config`
- Sidecar to Ollama — reads/writes `~/.ollama/models/` directly
- Wolllama manifest v2 with chunked blob support (`BlobRef{single|chunks}`)
- 256 MB chunks for blobs > 500 MB (Walrus aggregator limit)
- SHA256 checksum verification on every pulled blob
- `wolllama-api` Go server with embedded React SPA
- `wolllama-site` React/Vite/Tailwind: landing, model browser, submit, profile
- SQLite database for model registry
- Auth modes: open, token
- GoReleaser + GitHub Actions CI/CD

## v1.2 — Multi-Provider Storage ✅

Pluggable storage backends. Tatum gateway for managed Walrus access.

- `pkg/storage` provider abstraction with factory
- **WalrusProvider**: direct Walrus, 256 MB chunks, full push+pull
- **TatumProvider**: managed gateway with Sui key handling
  - Multipart upload + certification polling + quiltPatchId extraction
  - 45 MB chunks (Tatum's multipart-safe limit)
- `ReadBlobWithFallback` — auto-detects Tatum quilt-patch IDs
- `--provider walrus|tatum`, `--tatum-api-key` CLI flags
- `walrus_network` config for mainnet/testnet switching
- IPFS + S3 provider stubs

## v1.3 — Sui On-Chain Registry 🔜

Decentralized model registry on Sui blockchain. Users interact via Sui wallet.
SQLite remains the primary data store — the Sui contract provides on-chain
verification, audit trail, and decentralized identity.

- **Sui wallet login** — sign in with Sui wallet (zkLogin or standard)
  - Sign a challenge message to prove wallet ownership
  - Wallet address becomes the user's identity
- **On-chain model registry** — Move smart contract on Sui
  - `ModelRegistry`: `register_model(name, manifest_obj_id)`, `update_model(...)`
  - Model metadata stored on-chain: name, manifest object ID, publisher address, timestamp
  - Immutable audit trail — every submission/update is a Sui transaction
- **Frontend Sui integration** — `@mysten/dapp-kit` for wallet connection
  - "Connect Sui Wallet" button in header
  - Submit model triggers Sui transaction (sign + execute via wallet)
  - Profile page shows on-chain activity + registered models
- **API sync** — API indexes on-chain events into SQLite for fast queries
  - Sui event listener watches `ModelRegistry` events
  - Keeps SQLite in sync with on-chain state
  - Public endpoints read from SQLite (fast), mutations go through Sui (trustless)

## v2.0 — Private Registry + Encryption 📋

Access control and encryption for private model sharing.

- `wolllama signin` CLI command (browser-based OAuth flow)
- Public/private model visibility
- SealSDK encryption for private model storage
- `wolllama publish` — submit models from CLI

## v3.0 — Ecosystem 📋

- Model search with tags and categories
- Model versioning and update tracking
- Usage analytics dashboard
- Multi-user team/organization support

---

## Unplanned Backlog

Features acknowledged but not scheduled for any specific version:

- **GitHub OAuth login** — deferred in favor of Sui wallet auth (v1.3)
- **IPFS provider** — Pinata SDK integration
- **S3 provider** — AWS S3 SDK integration
- **Tatum storage gateway mode** — caching + API key auth layer
