import { WebWorkerMLCEngineHandler } from "@mlc-ai/web-llm";

// Handler that resides in the worker thread.
// The main thread sends messages via WebWorkerMLCEngine;
// this handler processes them.
const handler = new WebWorkerMLCEngineHandler();
self.onmessage = (msg: MessageEvent) => {
  handler.onmessage(msg);
};
