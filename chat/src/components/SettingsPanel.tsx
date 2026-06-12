import { useCallback, useEffect, useState } from "react";

export type CacheBackend = "cache" | "indexeddb" | "opfs";

export interface SettingsState {
  systemPrompt: string;
  temperature: number;
  maxTokens: number;
  cacheBackend: CacheBackend;
}

const DEFAULT_SETTINGS: SettingsState = {
  systemPrompt:
    "You are Wally, a helpful AI assistant powered by Wolllama. " +
    "You run entirely in the user's browser using WebGPU. " +
    "Be concise, friendly, and helpful.",
  temperature: 0.7,
  maxTokens: 1024,
  cacheBackend: "cache",
};

const STORAGE_KEY = "wolllama-chat-settings";

function loadSettings(): SettingsState {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) {
      const parsed = JSON.parse(raw);
      return {
        systemPrompt: parsed.systemPrompt ?? DEFAULT_SETTINGS.systemPrompt,
        temperature: Number(parsed.temperature) || DEFAULT_SETTINGS.temperature,
        maxTokens: Number(parsed.maxTokens) || DEFAULT_SETTINGS.maxTokens,
        cacheBackend: parsed.cacheBackend ?? DEFAULT_SETTINGS.cacheBackend,
      };
    }
  } catch {
    /* ignore */
  }
  return DEFAULT_SETTINGS;
}

function saveSettings(s: SettingsState) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(s));
  } catch {
    /* ignore */
  }
}

interface Props {
  open: boolean;
  onClose: () => void;
  onChange: (settings: SettingsState) => void;
}

export function SettingsPanel({ open, onClose, onChange }: Props) {
  const [settings, setSettings] = useState<SettingsState>(loadSettings);

  // Notify parent on open (so it can sync)
  useEffect(() => {
    if (open) {
      setSettings(loadSettings());
    }
  }, [open]);

  const handleSave = useCallback(() => {
    saveSettings(settings);
    onChange(settings);
    onClose();
  }, [settings, onChange, onClose]);

  const handleReset = useCallback(() => {
    setSettings(DEFAULT_SETTINGS);
    saveSettings(DEFAULT_SETTINGS);
    onChange(DEFAULT_SETTINGS);
    onClose();
  }, [onChange, onClose]);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-40 flex items-center justify-center bg-black/40 backdrop-blur-sm">
      <div className="w-full max-w-lg rounded-2xl bg-white p-6 shadow-2xl dark:bg-gray-800">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            Settings
          </h2>
          <button
            onClick={onClose}
            className="rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-700 dark:hover:text-gray-300"
          >
            <svg
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
            >
              <path d="M18 6L6 18M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* System prompt */}
        <label className="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">
          System prompt
        </label>
        <textarea
          rows={4}
          value={settings.systemPrompt}
          onChange={(e) =>
            setSettings((s) => ({ ...s, systemPrompt: e.target.value }))
          }
          className="mb-4 w-full resize-none rounded-lg border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-900 placeholder-gray-400 focus:border-purple-400 focus:outline-none focus:ring-2 focus:ring-purple-200 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 dark:focus:border-purple-500"
        />

        {/* Temperature slider */}
        <label className="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">
          Temperature: {settings.temperature.toFixed(1)}
        </label>
        <input
          type="range"
          min="0"
          max="2"
          step="0.1"
          value={settings.temperature}
          onChange={(e) =>
            setSettings((s) => ({
              ...s,
              temperature: parseFloat(e.target.value),
            }))
          }
          className="mb-4 w-full accent-purple-600"
        />

        {/* Max tokens slider */}
        <label className="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">
          Max tokens: {settings.maxTokens}
        </label>
        <input
          type="range"
          min="64"
          max="4096"
          step="64"
          value={settings.maxTokens}
          onChange={(e) =>
            setSettings((s) => ({
              ...s,
              maxTokens: parseInt(e.target.value, 10),
            }))
          }
          className="mb-6 w-full accent-purple-600"
        />

        {/* Cache backend */}
        <label className="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">
          Model cache
        </label>
        <select
          value={settings.cacheBackend}
          onChange={(e) =>
            setSettings((s) => ({
              ...s,
              cacheBackend: e.target.value as CacheBackend,
            }))
          }
          className="mb-6 w-full rounded-lg border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-900 focus:border-purple-400 focus:outline-none focus:ring-2 focus:ring-purple-200 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
        >
          <option value="cache">Cache API (default)</option>
          <option value="indexeddb">IndexedDB</option>
          <option value="opfs">OPFS (fastest)</option>
        </select>

        {/* Actions */}
        <div className="flex items-center justify-between">
          <button
            onClick={handleReset}
            className="rounded-lg px-4 py-2 text-sm text-gray-500 transition-colors hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700"
          >
            Reset defaults
          </button>
          <button
            onClick={handleSave}
            className="rounded-xl bg-purple-600 px-6 py-2 text-sm font-medium text-white transition-colors hover:bg-purple-700"
          >
            Save
          </button>
        </div>
      </div>
    </div>
  );
}
