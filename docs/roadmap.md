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
- Three auth modes: open, token, GitHub OAuth
- GoReleaser + GitHub Actions CI/CD

## v1.2 — Multi-Provider Storage ✅

Pluggable storage backends. Tatum gateway for managed Walrus access.

- `pkg/storage` provider abstraction with factory
- **WalrusProvider**: direct Walrus, 256 MB chunks, full push+pull
- **TatumProvider**: managed gateway with Sui key handling
  - Multipart upload to `POST /v4/data/storage/upload`
  - Certification polling every 5s
  - QuiltPatchId extraction from certified response
  - 45 MB chunks (Tatum's multipart-safe limit)
- `ReadBlobWithFallback` — auto-detects Tatum quilt-patch IDs
- Quilt-patch download endpoint for Tatum-uploaded blobs
- `--provider walrus|tatum`, `--tatum-api-key` CLI flags
- `walrus_network` config for mainnet/testnet switching
- IPFS + S3 provider stubs with clear "not yet implemented" errors

## v1.3 — IPFS Provider 🔜

Pinata SDK integration for IPFS storage backend.

- `IPFSProvider` implementation in `pkg/storage/ipfs.go`
- Pinata API for upload (`pinFileToIPFS`) and download (`gateway`)
- 100 MB chunks
- `--provider ipfs --pinata-api-key xxx --pinata-secret-key xxx`
- Content-addressed retrieval via IPFS CID

## v1.4 — S3 Provider 🔜

AWS S3 SDK integration for enterprise/private cloud storage.

- `S3Provider` implementation in `pkg/storage/s3.go`
- S3 multipart upload for 5 GB chunks
- Configurable bucket, region, credentials
- `--provider s3 --s3-bucket xxx --s3-region xxx`

## v2.0 — Private Registry 📋

Authenticated model registry with access control.

- `wolllama signin` CLI command (browser-based OAuth flow)
- Full GitHub OAuth with user accounts
- Public/private model visibility
- SealSDK encryption for private model storage
- `wolllama publish` — submit models from CLI
- API key management for programmatic access
- PostgreSQL option for production deployments

## v3.0 — Ecosystem 📋

Advanced features for broader adoption.

- Tatum storage gateway mode (caching + API key auth)
- Pinata IPFS gateway as alternate backend
- Model search with tags and categories
- Model versioning and update tracking
- Usage analytics dashboard
- Webhook notifications for new model submissions
- Multi-user team/organization support
