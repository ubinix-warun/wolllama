import { useEffect, useState } from "react";
import { useParams, Link } from "react-router-dom";
import { getModel, type Model } from "../lib/api";
import ReactMarkdown from "react-markdown";

export function ModelDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [model, setModel] = useState<Model | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    getModel(Number(id))
      .then(setModel)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) return <div className="max-w-3xl mx-auto px-6 py-12 text-gray-400">Loading...</div>;
  if (error) return <div className="max-w-3xl mx-auto px-6 py-12 text-red-400">Error: {error}</div>;
  if (!model) return <div className="max-w-3xl mx-auto px-6 py-12 text-gray-400">Model not found.</div>;

  const formatSize = (bytes?: number) => {
    if (!bytes) return "Unknown";
    const gb = bytes / 1e9;
    return gb >= 1 ? `${gb.toFixed(1)} GB` : `${(bytes / 1e6).toFixed(1)} MB`;
  };

  return (
    <div className="max-w-3xl mx-auto px-6 py-12">
      <Link to="/models" className="text-sm text-gray-500 hover:text-white transition-colors mb-4 inline-block">
        ← Back to models
      </Link>

      <div className="bg-[#111] border border-[#2a2a2a] rounded-xl p-6 mb-6">
        <div className="flex items-start gap-4 mb-6">
          {model.avatar_url && (
            <img src={model.avatar_url} alt="" className="w-10 h-10 rounded-full mt-1" />
          )}
          <div>
            <h1 className="text-2xl font-bold text-white">{model.display_name}</h1>
            <p className="text-sm text-gray-400">
              by{" "}
              <Link to={`/profile`} className="hover:text-white transition-colors">
                {model.submitter_name}
              </Link>
            </p>
          </div>

          {!model.available && (
            <span className="ml-auto bg-red-900/50 text-red-400 text-xs px-2 py-1 rounded">
              Unavailable
            </span>
          )}
        </div>

        {/* Meta grid */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-6">
          <div>
            <div className="text-xs text-gray-500 uppercase">Size</div>
            <div className="text-white">{formatSize(model.total_size)}</div>
          </div>
          <div>
            <div className="text-xs text-gray-500 uppercase">Blobs</div>
            <div className="text-white">{model.blob_count ?? "—"}</div>
          </div>
          <div>
            <div className="text-xs text-gray-500 uppercase">Added</div>
            <div className="text-white">{new Date(model.created_at).toLocaleDateString()}</div>
          </div>
          <div>
            <div className="text-xs text-gray-500 uppercase">Status</div>
            <div className="text-white">{model.available ? "Available" : "Expired"}</div>
          </div>
        </div>

        {/* Pull command */}
        <div className="bg-[#1a1a2e] rounded-lg p-4 font-mono text-sm flex items-center justify-between">
          <code className="text-green-400">wolllama pull {model.manifest_obj_id}</code>
          <button
            onClick={() => navigator.clipboard.writeText(`wolllama pull ${model.manifest_obj_id}`)}
            className="text-xs bg-[#333] text-white px-2 py-1 rounded hover:bg-[#444] transition-colors"
          >
            Copy
          </button>
        </div>
      </div>

      {/* Description */}
      {model.description_md && (
        <div className="bg-[#111] border border-[#2a2a2a] rounded-xl p-6 prose prose-invert max-w-none">
          <h2 className="text-lg font-semibold text-white mb-4">Description</h2>
          <ReactMarkdown>{model.description_md}</ReactMarkdown>
        </div>
      )}
    </div>
  );
}
