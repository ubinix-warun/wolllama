import { useCallback, useState } from "react";
import { EngineProvider, useEngine, type ChatResult } from "./lib/engine";
import { useTheme, type Theme } from "./lib/theme";
import { type ChatMessage, ChatWindow } from "./components/ChatWindow";
import { ChatInput } from "./components/ChatInput";
import { ModelSelector } from "./components/ModelSelector";
import { LoadProgress } from "./components/LoadProgress";
import { ErrorBoundary } from "./components/ErrorBoundary";
import { SettingsPanel, type SettingsState } from "./components/SettingsPanel";
import type { ChatCompletionMessageParam } from "@mlc-ai/web-llm";

const DEFAULT_SYSTEM_PROMPT =
  "You are Wally, a helpful AI assistant powered by Wolllama. " +
  "You run entirely in the user's browser using WebGPU. " +
  "Be concise, friendly, and helpful.";

function loadStoredSettings(): SettingsState {
  try {
    const raw = localStorage.getItem("wolllama-chat-settings");
    if (raw) {
      const parsed = JSON.parse(raw);
      return {
        systemPrompt: parsed.systemPrompt ?? DEFAULT_SYSTEM_PROMPT,
        temperature: Number(parsed.temperature) || 0.7,
        maxTokens: Number(parsed.maxTokens) || 1024,
        cacheBackend: parsed.cacheBackend ?? "cache",
      };
    }
  } catch {
    /* ignore */
  }
  return {
    systemPrompt: DEFAULT_SYSTEM_PROMPT,
    temperature: 0.7,
    maxTokens: 1024,
    cacheBackend: "cache" as const,
  };
}

