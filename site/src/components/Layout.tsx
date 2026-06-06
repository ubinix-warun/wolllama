import { Outlet, Link } from "react-router-dom";
import { useEffect, useState } from "react";
import { ConnectButton, useCurrentAccount } from "@mysten/dapp-kit";
import { getAuthMode } from "../lib/api";

export function Layout() {
  const account = useCurrentAccount();
  const [authMode, setAuthMode] = useState<string>("open");

  useEffect(() => {
    getAuthMode().then(setAuthMode).catch(() => {});
  }, []);

  return (
    <div className="min-h-screen bg-[#080c17] text-[#e0e0e0] font-sans">
      <header className="border-b border-white/5 px-6 py-4 flex items-center justify-between backdrop-blur-md bg-[#080c17]/70 sticky top-0 z-50">
        <Link to="/" className="text-xl font-bold tracking-tight text-white">
          wolllama
        </Link>

        <nav className="flex items-center gap-6 text-sm">
          <Link to="/models" className="hover:text-white transition-colors">
            Models
          </Link>
          <Link to="/submit" className="hover:text-white transition-colors">
            Submit
          </Link>
          {authMode === "sui" && account && (
            <Link to="/profile" className="hover:text-white transition-colors">
              {account.address.slice(0, 6)}...{account.address.slice(-4)}
            </Link>
          )}
          {authMode === "sui" && <ConnectButton />}
        </nav>
      </header>

      <main>
        <Outlet />
      </main>

      <footer className="border-t border-white/5 px-6 py-8 text-center text-sm text-gray-500">
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
