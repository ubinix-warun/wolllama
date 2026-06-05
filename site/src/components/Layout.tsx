import { Outlet, Link } from "react-router-dom";
import { useEffect, useState } from "react";
import { getMe, getLoginUrl, type User } from "../lib/api";

export function Layout() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getMe()
      .then(setUser)
      .catch(() => setUser(null))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="min-h-screen bg-[#0d0d0d] text-[#e0e0e0] font-sans">
      <header className="border-b border-[#2a2a2a] px-6 py-4 flex items-center justify-between">
        <Link to="/" className="text-xl font-bold tracking-tight text-white">
          wolllama
        </Link>

        <nav className="flex items-center gap-6 text-sm">
          <Link to="/models" className="hover:text-white transition-colors">
            Models
          </Link>
          {user ? (
            <>
              <Link to="/submit" className="hover:text-white transition-colors">
                Submit
              </Link>
              <Link to="/profile" className="flex items-center gap-2 hover:text-white transition-colors">
                {user.avatar_url && (
                  <img src={user.avatar_url} alt="" className="w-5 h-5 rounded-full" />
                )}
                {user.username}
              </Link>
            </>
          ) : (
            !loading && (
              <a
                href={getLoginUrl()}
                className="bg-white text-black px-3 py-1.5 rounded-md text-sm font-medium hover:bg-gray-200 transition-colors"
              >
                Sign in with GitHub
              </a>
            )
          )}
        </nav>
      </header>

      <main>
        <Outlet />
      </main>

      <footer className="border-t border-[#2a2a2a] px-6 py-8 text-center text-sm text-gray-500">
        <p>Wolllama — Decentralized model registry powered by Walrus</p>
        <div className="flex justify-center gap-4 mt-2">
          <a href="https://github.com/wolllama-org" className="hover:text-white transition-colors">
            GitHub
          </a>
          <a href="https://twitter.com/wolllama" className="hover:text-white transition-colors">
            Twitter
          </a>
        </div>
      </footer>
    </div>
  );
}
