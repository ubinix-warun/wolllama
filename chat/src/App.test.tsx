import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { EngineProvider } from "./lib/engine";
import "@testing-library/jest-dom/vitest";

// Mock web-llm to avoid WebGPU/WASM dependencies in test environment
vi.mock("@mlc-ai/web-llm", () => ({
  CreateWebWorkerMLCEngine: vi.fn(),
  prebuiltAppConfig: {
    model_list: [
      {
        model_id: "Qwen2-0.5B-Instruct-q4f16_1-MLC",
        model: "https://huggingface.co/mlc-ai/Qwen2-0.5B-Instruct-q4f16_1-MLC",
        model_lib: "",
        vram_required_MB: 500,
        low_resource_required: false,
        required_features: [],
      },
      {
        model_id: "Llama-3.2-1B-Instruct-q4f16_1-MLC",
        model: "https://huggingface.co/mlc-ai/Llama-3.2-1B-Instruct-q4f16_1-MLC",
        model_lib: "",
        vram_required_MB: 1000,
        low_resource_required: false,
        required_features: [],
      },
    ],
    use_cdn: true,
  },
}));

function SmokeApp() {
  return (
    <EngineProvider>
      <div data-testid="smoke">Wolllama Chat</div>
    </EngineProvider>
  );
}

describe("Chat App", () => {
  it("renders without crashing", () => {
    render(<SmokeApp />);
    expect(screen.getByTestId("smoke")).toBeInTheDocument();
  });
});
