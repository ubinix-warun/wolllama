import { BrowserRouter, Routes, Route } from "react-router-dom";
import { SuiClientProvider, WalletProvider } from "@mysten/dapp-kit";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { LandingPage } from "./pages/Landing";
import { ModelsPage } from "./pages/Models";
import { ModelDetailPage } from "./pages/ModelDetail";
import { ProfilePage } from "./pages/Profile";
import { SubmitPage } from "./pages/Submit";
import { Layout } from "./components/Layout";
import "@mysten/dapp-kit/dist/index.css";
import "./index.css";

const queryClient = new QueryClient();

function AppRoutes() {
  const [networks, setNetworks] = useState<Record<string, any>>({
    testnet: { url: "https://fullnode.testnet.sui.io:443" },
    mainnet: { url: "https://fullnode.mainnet.sui.io:443" },
  });
  const [defaultNetwork, setDefaultNetwork] = useState("testnet");

  useEffect(() => {
    fetch("/api/config")
      .then(r => r.json())
      .then(c => {
        const net = c.sui_network || c.walrus_network || "testnet";
        const rpcUrl = c.sui_rpc_url;
        if (rpcUrl) {
          setNetworks({ [net]: { url: rpcUrl } });
        }
        setDefaultNetwork(net);
      })
      .catch(() => {});
  }, []);

  return (
    <SuiClientProvider networks={networks} defaultNetwork={defaultNetwork}>
      <WalletProvider autoConnect>
        <BrowserRouter>
          <Routes>
            <Route element={<Layout />}>
              <Route path="/" element={<LandingPage />} />
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
