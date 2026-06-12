import { type LoadProgress as LoadProgressType } from "../lib/engine";

interface Props {
  progress: LoadProgressType;
  modelName: string;
  onCancel?: () => void;
}

export function LoadProgress({ progress, modelName, onCancel }: Props) {
  const pct = Math.round(progress.progress * 100);
  const elapsed = Math.round(progress.timeElapsed);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
      <div className="w-full max-w-md rounded-2xl bg-white p-8 shadow-2xl dark:bg-gray-800">
        <h2 className="mb-1 text-lg font-semibold text-gray-900 dark:text-white">
          Loading {modelName}
        </h2>
        <p className="mb-4 text-sm text-gray-500 dark:text-gray-400">
          {progress.text}
        </p>

        {/* Progress bar */}
        <div className="mb-2 h-3 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-gray-700">
          <div
            className="h-full rounded-full bg-gradient-to-r from-purple-500 to-blue-500 transition-all duration-300"
            style={{ width: `${Math.max(pct, 2)}%` }}
          />
        </div>

        <div className="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
          <span>
            {pct}% — {elapsed}s elapsed
          </span>
          {onCancel && (
            <button
              onClick={onCancel}
              className="text-purple-600 hover:text-purple-700 dark:text-purple-400"
            >
              Cancel
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
