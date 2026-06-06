import type { Model } from "../lib/api";

export function ModelCard({ model }: { model: Model }) {
  const formatSize = (bytes?: number) => {
    if (!bytes) return "Unknown";
    const gb = bytes / 1e9;
    return gb >= 1 ? `${gb.toFixed(1)} GB` : `${(bytes / 1e6).toFixed(1)} MB`;
  };

  return (
    <div className="bg-[#0a0e1a] border border-white/10 rounded-xl p-5 hover:border-blue-400/40 transition-colors h-full">
      <div className="flex items-start gap-3 mb-3">
        {model.avatar_url && (
          <img src={model.avatar_url} alt="" className="w-6 h-6 rounded-full mt-0.5" />
        )}
        <div className="min-w-0">
          <h3 className="text-white font-medium truncate">{model.display_name}</h3>
          <p className="text-xs text-gray-500">
            {model.submitter_address
              ? `${model.submitter_address.slice(0, 6)}...${model.submitter_address.slice(-4)}`
              : model.submitter_name || "anonymous"}
          </p>
        </div>
        {!model.available && (
          <span className="ml-auto shrink-0 bg-red-900/50 text-red-400 text-xs px-1.5 py-0.5 rounded">
            Expired
          </span>
        )}
      </div>

      <div className="flex gap-4 text-xs text-gray-500">
        <span>{formatSize(model.total_size)}</span>
        <span>{model.blob_count} blobs</span>
        <span>{new Date(model.created_at).toLocaleDateString()}</span>
      </div>
    </div>
  );
}
