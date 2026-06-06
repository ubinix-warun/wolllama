import { BrowserRouter, Routes, Route } from "react-router-dom";
import { SuiClientProvider, WalletProvider, createNetworkConfig } from "@mysten/dapp-kit";
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
const { networkConfig } = createNetworkConfig({
  testnet: { url: "https://fullnode.testnet.sui.io:443" } as any,
  mainnet: { url: "https://fullnode.mainnet.sui.io:443" } as any,
});

function AppRoutes() {
  const [suiNetwork, setSuiNetwork] = useState<string>("testnet");

  useEffect(() => {
    fetch("/api/config")
      .then(r => r.json())
      .then(c => {
        const net = c.sui_network || c.walrus_network || "testnet";
        if (net === "mainnet" || net === "testnet") setSuiNetwork(net);
      })
      .catch(() => {});
  }, []);

  return (
    <SuiClientProvider networks={networkConfig} defaultNetwork={suiNetwork as "testnet" | "mainnet"}>
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
