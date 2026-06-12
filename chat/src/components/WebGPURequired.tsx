function detectBrowser(): string {
  const ua = navigator.userAgent;
  if (ua.includes("Edg/")) return "Edge";
  if (ua.includes("Chrome/")) return "Chrome";
  if (ua.includes("Firefox/")) return "Firefox";
  if (ua.includes("Safari/")) return "Safari";
  return "this browser";
}

export function WebGPURequired() {
  const browser = detectBrowser();

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4 dark:bg-gray-950">
      <div className="w-full max-w-md rounded-2xl bg-white p-8 text-center shadow-xl dark:bg-gray-800">
        <div className="mb-4 text-5xl">🖥️</div>
        <h1 className="mb-2 text-2xl font-bold text-gray-900 dark:text-white">
          WebGPU Required
        </h1>
        <p className="mb-4 text-sm text-gray-500 dark:text-gray-400">
          Wolllama Chat runs AI models directly in your browser using WebGPU.
          Your current browser ({browser}) doesn't support WebGPU, or it's not
          enabled.
        </p>
        <div className="mb-6 rounded-lg bg-gray-50 p-4 text-left text-xs text-gray-600 dark:bg-gray-700 dark:text-gray-300">
          <p className="mb-1 font-medium">Supported browsers:</p>
          <ul className="list-inside list-disc space-y-0.5">
            <li>Chrome 113 or newer</li>
            <li>Edge 113 or newer</li>
            <li>
              Firefox Nightly (enable <code className="rounded bg-gray-200 px-1 dark:bg-gray-600">dom.webgpu.enabled</code>)
            </li>
          </ul>
        </div>
        <a
          href="https://www.google.com/chrome/"
          target="_blank"
          rel="noopener noreferrer"
          className="inline-block rounded-xl bg-purple-600 px-6 py-2.5 text-sm font-medium text-white transition-colors hover:bg-purple-700"
        >
          Download Chrome
        </a>
      </div>
    </div>
  );
}
