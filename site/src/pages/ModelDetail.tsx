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

      {/* Blob Details */}
      <BlobDetails manifestJSON={model.manifest_json} />

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

// BlobDetails parses the manifest JSON and renders individual blob details.
function BlobDetails({ manifestJSON }: { manifestJSON?: string }) {
  if (!manifestJSON) return null;

  let wm: {
    ollamaManifest?: {
      config?: { digest: string; mediaType: string; size: number };
      layers?: { digest: string; mediaType: string; size: number }[];
    };
    blobs?: Record<string, { single?: string; chunks?: string[] }>;
  };

  try {
    wm = JSON.parse(manifestJSON);
  } catch {
    return null;
  }

  const blobs = wm.blobs || {};
  const config = wm.ollamaManifest?.config;
  const layers = wm.ollamaManifest?.layers || [];

  const allBlobs: { digest: string; mediaType: string; size: number; kind: string }[] = [];
  if (config) allBlobs.push({ ...config, kind: "model config" });
  for (const l of layers) {
    const kind = l.mediaType.includes("license") ? "license"
      : l.mediaType.includes("params") ? "params"
      : l.mediaType.includes("image.model") ? "model"
      : "blob";
    allBlobs.push({ ...l, kind });
  }

  const formatSize = (bytes: number) => {
    if (bytes >= 1_000_000_000) return `${(bytes / 1_000_000_000).toFixed(1)} GB`;
    if (bytes >= 1_000_000) return `${(bytes / 1_000_000).toFixed(0)} MB`;
    if (bytes >= 1_000) return `${(bytes / 1_000).toFixed(0)} KB`;
    return `${bytes} B`;
  };

  return (
    <div className="bg-[#111] border border-[#2a2a2a] rounded-xl p-6 mb-6">
      <h2 className="text-lg font-semibold text-white mb-4">Blob Details</h2>
      <div className="space-y-3">
        {allBlobs.map((blob) => {
          const ref = blobs[blob.digest];
          const isChunked = ref?.chunks && ref.chunks.length > 0;
          const ids: string[] = isChunked ? ref.chunks! : ref?.single ? [ref.single] : [];
          const isSmallBlob = blob.kind === "model config" || blob.kind === "license" || blob.kind === "params";

          return (
            <div key={blob.digest} className="bg-[#0d0d0d] border border-[#2a2a2a] rounded-lg p-3">
              <div className="flex items-center gap-2 mb-2">
                <span className={`text-xs px-1.5 py-0.5 rounded ${
                  blob.kind === "model" ? "bg-blue-900/50 text-blue-400" :
                  blob.kind === "model config" ? "bg-purple-900/50 text-purple-400" :
                  blob.kind === "license" ? "bg-green-900/50 text-green-400" :
                  blob.kind === "params" ? "bg-amber-900/50 text-amber-400" :
                  "bg-gray-900/50 text-gray-400"
                }`}>{blob.kind}</span>
                <span className="text-gray-400 text-xs">{formatSize(blob.size)}</span>
              </div>
              <div className="font-mono text-xs text-gray-500 break-all mb-2">
                {blob.digest}
              </div>
              {isChunked ? (
                <div>
                  <div className="text-xs text-amber-400 mb-1">
                    Split into {ref.chunks!.length} × 256 MB chunks
                  </div>
                  <div className="space-y-1">
                    {ids.map((id, i) => (
                      <div key={i} className="font-mono text-xs text-gray-500 pl-2 border-l border-[#333]">
                        chunk {i + 1}: {id}
                      </div>
                    ))}
                  </div>
                </div>
              ) : (
                <div className="font-mono text-xs text-gray-500">
                  Walrus ID: {ids.length > 0 ? ids[0] : "—"}
                </div>
              )}
              {isSmallBlob && ids.length > 0 && (
                <BlobContentFetcher objId={ids[0]} kind={blob.kind} />
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}

// BlobContentFetcher fetches and displays the content of small metadata blobs.
function BlobContentFetcher({ objId, kind }: { objId: string; kind: string }) {
  const [expanded, setExpanded] = useState(false);
  const [content, setContent] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const toggle = async () => {
    if (expanded) { setExpanded(false); return; }
    setExpanded(true);
    if (content !== null) return; // already fetched

    setLoading(true);
    try {
      const res = await fetch(`/api/blobs/${encodeURIComponent(objId)}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const text = await res.text();

      // Try to pretty-print JSON
      try {
        const parsed = JSON.parse(text);
        setContent(JSON.stringify(parsed, null, 2));
      } catch {
        setContent(text);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to fetch");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mt-2">
      <button
        onClick={toggle}
        className="text-xs text-gray-400 hover:text-white transition-colors"
      >
        {expanded ? "▾ Hide" : "▸ View"} {kind} content
      </button>
      {expanded && loading && (
        <div className="mt-1 text-xs text-gray-500">Loading...</div>
      )}
      {expanded && error && (
        <div className="mt-1 text-xs text-red-400">{error}</div>
      )}
      {expanded && content && (
        <pre className="mt-1 bg-[#1a1a2e] rounded p-2 text-xs text-gray-300 overflow-x-auto max-h-64 overflow-y-auto font-mono whitespace-pre-wrap">
          {content}
        </pre>
      )}
    </div>
  );
}
