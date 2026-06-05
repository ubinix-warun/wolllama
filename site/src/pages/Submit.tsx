import { useState, useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { getMe, submitModel, getLoginUrl, type User } from "../lib/api";

interface ManifestPreview {
  manifest_obj_id: string;
  name: string;
  tag: string;
  blob_count: number;
  total_size: number;
}

export function SubmitPage() {
  const [user, setUser] = useState<User | null>(null);
  const [manifestObjId, setManifestObjId] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [descriptionMd, setDescriptionMd] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [preview, setPreview] = useState<ManifestPreview | null>(null);
  const [previewing, setPreviewing] = useState(false);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    getMe()
      .then(setUser)
      .catch(() => {});
  }, []);

  // Pre-fetch manifest info when object ID changes (debounced)
  useEffect(() => {
    if (!manifestObjId.trim()) {
      setPreview(null);
      setPreviewError(null);
      return;
    }

    if (debounceRef.current) clearTimeout(debounceRef.current);

    debounceRef.current = setTimeout(async () => {
      setPreviewing(true);
      setPreviewError(null);
      try {
        const res = await fetch(`/api/manifest/preview?obj_id=${encodeURIComponent(manifestObjId.trim())}`);
        if (!res.ok) {
          const err = await res.json().catch(() => ({ error: "Unknown error" }));
          throw new Error(err.error || `HTTP ${res.status}`);
        }
        const data: ManifestPreview = await res.json();
        setPreview(data);
        // Auto-fill display name from manifest if not user-edited yet
        if (!displayName) {
          setDisplayName(data.name);
        }
      } catch (err) {
        setPreviewError(err instanceof Error ? err.message : "Failed to fetch manifest");
        setPreview(null);
      } finally {
        setPreviewing(false);
      }
    }, 600); // 600ms debounce

    return () => { if (debounceRef.current) clearTimeout(debounceRef.current); };
  }, [manifestObjId]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!manifestObjId.trim()) {
      setError("Manifest object ID is required.");
      return;
    }
    if (!displayName.trim()) {
      setError("Display name is required.");
      return;
    }

    setSubmitting(true);
    try {
      const model = await submitModel(
        manifestObjId.trim(),
        displayName.trim(),
        descriptionMd.trim() || undefined
      );
      navigate(`/models/${model.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Submission failed");
    } finally {
      setSubmitting(false);
    }
  };

  const formatSize = (bytes: number) => {
    if (bytes >= 1_000_000_000) return `${(bytes / 1_000_000_000).toFixed(1)} GB`;
    if (bytes >= 1_000_000) return `${(bytes / 1_000_000).toFixed(0)} MB`;
    if (bytes >= 1_000) return `${(bytes / 1_000).toFixed(0)} KB`;
    return `${bytes} B`;
  };

  if (!user) {
    return (
      <div className="max-w-lg mx-auto px-6 py-12 text-center">
        <h1 className="text-2xl font-bold text-white mb-4">Sign in to submit</h1>
        <p className="text-gray-400 mb-6">You need to sign in before submitting a model.</p>
        <a
          href={getLoginUrl()}
          className="inline-block bg-white text-black px-6 py-3 rounded-lg font-medium hover:bg-gray-200 transition-colors"
        >
          Sign in with GitHub
        </a>
      </div>
    );
  }

  return (
    <div className="max-w-lg mx-auto px-6 py-12">
      <h1 className="text-2xl font-bold text-white mb-2">Submit a Model</h1>
      <p className="text-gray-400 mb-8">
        Paste the manifest object ID from <code className="text-green-400">wolllama push</code> to list your model.
      </p>

      <form onSubmit={handleSubmit} className="space-y-5">
        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1">
            Manifest Object ID <span className="text-red-400">*</span>
          </label>
          <input
            type="text"
            value={manifestObjId}
            onChange={(e) => setManifestObjId(e.target.value)}
            placeholder="O1ABCdef...xyz"
            className="w-full px-4 py-2 rounded-lg bg-[#1a1a2e] border border-[#333] text-white placeholder-gray-500 focus:outline-none focus:border-white font-mono text-sm"
          />
          <p className="text-xs text-gray-500 mt-1">
            Paste the ID from the "wolllama push" output
          </p>

          {/* Preview card */}
          {previewing && (
            <div className="mt-3 text-sm text-gray-400">Fetching manifest...</div>
          )}
          {previewError && (
            <div className="mt-3 text-sm text-red-400">⚠ {previewError}</div>
          )}
          {preview && (
            <div className="mt-3 bg-[#1a1a2e] border border-[#333] rounded-lg p-3 text-sm">
              <div className="text-green-400 font-mono text-xs mb-2">✓ Manifest found on Walrus</div>
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <span className="text-gray-500">Model:</span>{" "}
                  <span className="text-white">{preview.name}</span>
                </div>
                <div>
                  <span className="text-gray-500">Size:</span>{" "}
                  <span className="text-white">{formatSize(preview.total_size)}</span>
                </div>
                <div>
                  <span className="text-gray-500">Blobs:</span>{" "}
                  <span className="text-white">{preview.blob_count}</span>
                </div>
                <div>
                  <span className="text-gray-500">Tag:</span>{" "}
                  <span className="text-white">{preview.tag || "—"}</span>
                </div>
              </div>
            </div>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1">
            Display Name <span className="text-red-400">*</span>
          </label>
          <input
            type="text"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            placeholder="llama3.2:3b-q4_K_M"
            className="w-full px-4 py-2 rounded-lg bg-[#1a1a2e] border border-[#333] text-white placeholder-gray-500 focus:outline-none focus:border-white"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1">
            Description <span className="text-gray-500">(supports Markdown)</span>
          </label>
          <textarea
            value={descriptionMd}
            onChange={(e) => setDescriptionMd(e.target.value)}
            placeholder="A brief description of this model..."
            rows={5}
            className="w-full px-4 py-2 rounded-lg bg-[#1a1a2e] border border-[#333] text-white placeholder-gray-500 focus:outline-none focus:border-white resize-y font-mono text-sm"
          />
        </div>

        {error && (
          <div className="bg-red-900/30 border border-red-800 text-red-400 px-4 py-2 rounded-lg text-sm">
            {error}
          </div>
        )}

        <button
          type="submit"
          disabled={submitting}
          className="w-full bg-white text-black py-2.5 rounded-lg font-medium hover:bg-gray-200 transition-colors disabled:opacity-50"
        >
          {submitting ? "Submitting..." : "Submit Model"}
        </button>
      </form>
    </div>
  );
}
