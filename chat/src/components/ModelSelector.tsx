import { type ModelInfo } from "../lib/engine";

interface Props {
  models: ModelInfo[];
  selectedId: string | null;
  disabled: boolean;
  onChange: (modelId: string) => void;
}

export function ModelSelector({ models, selectedId, disabled, onChange }: Props) {
  return (
    <select
      value={selectedId ?? ""}
      disabled={disabled}
      onChange={(e) => onChange(e.target.value)}
      className="rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 shadow-sm transition-colors hover:border-gray-300 focus:border-purple-400 focus:outline-none focus:ring-2 focus:ring-purple-200 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200 dark:hover:border-gray-600"
    >
      {models.map((m) => (
        <option key={m.id} value={m.id}>
          {m.name} ({m.size})
        </option>
      ))}
    </select>
  );
}
