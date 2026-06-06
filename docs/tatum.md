# Tatum Integration — Managed Sui & Walrus Access

Tatum is a blockchain infrastructure platform that provides managed access to Sui RPC nodes
and Walrus decentralized storage. Using Tatum with Wolllama eliminates the need to manage
your own Sui keypair, fund gas, or run infrastructure.

## Why Tatum?

### 1. No Sui Key Management

Direct Walrus access requires a Sui wallet with funded SUI tokens for every storage operation.
Tatum handles key management, gas funding, and transaction signing behind the scenes.

| | Direct Walrus | Tatum Gateway |
|---|---|---|
| Sui wallet required | ✅ Yes | ❌ No |
| Fund SUI for gas | ✅ Yes | ❌ No |
| Sign transactions | ✅ Manual | ❌ Automatic |
| Setup time | ~30 min | ~2 min |

### 2. Simplified Billing

Tatum consolidates all Walrus storage costs into a single API plan billing. Instead of managing
SUI token balances and gas estimates, you pay through Tatum's usage-based pricing.

### 3. High-Availability RPC

Direct Sui fullnode access can be rate-limited or unavailable during network congestion.
Tatum provides load-balanced, redundant RPC endpoints with automatic failover.

### 4. Storage Gateway (Tatum → Walrus)

Tatum acts as a managed gateway to Walrus storage:
- **Upload**: `POST /v4/data/storage/upload` — Tatum stages files and certifies them on Walrus
- **Automatic renewal**: Tatum renews storage epochs automatically before they expire
- **Download**: Access blobs via Walrus aggregator using Tatum's quilt-patch IDs

## How Wolllama Uses Tatum

### Push (Storage)

```bash
# With Tatum API key — no Sui wallet needed
wolllama push llama3.2:latest \
  --provider tatum \
  --tatum-api-key YOUR_KEY

# Persistent config
wolllama config set provider tatum
wolllama config set tatum-api-key YOUR_KEY
wolllama push llama3.2:latest
```

Flow:
```
CLI → multipart POST to Tatum /v4/data/storage/upload
    → Tatum stages on Walrus
    → CLI polls GET /v4/data/storage/upload/{jobId}
    → status: CERTIFIED → returns quiltPatchId
    → quiltPatchId stored in Wolllama manifest
```

### Pull (Storage)

```bash
# Pull automatically handles Tatum blobs
wolllama pull <manifest-obj-id>

# Manifest IDs from Tatum push have quiltPatchId suffixes
# Wolllama detects and resolves them transparently
```

Flow:
```
CLI → ReadBlobWithFallback(blobId)
    → 1. Try /v1/blobs/{id} (regular Walrus)
    → 2. If \x01 wrapper detected → try /v1/blobs/by-quilt-patch-id/{id}
    → 3. If that fails → try /v1/blobs/by-quilt-id/{id}/blob
    → Returns raw content ✓
```

### Wallet Connection (Sui RPC)

```bash
# Use Tatum as Sui RPC for wallet connections
WOLLLAMA_AUTH_MODE=sui \
  WOLLLAMA_SUI_RPC_URL="https://api.tatum.io/v3/sui/node/YOUR_KEY" \
  ./api/wolllama-api
```

The frontend's `@mysten/dapp-kit` connects to Sui through Tatum's RPC endpoint
instead of public fullnodes. This provides:
- No rate limiting
- Higher reliability
- Unified billing with storage

### Submit + Browse (API)

```bash
# API handles Tatum manifests transparently
WOLLLAMA_AUTH_MODE=open \
  WOLLLAMA_WALRUS_NETWORK=mainnet \
  ./api/wolllama-api
```

The API's `ReadBlobWithFallback` automatically handles Tatum quilt-patch IDs
when fetching manifest previews, validating submissions, and serving blob content.

## Configuration Reference

| Setting | CLI | API | Description |
|---------|-----|-----|-------------|
| Provider | `--provider tatum` | — | Use Tatum for storage push |
| API Key | `--tatum-api-key` / `config set tatum_api_key` | — | Tatum API key |
| API URL | `config set tatum_api_url` | — | Override Tatum base URL |
| Sui RPC URL | — | `WOLLLAMA_SUI_RPC_URL` | Tatum Sui RPC endpoint |
| Network | `config set walrus_network mainnet` | `WOLLLAMA_WALRUS_NETWORK=mainnet` | Walrus mainnet (required for Tatum) |

## Getting a Tatum API Key

1. Sign up at [dashboard.tatum.io](https://dashboard.tatum.io)
2. Create a new API key from the dashboard
3. Copy the key (format: `t-...`)
4. Set it in wolllama: `wolllama config set tatum-api-key t-...`

## Architecture Diagram

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│ wolllama CLI │     │ wolllama API │     │ wolllama Site│
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │                    │                    │
       │ Tatum Storage      │ ReadBlobWith       │ @mysten/dapp-kit
       │ API (push)         │ Fallback (pull)    │ (wallet)
       ▼                    ▼                    ▼
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│    Tatum     │     │   Walrus     │     │  Tatum Sui   │
│   Gateway    │────▶│  Aggregator  │     │     RPC      │
│              │     │              │     │              │
│ POST /upload │     │ GET /blobs   │     │ Fullnode API │
└──────────────┘     └──────────────┘     └──────────────┘
       │                    ▲
       │ certifies          │
       └────────────────────┘
         Walrus Publisher
```

## Limitations

- **50 MiB upload limit** — Tatum's multipart-safe max. Wolllama auto-chunks at 45 MiB for Tatum.
- **Push-only storage** — Tatum's storage API is for upload. Downloads go through Walrus aggregator.
- **Mainnet only** — Tatum storage requires Walrus mainnet.
