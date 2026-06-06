import { useState, useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { useCurrentAccount, useSignPersonalMessage } from "@mysten/dapp-kit";
import { submitModel, getAuthMode } from "../lib/api";

interface ManifestPreview {
  manifest_obj_id: string;
  name: string;
  tag: string;
  blob_count: number;
  total_size: number;
}

export function SubmitPage() {
  const [manifestObjId, setManifestObjId] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [descriptionMd, setDescriptionMd] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [preview, setPreview] = useState<ManifestPreview | null>(null);
  const [previewing, setPreviewing] = useState(false);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const [authMode, setAuthMode] = useState<string>("open");
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const navigate = useNavigate();
  const account = useCurrentAccount();
  const { mutateAsync: signPersonalMessage } = useSignPersonalMessage();

  useEffect(() => {
    getAuthMode().then(setAuthMode);
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
        if (!displayName) {
          setDisplayName(data.name);
        }
      } catch (err) {
        setPreviewError(err instanceof Error ? err.message : "Failed to fetch manifest");
        setPreview(null);
      } finally {
        setPreviewing(false);
      }
    }, 600);

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
    if (authMode === "sui" && !account) {
      setError("Connect your Sui wallet to submit.");
      return;
    }

    setSubmitting(true);
    try {
      let submitterAddress: string | undefined;
      let signature: string | undefined;

      if (authMode === "sui" && account) {
        // Build payload to sign
        const payload = JSON.stringify({
          manifest_obj_id: manifestObjId.trim(),
          display_name: displayName.trim(),
          timestamp: Date.now(),
        });
        const payloadBytes = new TextEncoder().encode(payload);

        // Sign with Sui wallet
        const result = await signPersonalMessage({ message: payloadBytes });
        signature = result.signature;
        submitterAddress = account.address;
      }

      // Submit
      const model = await submitModel(
        manifestObjId.trim(),
        displayName.trim(),
        descriptionMd.trim() || undefined,
        submitterAddress,
        signature
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

  // Only block in sui mode when no wallet connected
  if (authMode === "sui" && !account) {
    return (
      <div className="max-w-lg mx-auto px-6 py-12 text-center">
        <h1 className="text-2xl font-bold text-white mb-4">Connect Wallet to Submit</h1>
        <p className="text-gray-400 mb-6">
          Connect your Sui wallet to sign and submit models.
        </p>
      </div>
    );
  }

  return (
    <div className="max-w-lg mx-auto px-6 py-12">
      <h1 className="text-2xl font-bold text-white mb-2">Submit a Model</h1>
      <p className="text-gray-400 mb-2">
        Paste the manifest object ID from <code className="text-green-400">wolllama push</code>.
      </p>
      {authMode === "sui" && account && (
        <p className="text-sm text-amber-400 mb-8">
          Signed by: {account.address.slice(0, 10)}...{account.address.slice(-6)}
        </p>
      )}
      {authMode === "open" && (
        <p className="text-sm text-gray-500 mb-8">
          Open mode — no authentication required.
        </p>
      )}

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
          {previewing && <div className="mt-3 text-sm text-gray-400">Fetching manifest...</div>}
          {previewError && <div className="mt-3 text-sm text-red-400">⚠ {previewError}</div>}
          {preview && (
            <div className="mt-3 bg-[#1a1a2e] border border-[#333] rounded-lg p-3 text-sm">
              <div className="text-green-400 font-mono text-xs mb-2">✓ Manifest found on Walrus</div>
              <div className="grid grid-cols-2 gap-2">
                <div><span className="text-gray-500">Model:</span> <span className="text-white">{preview.name}</span></div>
                <div><span className="text-gray-500">Size:</span> <span className="text-white">{formatSize(preview.total_size)}</span></div>
                <div><span className="text-gray-500">Blobs:</span> <span className="text-white">{preview.blob_count}</span></div>
                <div><span className="text-gray-500">Tag:</span> <span className="text-white">{preview.tag || "—"}</span></div>
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
          {submitting
            ? (authMode === "sui" ? "Signing & Submitting..." : "Submitting...")
            : (authMode === "sui" ? "Sign & Submit" : "Submit Model")}
        </button>
      </form>
    </div>
  );
}
