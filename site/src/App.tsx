import { BrowserRouter, Routes, Route } from "react-router-dom";
import { SuiClientProvider, WalletProvider } from "@mysten/dapp-kit";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { lazy, Suspense, useEffect, useState, useMemo } from "react";
import { ModelsPage } from "./pages/Models";
import { ModelDetailPage } from "./pages/ModelDetail";
import { ProfilePage } from "./pages/Profile";
import { SubmitPage } from "./pages/Submit";
import { Layout } from "./components/Layout";

const LandingPage = lazy(() => import("./pages/Landing"));
import "@mysten/dapp-kit/dist/index.css";
import "./index.css";

const queryClient = new QueryClient();

function AppRoutes() {
  const [config, setConfig] = useState<{
    suiNetwork: string;
    suiRpcUrl: string;
  }>({ suiNetwork: "testnet", suiRpcUrl: "" });

  useEffect(() => {
    fetch("/api/config")
      .then(r => r.json())
      .then(c => {
        setConfig({
          suiNetwork: c.sui_network || c.walrus_network || "testnet",
          suiRpcUrl: c.sui_rpc_url || "",
        });
      })
      .catch(() => {});
  }, []);

  const networks = useMemo(() => {
    const base: Record<string, any> = {
      testnet: { url: "https://fullnode.testnet.sui.io:443" },
      mainnet: { url: "https://fullnode.mainnet.sui.io:443" },
    };

    if (config.suiRpcUrl) {
      // Override the active network with custom RPC URL (e.g. Tatum)
      base[config.suiNetwork] = { url: config.suiRpcUrl };
    }

    return base;
  }, [config.suiNetwork, config.suiRpcUrl]);

  return (
    <SuiClientProvider networks={networks} defaultNetwork={config.suiNetwork}>
      <WalletProvider autoConnect>
        <BrowserRouter>
          <Routes>
            <Route element={<Layout />}>
              <Route path="/" element={<Suspense fallback={<div className="py-24" />}><LandingPage /></Suspense>} />
              <Route path="/models" element={<ModelsPage />} />
              <Route path="/models/:id" element={<ModelDetailPage />} />
              <Route path="/profile" element={<ProfilePage />} />
              <Route path="/submit" element={<SubmitPage />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </WalletProvider>
    </SuiClientProvider>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AppRoutes />
    </QueryClientProvider>
  );
}
