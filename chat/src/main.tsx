import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import { WebGPURequired } from "./components/WebGPURequired";
import "./index.css";

function rootElement(): HTMLElement {
  const el = document.getElementById("root");
  if (!el) throw new Error("Missing #root element");
  return el;
}

function webgpuAvailable(): boolean {
  return typeof navigator !== "undefined" && "gpu" in navigator;
}

createRoot(rootElement()).render(
  <StrictMode>
    {webgpuAvailable() ? <App /> : <WebGPURequired />}
  </StrictMode>
);
