import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { getMe, getUserModels, type User, type Model } from "../lib/api";
import { ModelCard } from "../components/ModelCard";

export function ProfilePage() {
  const [user, setUser] = useState<User | null>(null);
  const [models, setModels] = useState<Model[]>([]);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    getMe()
      .then((u) => {
        setUser(u);
        return getUserModels(u.id);
      })
      .then((data) => setModels(data.models))
      .catch(() => {
        // Not authenticated — redirect to login
        navigate("/");
      })
      .finally(() => setLoading(false));
  }, [navigate]);

  if (loading) return <div className="max-w-4xl mx-auto px-6 py-12 text-gray-400">Loading...</div>;
  if (!user) return null;

  return (
    <div className="max-w-4xl mx-auto px-6 py-12">
      <div className="flex items-center gap-4 mb-8">
        {user.avatar_url && (
          <img src={user.avatar_url} alt="" className="w-12 h-12 rounded-full" />
        )}
        <div>
          <h1 className="text-2xl font-bold text-white">{user.username}</h1>
          <p className="text-gray-400 text-sm">GitHub ID: {user.github_id}</p>
        </div>
        <Link
          to="/submit"
          className="ml-auto bg-white text-black px-4 py-2 rounded-lg text-sm font-medium hover:bg-gray-200 transition-colors"
        >
          Submit Model
        </Link>
      </div>

      <h2 className="text-lg font-semibold text-white mb-4">
        Published Models ({models.length})
      </h2>

      {models.length === 0 ? (
        <p className="text-gray-400">No models published yet.</p>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {models.map((model) => (
            <Link to={`/models/${model.id}`} key={model.id}>
              <ModelCard model={model} />
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
