# Wolllama Chat — Roadmap

> Independent roadmap. Converges with [Wolllama](../docs/roadmap.md) at v1.0 when custom TVM/web-llm forks and MLC-format push support land.

## v0.1 — Browser Chat PoC ✅

First release: in-browser LLM chat via web-llm + WebGPU.

- React/Vite/Tailwind/TypeScript SPA (`chat/`)
- Web Worker inference via `@mlc-ai/web-llm` (`CreateWebWorkerMLCEngine`)
- 2 prebuilt MLC models: Qwen2-0.5B (~400 MB), Llama-3.2-1B (~800 MB)
- Auto-load smallest model on first visit
- Streaming responses with `requestAnimationFrame` batching
- Model selector (friendly name + size badge), model switching (wipe + reload)
- Stop generation button (toggles send → stop)
- Markdown rendering in assistant messages (`react-markdown` + `remark-gfm`)
- Browser Cache API backend for model weights
- WebGPU detection gate (full-page block + Chrome download link for unsupported browsers)
- Error boundary at App root (catch + reload)
- Dark mode via `prefers-color-scheme`
- Cross-origin isolation headers (COOP/COEP) in Vite dev server
- Wally-branded system prompt + favicon
- `Taskfile.yml` integration: `dev-chat`, `build-chat`, `lint-chat`, `test-chat`
- Smoke test (Vitest + jsdom)

## v0.2 — Polish & UX ✅

Gaps from the PoC that improve first impressions.

- **Token usage display** — show tokens/second and total tokens after each response
- **Syntax highlighting in code blocks** — `rehype-highlight` or `prism-react-renderer`
- **Dark mode manual toggle** — sun/moon icon in header, stored in `localStorage`, overrides OS preference
- **User-editable system prompt** — settings panel to customize Wally's persona
- **Generation parameter sliders** — temperature, max_tokens exposed in a collapsible settings panel
- **More models** — expand model catalogue beyond 2 PoC models (Phi, Gemma, Mistral)
- **Cache backend selector** — dropdown: Cache API / IndexedDB / OPFS
- **Granular error boundaries** — separate boundaries around ChatWindow and ModelSelector
- **Mobile testing + polish** — verify on Android Chrome with WebGPU, safe-area fixes

## v0.3 — Multi-Conversation 📋

Conversation management and auth, in incremental sub-versions.

### v0.3.1 — Sidebar UI 🔜

In-memory multi-conversation — no persistence, no auth.

- **Multi-conversation sidebar** — ChatGPT-style sidebar (260px, collapsible on mobile)
  - "New Chat" button — creates empty conversation, auto-loads default model
  - Conversation list — titles auto-generated from first user message
  - Click to switch conversations (messages + model state swap)
  - Delete conversation (trash icon + confirmation)
  - Active conversation highlighted
- **In-memory only** — conversations live in React state, lost on refresh
  - Each conversation tracks: id, title, modelId, messages[], createdAt
  - `useConversations()` hook — `createConversation`, `deleteConversation`, `switchConversation`, `updateTitle`
- **Context window management** — token counter + auto-truncate
  - Token counter in ChatInput footer (e.g. "1,204 / 4,096")
  - Auto-truncate oldest messages when approaching limit (sliding window)
  - Per-model context limits displayed in ModelSelector
  - System prompt always preserved during truncation

### v0.3.2 — Auth Modes 🔜

Sui wallet auth, mirroring `site/` auth patterns.

- **Auth modes** — `open` (default, no auth) and `sui` (wallet signature)
  - `WOLLLAMA_AUTH_MODE` env/config toggle
  - In `open` mode: anonymous usage, conversations in-memory only
  - In `sui` mode: wallet connect button in header/sidebar, signed requests
- **Sui wallet integration** — reuse `@mysten/dapp-kit` + `@mysten/sui` patterns from `site/`
  - `SuiClientProvider` + `WalletProvider` + `ConnectButton`
  - Nonce challenge → sign → verify flow (same as `api/auth/sui.go`)
  - Wallet address stored in React context, displayed in sidebar
- **Auth-gated features** — conversation persistence, model submissions (later)
  - Auth state determines what sidebar actions are available
  - "Connect wallet to save conversations" prompt in sidebar when not authed

### v0.3.3 — Context Window Polish 📋

- Per-model context display in ModelSelector
- Warning when approaching limit (yellow → red)
- Manual "summarize and continue" action for long conversations

## v0.4 — Persistence (memwal) 📋

Long-term memory on Walrus, requires Sui wallet auth from v0.3.2.

- **memwal client** — save/restore conversation history to Walrus
  - Auto-save on each message (debounced 500ms)
  - Restore on page load (fetch conversation manifest)
  - Local IndexedDB pointer cache (conversation ID → Walrus blob ID)
- **Encrypted storage** — SealSDK or similar for private conversations
- **Cross-device sync** — same Sui address = same conversations across devices

## v0.5 — PWA & Offline 📋

Installable, offline-capable chat.

- **PWA manifest** — `vite-plugin-pwa`, install prompt, app icon
- **Service Worker caching** — app shell (HTML/JS/CSS) cached for offline launch
- **Offline model serving** — Service Worker intercepts HuggingFace CDN requests, serves from Cache API
- **Cross-origin cache backend** — Chrome extension storage for larger models

## v1.0 — Custom TVM + Walrus Integration 📋

Swap upstream web-llm/TVM for forked versions with Walrus storage.

- **Custom TVM fork** (`ubinix-warun/tvm`) — compile models with Walrus blob storage support
  - TVM compilation pipeline → MLC-format model artifacts
  - Upload compiled models to Walrus (via wolllama publish pipeline)
- **Custom web-llm fork** (`ubinix-warun/web-llm`) — fetch model weights from Walrus instead of HuggingFace
  - `model_list` URLs point at Walrus blob IDs
  - Walrus client in the browser fetch layer
  - IPFS provider as alternative backend (later phase)
- **wolllama MLC-format push** — CLI support for pushing MLC/TVM-format models
  - `wolllama push --format mlc` — upload MLC model artifacts to Walrus
  - wolllama manifest v3 with MLC model metadata
  - Model registry integration: chat models appear alongside Ollama models in `site/`

## v1.x — Wolllama Registry Convergence 📋

Chat becomes a first-class consumer of the Wolllama model registry.

- **Model discovery from registry** — browse Wolllama-published models from chat
- **One-click load** — click a model in `site/` → "Open in Chat" → auto-loads in `chat/`
- **Shared auth** — Sui wallet login shared between `site/` and `chat/`
- **Unified monorepo** — `chat/` merges into `site/` as a route or monorepo package

## Unplanned Backlog

- **Service Worker MLCEngine** — model survives page refresh (alternative to Web Worker path)
- **Multi-modal models** — vision, audio input when web-llm supports them
- **Function calling / tool use** — when web-llm's function-calling support matures
- **Model fine-tuning UI** — in-browser LoRA/QLoRA via WebGPU
- **Collaborative chat** — shared conversations via Walrus (multi-user, real-time)
- **iOS Safari support** — blocked on Apple shipping WebGPU in Safari stable
