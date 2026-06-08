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

  const isSui = authMode === "sui";

  return (
    <div className="min-h-screen bg-[#0d0d0d] text-[#e0e0e0] font-sans">
      <header className="border-b border-[#2a2a2a] px-6 py-4 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Link to="/" className="text-xl font-bold tracking-tight text-white">
            wolllama
          </Link>
          {!isSui && (
            <span className="text-[10px] px-2 py-0.5 rounded-full bg-[#1a1a2e] text-gray-500 border border-[#2a2a2a]">
              coming soon
            </span>
          )}
        </div>

        <nav className="flex items-center gap-6 text-sm">
          <Link to="/models" className="hover:text-white transition-colors">
            Models
          </Link>
          {isSui && (
            <>
              <Link to="/submit" className="hover:text-white transition-colors">
                Submit
              </Link>
              {account && (
                <Link to="/profile" className="hover:text-white transition-colors">
                  {account.address.slice(0, 6)}...{account.address.slice(-4)}
                </Link>
              )}
              <ConnectButton />
            </>
          )}
        </nav>
      </header>

      <main>
        <Outlet />
      </main>

      <footer className="border-t border-[#2a2a2a] px-6 py-8 text-center text-sm text-gray-500">
        <p>Wolllama — Decentralized model registry powered by Walrus</p>
        <div className="flex justify-center gap-4 mt-2">
          <a href="https://wolllama.xyz" className="hover:text-white transition-colors">
            Website
          </a>
          <a href="https://github.com/ubinix-warun/wolllama" className="hover:text-white transition-colors">
            GitHub
          </a>
          <a href="https://x.com/wolllama" className="hover:text-white transition-colors">
            X / Twitter
          </a>
        </div>
      </footer>
    </div>
  );
}
