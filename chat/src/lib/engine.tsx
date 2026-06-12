import {
  type MLCEngineInterface,
  type ChatCompletionMessageParam,
  type ChatCompletionChunk,
  type InitProgressReport,
  CreateWebWorkerMLCEngine,
  prebuiltAppConfig,
} from "@mlc-ai/web-llm";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";

// ─── Model catalogue ───────────────────────────────────────────────

/** Friendly metadata for the two PoC models. */
const MODEL_META: Record<string, { name: string; size: string }> = {
  "SmolLM2-360M-Instruct-q4f16_1-MLC": {
    name: "SmolLM2 360M",
    size: "~370 MB",
  },
  "Qwen2-0.5B-Instruct-q4f16_1-MLC": {
    name: "Qwen2 0.5B",
    size: "~400 MB",
  },
  "Llama-3.2-1B-Instruct-q4f16_1-MLC": {
    name: "Llama 3.2 1B",
    size: "~800 MB",
  },
  "Phi-3.5-mini-instruct-q4f16_1-MLC": {
    name: "Phi-3.5 Mini",
    size: "~2.1 GB",
  },
  "Gemma-2B-it-q4f16_1-MLC": {
    name: "Gemma 2B",
    size: "~1.4 GB",
  },
  "Mistral-7B-Instruct-v0.3-q4f16_1-MLC": {
    name: "Mistral 7B",
    size: "~4.3 GB",
  },
};

const MODEL_IDS = Object.keys(MODEL_META);

function buildAppConfig(cacheBackend?: "cache" | "indexeddb" | "opfs") {
  return {
    ...prebuiltAppConfig,
    cacheBackend: cacheBackend ?? prebuiltAppConfig.cacheBackend,
    model_list: prebuiltAppConfig.model_list.filter((m) =>
      MODEL_IDS.includes(m.model_id),
    ),
  };
}

/** Smallest model — auto-loaded on first visit. */
const DEFAULT_MODEL = "SmolLM2-360M-Instruct-q4f16_1-MLC";

// ─── State types ───────────────────────────────────────────────────

export type LoadProgress = InitProgressReport;

interface EngineIdle {
  status: "idle";
  modelId: null;
}

interface EngineLoading {
  status: "loading";
  modelId: string;
  progress: LoadProgress;
}

interface EngineReady {
  status: "ready";
  modelId: string;
  engine: MLCEngineInterface;
}

interface EngineError {
  status: "error";
  modelId: string | null;
  error: string;
}

export type EngineState =
  | EngineIdle
  | EngineLoading
  | EngineReady
  | EngineError;

export interface ModelInfo {
  id: string;
  name: string;
  size: string;
}

export interface ChatResult {
  reply: string;
  usage?: {
    completionTokens: number;
    promptTokens: number;
    totalTokens: number;
    tokensPerSec: number;
  };
}

export interface EngineAPI {
  state: EngineState;
  models: ModelInfo[];
  loadModel: (modelId: string) => Promise<void>;
  chat: (
    messages: ChatCompletionMessageParam[],
    onToken: (token: string) => void,
    opts?: { temperature?: number; maxTokens?: number },
  ) => Promise<ChatResult>;
  stopGeneration: () => void;
}

// ─── Context ───────────────────────────────────────────────────────

const EngineCtx = createContext<EngineAPI | null>(null);

export function useEngine(): EngineAPI {
  const ctx = useContext(EngineCtx);
  if (!ctx) throw new Error("useEngine must be used within EngineProvider");
  return ctx;
}

export function EngineProvider({
  children,
  cacheBackend = "cache",
}: {
  children: ReactNode;
  cacheBackend?: "cache" | "indexeddb" | "opfs";
}) {
  const [state, setState] = useState<EngineState>({ status: "idle", modelId: null });
  const engineRef = useRef<MLCEngineInterface | null>(null);
  const abortRef = useRef<AbortController | null>(null);
  const loadingRef = useRef(false);

  const models: ModelInfo[] = MODEL_IDS.map((id) => ({
    id,
    name: MODEL_META[id].name,
    size: MODEL_META[id].size,
  }));

  // ── loadModel ──────────────────────────────────────────────────

  const loadModel = useCallback(async (modelId: string) => {
    if (loadingRef.current) return;
    loadingRef.current = true;

    // Dispose previous engine if any
    if (engineRef.current) {
      try {
        engineRef.current.unload();
      } catch {
        /* engine may already be gone */
      }
      engineRef.current = null;
    }

    setState({
      status: "loading",
      modelId,
      progress: { text: "Initialising…", progress: 0, timeElapsed: 0 },
    });

    try {
      const initProgressCallback = (p: LoadProgress) => {
        setState({ status: "loading", modelId, progress: p });
      };

      const engine = await CreateWebWorkerMLCEngine(
        new Worker(new URL("../worker.ts", import.meta.url), {
          type: "module",
        }),
        modelId,
        { initProgressCallback, appConfig: buildAppConfig(cacheBackend) },
      );

      engineRef.current = engine;
      setState({ status: "ready", modelId, engine });
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Unknown error loading model";
      setState({ status: "error", modelId, error: message });
    } finally {
      loadingRef.current = false;
    }
  }, []);

  // ── chat (streaming, rAF-batched) ──────────────────────────────

  const chat = useCallback(
    async (
      messages: ChatCompletionMessageParam[],
      onToken: (token: string) => void,
      opts?: { temperature?: number; maxTokens?: number },
    ): Promise<ChatResult> => {
      const engine = engineRef.current;
      if (!engine) throw new Error("No engine loaded");

      abortRef.current = new AbortController();

      const chunks = await engine.chat.completions.create({
        messages,
        stream: true,
        stream_options: { include_usage: true },
        temperature: opts?.temperature ?? 0.7,
        max_tokens: opts?.maxTokens ?? 1024,
      });

      let buffer = "";
      let rafId: number | null = null;

      function flush() {
        if (buffer.length > 0) {
          onToken(buffer);
          buffer = "";
        }
        rafId = null;
      }

      let fullReply = "";
      let lastUsage: ChatCompletionChunk["usage"] = undefined;
      const startTime = performance.now();

      for await (const chunk of chunks) {
        const delta = chunk.choices[0]?.delta?.content ?? "";
        if (delta) {
          buffer += delta;
          fullReply += delta;
          if (rafId === null) {
            rafId = requestAnimationFrame(flush);
          }
        }
        if (chunk.usage) {
          lastUsage = chunk.usage;
        }
      }

      // Flush any remaining buffered tokens
      if (rafId !== null) {
        cancelAnimationFrame(rafId);
      }
      flush();

      abortRef.current = null;

      const elapsed = (performance.now() - startTime) / 1000;

      return {
        reply: fullReply,
        usage: lastUsage
          ? {
              completionTokens: lastUsage.completion_tokens,
              promptTokens: lastUsage.prompt_tokens,
              totalTokens: lastUsage.total_tokens,
              tokensPerSec:
                elapsed > 0
                  ? Math.round(lastUsage.completion_tokens / elapsed)
                  : 0,
            }
          : undefined,
      };
    },
    [],
  );

  // ── stopGeneration ─────────────────────────────────────────────

  const stopGeneration = useCallback(() => {
    if (abortRef.current) {
      abortRef.current.abort();
      abortRef.current = null;
    }
  }, []);

  // ── Auto-load default model ────────────────────────────────────

  useEffect(() => {
    loadModel(DEFAULT_MODEL);
  }, [loadModel]);

  return (
    <EngineCtx.Provider
      value={{ state, models, loadModel, chat, stopGeneration }}
    >
      {children}
    </EngineCtx.Provider>
  );
}
