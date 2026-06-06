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
- `ReadBlobWithFallback` — three-tier fallback (regular → quilt-patch → quilt-id)
- `--provider walrus|tatum`, `--tatum-api-key` CLI flags
- `walrus_network` config for mainnet/testnet switching
- IPFS + S3 provider stubs

## v1.3 — Sui Off-Chain Registry ✅ (Phase 1)

Sui wallet login via pure ed25519 signature verification. No external auth provider.
SQLite stores wallet address + signature alongside model submissions.

- **Sui wallet auth** — `api/auth/sui.go`: nonce generation, ed25519 verification
  - `GET /api/auth/sui/nonce` + `POST /api/auth/sui/verify`
  - `CreateUserByWallet` — lookup or create user by Sui address
- **Frontend Sui integration** — `@mysten/dapp-kit` + `@mysten/sui`
  - `SuiClientProvider` + `WalletProvider` + `ConnectButton`
  - Auto-detects network from `GET /api/config`
  - Header shows wallet button only in `sui` mode
- **Signed submissions** — `useSignPersonalMessage` on submit
  - `submitter_address` + `signature` stored in SQLite
  - Model detail shows "✓ Sui wallet signed" badge
- **New auth mode** — `WOLLLAMA_AUTH_MODE=sui`
  - Requires wallet address + signature for submissions
  - `open` mode works without wallet (anonymous submissions)
- **API improvements**
  - `WOLLLAMA_SUI_NETWORK` — Sui mainnet/testnet selection
  - `GET /api/config` — exposes walrus_network + sui_network to frontend
  - `GetOrCreateAnonUser` — robust anonymous user creation

## v1.3 Phase 2 — Sui On-Chain Registry 🔜

Move smart contract for full on-chain model registry.

- Move smart contract: `ModelRegistry` with `register_model`, `update_model`
- On-chain audit trail — every submission is a Sui transaction
- Sui event listener → SQLite sync for fast queries
- Replace submit flow: form → wallet sign → Sui transaction → event → API indexes

## v2.0 — Private Registry + Encryption 📋

- `wolllama signin` CLI command
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

- **GitHub OAuth login** — deferred in favor of Sui wallet auth
- **IPFS provider** — Pinata SDK integration
- **S3 provider** — AWS S3 SDK integration
- **Tatum storage gateway mode** — caching + API key auth layer
