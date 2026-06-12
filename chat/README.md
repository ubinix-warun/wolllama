# Wolllama Chat

Browser-based LLM chat powered by [web-llm](https://github.com/mlc-ai/web-llm). Models run entirely in your browser via WebGPU — no server, no data leaves your device.

## Getting Started

```bash
npm install
npm run dev
```

Open the URL printed by Vite (default `http://localhost:5173`) in Chrome 113+ or Edge 113+. Firefox and Safari require WebGPU flags.

## Models

| Model | Size | Load Time (est.) |
|-------|------|------------------|
| Qwen2 0.5B | ~400 MB | ~1 min |
| Llama 3.2 1B | ~800 MB | ~2-3 min |

Models are downloaded from HuggingFace CDN and cached in the browser's Cache API.

## Requirements

- **Chrome 113+** or **Edge 113+** with WebGPU enabled
- The dev server sends COOP/COEP headers required for `SharedArrayBuffer`

## Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Start Vite dev server |
| `npm run build` | Type-check and production build |
| `npm run lint` | Run ESLint |
| `npm test` | Run Vitest smoke test |
| `npm run preview` | Preview production build locally |
