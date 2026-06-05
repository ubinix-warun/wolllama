import { BrowserRouter, Routes, Route } from "react-router-dom";
import { LandingPage } from "./pages/Landing";
import { ModelsPage } from "./pages/Models";
import { ModelDetailPage } from "./pages/ModelDetail";
import { ProfilePage } from "./pages/Profile";
import { SubmitPage } from "./pages/Submit";
import { Layout } from "./components/Layout";
import "./index.css";

export default function App() {
  return (
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
  );
}