function ChatApp({
  settings,
  onSettingsChange,
}: {
  settings: SettingsState;
  onSettingsChange: (s: SettingsState) => void;
}) {
  const { state, models, loadModel, chat, stopGeneration } = useEngine();
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [streaming, setStreaming] = useState(false);
  const [lastUsage, setLastUsage] = useState<ChatResult["usage"]>(undefined);
  const [theme, setTheme] = useTheme();
  const [settingsOpen, setSettingsOpen] = useState(false);

  const cycleTheme = useCallback(() => {
    const next: Record<Theme, Theme> = {
      system: "light",
      light: "dark",
      dark: "system",
    };
    setTheme(next[theme]);
  }, [theme, setTheme]);

  const themeIcon = theme === "dark" ? "☀️" : theme === "light" ? "🌙" : "💻";

  const loadedModelId = state.modelId;

  const handleSend = useCallback(
    async (text: string) => {
      const userMsg: ChatMessage = { role: "user", content: text };
      const assistantMsg: ChatMessage = { role: "assistant", content: "" };
      setMessages((prev) => [...prev, userMsg, assistantMsg]);
      setStreaming(true);

      const chatMessages: ChatCompletionMessageParam[] = [
        { role: "system", content: settings.systemPrompt },
        ...[...messages, userMsg].map((m) => ({
          role: m.role as "user" | "assistant",
          content: m.content,
        })),
      ];

      try {
        const result = await chat(
          chatMessages,
          (token) => {
            setMessages((prev) => {
              const next = [...prev];
              const last = next[next.length - 1];
              if (last && last.role === "assistant") {
                next[next.length - 1] = {
                  ...last,
                  content: last.content + token,
                };
              }
              return next;
            });
          },
          { temperature: settings.temperature, maxTokens: settings.maxTokens },
        );
        setLastUsage(result.usage);
      } catch (err) {
        const msg =
          err instanceof Error ? err.message : "Generation failed";
        setMessages((prev) => {
          const next = [...prev];
          const last = next[next.length - 1];
          if (last && last.role === "assistant") {
            next[next.length - 1] = {
              ...last,
              content: last.content + `\n\n*Error: ${msg}*`,
            };
          }
          return next;
        });
      } finally {
        setStreaming(false);
      }
    },
    [messages, chat, settings],
  );

  const handleStop = useCallback(() => {
    stopGeneration();
    setStreaming(false);
  }, [stopGeneration]);

  const handleModelChange = useCallback(
    (modelId: string) => {
      setMessages([]);
      loadModel(modelId);
    },
    [loadModel],
  );

  const handleClear = useCallback(() => {
    setMessages([]);
  }, []);

  const isReady = state.status === "ready";
  const isLoading = state.status === "loading";
  const isError = state.status === "error";
  const inputDisabled = !isReady;

  const currentModel = models.find((m) => m.id === loadedModelId);

  return (
    <div className="flex h-screen flex-col bg-white dark:bg-gray-950">
      {/* Header */}
      <ErrorBoundary fallback={
        <header className="flex-shrink-0 border-b border-gray-200 bg-white px-4 py-3 dark:border-gray-700 dark:bg-gray-900">
          <div className="mx-auto flex max-w-3xl items-center">
            <h1 className="text-lg font-bold text-gray-900 dark:text-white">Wolllama Chat</h1>
          </div>
        </header>
      }>
      <header className="flex-shrink-0 border-b border-gray-200 bg-white px-4 py-3 dark:border-gray-700 dark:bg-gray-900">
        <div className="mx-auto flex max-w-3xl items-center justify-between">
          <div className="flex items-center gap-3">
            <h1 className="text-lg font-bold text-gray-900 dark:text-white">
              Wolllama Chat
            </h1>
            {currentModel && (
              <span className="rounded-full bg-purple-100 px-2 py-0.5 text-xs font-medium text-purple-700 dark:bg-purple-900/50 dark:text-purple-300">
                {currentModel.size}
              </span>
            )}
          </div>

          <div className="flex items-center gap-2">
            <ModelSelector
              models={models}
              selectedId={loadedModelId}
              disabled={isLoading}
              onChange={handleModelChange}
            />
            <button
              onClick={handleClear}
              disabled={messages.length === 0}
              className="rounded-lg px-3 py-1.5 text-sm text-gray-500 transition-colors hover:bg-gray-100 disabled:cursor-not-allowed disabled:opacity-30 dark:text-gray-400 dark:hover:bg-gray-800"
              title="Clear chat"
            >
              Clear
            </button>
            <button
              onClick={cycleTheme}
              className="rounded-lg px-2 py-1.5 text-sm transition-colors hover:bg-gray-100 dark:hover:bg-gray-800"
              title={`Theme: ${theme} (click to cycle)`}
            >
              {themeIcon}
            </button>
            <button
              onClick={() => setSettingsOpen(true)}
              className="rounded-lg px-2 py-1.5 text-sm text-gray-500 transition-colors hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800"
              title="Settings"
            >
              ⚙️
            </button>
          </div>
        </div>
      </header>
      </ErrorBoundary>

      {/* Error state */}
      {isError && (
        <div className="flex-shrink-0 bg-red-50 px-4 py-3 dark:bg-red-900/20">
          <div className="mx-auto flex max-w-3xl items-center justify-between">
            <p className="text-sm text-red-700 dark:text-red-400">
              {(state as { error: string }).error}
            </p>
            <button
              onClick={() => loadModel(loadedModelId ?? models[0].id)}
              className="rounded-lg bg-red-600 px-4 py-1.5 text-sm font-medium text-white hover:bg-red-700"
            >
              Retry
            </button>
          </div>
        </div>
      )}

      {/* Chat */}
      <ErrorBoundary fallback={
        <div className="flex flex-1 items-center justify-center px-4">
          <p className="text-sm text-gray-400 dark:text-gray-500">Chat window crashed. Try clearing the chat.</p>
        </div>
      }>
      <ChatWindow
        messages={messages}
        streaming={streaming}
        emptyState={
          <p className="text-center text-gray-400 dark:text-gray-500">
            Send a message to start chatting with{" "}
            {currentModel?.name ?? "Wally"}
          </p>
        }
      />
      </ErrorBoundary>

      {/* Input */}
      <ChatInput
        disabled={inputDisabled}
        generating={streaming}
        onSend={handleSend}
        onStop={handleStop}
        usage={lastUsage}
      />

      {/* Loading overlay */}
      {isLoading && state.status === "loading" && (
        <LoadProgress
          progress={state.progress}
          modelName={currentModel?.name ?? "model"}
        />
      )}

      {/* Settings modal */}
      <SettingsPanel
        open={settingsOpen}
        onClose={() => setSettingsOpen(false)}
        onChange={onSettingsChange}
      />
    </div>
  );
}

export default function App() {
  const [settings, setSettings] = useState<SettingsState>(loadStoredSettings);

  return (
    <ErrorBoundary>
      <EngineProvider cacheBackend={settings.cacheBackend}>
        <ChatApp settings={settings} onSettingsChange={setSettings} />
      </EngineProvider>
    </ErrorBoundary>
  );
}
