import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { listModels, type Model } from "../lib/api";
import { ModelCard } from "../components/ModelCard";

export function ModelsPage() {
  const [models, setModels] = useState<Model[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchParams, setSearchParams] = useSearchParams();
  const search = searchParams.get("search") || "";

  useEffect(() => {
    setLoading(true);
    listModels(0, 20, search)
      .then((data) => setModels(data.models))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [search]);

  return (
    <div className="max-w-5xl mx-auto px-6 py-12">
      <h1 className="text-3xl font-bold text-white mb-2">Models</h1>
      <p className="text-gray-400 mb-8">Browse community-published models</p>

      {/* Search */}
      <div className="mb-8">
        <input
          type="text"
          placeholder="Search models..."
          value={search}
          onChange={(e) => {
            setSearchParams(e.target.value ? { search: e.target.value } : {});
          }}
          className="w-full max-w-md px-4 py-2 rounded-lg bg-[#1a1a2e] border border-[#333] text-white placeholder-gray-500 focus:outline-none focus:border-white"
        />
      </div>

      {loading && <p className="text-gray-400">Loading...</p>}
      {error && <p className="text-red-400">Error: {error}</p>}

      {!loading && !error && models.length === 0 && (
        <p className="text-gray-400">No models found. Be the first to submit one!</p>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {models.map((model) => (
          <Link to={`/models/${model.id}`} key={model.id}>
            <ModelCard model={model} />
          </Link>
        ))}
      </div>
    </div>
  );
}
