import { Link } from "react-router-dom";

export function LandingPage() {
  return (
    <div>
      {/* Hero Section */}
      <section className="py-24 px-6 text-center">
        <h1 className="text-5xl font-bold text-white mb-4 max-w-2xl mx-auto leading-tight">
          Your models. Your storage.
          <br />
          No limits.
        </h1>
        <p className="text-lg text-gray-400 mb-10 max-w-xl mx-auto">
          Push and pull Ollama models to decentralized storage on Walrus.
          No central registry. No gatekeepers. Just you and your models.
        </p>

        {/* Terminal mockup */}
        <div className="max-w-lg mx-auto mb-8">
          <div className="bg-[#1a1a2e] rounded-lg overflow-hidden border border-[#333] shadow-2xl">
            <div className="flex items-center gap-2 px-4 py-2 bg-[#111] border-b border-[#333]">
              <span className="w-3 h-3 rounded-full bg-red-500" />
              <span className="w-3 h-3 rounded-full bg-yellow-500" />
              <span className="w-3 h-3 rounded-full bg-green-500" />
              <span className="text-xs text-gray-500 ml-2">terminal</span>
            </div>
            <div className="p-4 text-left font-mono text-sm">
              <div className="text-green-400 mb-1">$ wolllama push llama3.2:3b-q4_K_M</div>
              <div className="text-gray-300 mb-1">[1/4] sha256:abc123... uploading (1 KB)</div>
              <div className="text-gray-300 mb-1">[2/4] sha256:def456... uploading (44 B)</div>
              <div className="text-gray-300 mb-1">[3/4] sha256:ghi789... uploading (1.2 GB)</div>
              <div className="text-gray-300 mb-2">[4/4] sha256:jkl012... uploading (3.5 GB)</div>
              <div className="text-green-400 mb-1">✓ llama3.2:3b-q4_K_M pushed to Walrus</div>
              <div className="text-gray-300 mb-1">  Manifest: O1ABCdef...xyz</div>
              <div className="text-blue-400">  Share: wolllama pull O1ABCdef...xyz</div>
            </div>
          </div>
        </div>
      </section>

      {/* How It Works */}
      <section className="py-20 px-6 bg-[#111]">
        <h2 className="text-3xl font-bold text-white text-center mb-12">How It Works</h2>
        <div className="max-w-4xl mx-auto grid grid-cols-1 md:grid-cols-3 gap-8">
          <div className="text-center p-6 rounded-xl border border-[#2a2a2a] bg-[#0d0d0d]">
            <div className="text-2xl mb-3">📤</div>
            <h3 className="text-lg font-semibold text-white mb-2">Push</h3>
            <p className="text-sm text-gray-400">
              Upload models directly from your Ollama store to Walrus decentralized storage.
            </p>
            <code className="block mt-3 text-xs text-green-400 bg-[#1a1a2e] rounded p-2">
              wolllama push llama3.2:latest
            </code>
          </div>

          <div className="text-center p-6 rounded-xl border border-[#2a2a2a] bg-[#0d0d0d]">
            <div className="text-2xl mb-3">🔗</div>
            <h3 className="text-lg font-semibold text-white mb-2">Share</h3>
            <p className="text-sm text-gray-400">
              Get a manifest object ID. Share it with anyone, or submit to the registry for discovery.
            </p>
            <code className="block mt-3 text-xs text-blue-400 bg-[#1a1a2e] rounded p-2">
              O1ABCdef...xyz
            </code>
          </div>

          <div className="text-center p-6 rounded-xl border border-[#2a2a2a] bg-[#0d0d0d]">
            <div className="text-2xl mb-3">📥</div>
            <h3 className="text-lg font-semibold text-white mb-2">Pull</h3>
            <p className="text-sm text-gray-400">
              Download models from Walrus and load them into Ollama with a single command.
            </p>
            <code className="block mt-3 text-xs text-green-400 bg-[#1a1a2e] rounded p-2">
              wolllama pull O1ABCdef...xyz
            </code>
          </div>
        </div>
      </section>

      {/* Featured Models */}
      <section className="py-20 px-6">
        <h2 className="text-3xl font-bold text-white text-center mb-4">Featured Models</h2>
        <p className="text-center text-gray-400 mb-12">
          Discover models published by the community
        </p>
        <div className="max-w-4xl mx-auto text-center">
          <Link
            to="/models"
            className="inline-block bg-white text-black px-6 py-3 rounded-lg font-medium hover:bg-gray-200 transition-colors"
          >
            Browse All Models →
          </Link>
        </div>
      </section>

      {/* CTA */}
      <section className="py-20 px-6 bg-[#111] text-center">
        <h2 className="text-3xl font-bold text-white mb-4">Ready to own your models?</h2>
        <p className="text-gray-400 mb-8 max-w-lg mx-auto">
          Install the CLI and start pushing models to decentralized storage today.
        </p>
        <div className="flex justify-center gap-4 flex-wrap">
          <a
            href="https://github.com/wolllama-org/wolllama/releases"
            className="bg-white text-black px-6 py-3 rounded-lg font-medium hover:bg-gray-200 transition-colors"
          >
            Download CLI
          </a>
          <a
            href="https://github.com/wolllama-org/wolllama"
            className="border border-[#444] text-white px-6 py-3 rounded-lg font-medium hover:border-white transition-colors"
          >
            GitHub →
          </a>
          <a
            href="https://twitter.com/wolllama"
            className="border border-[#444] text-white px-6 py-3 rounded-lg font-medium hover:border-white transition-colors"
          >
            Twitter →
          </a>
        </div>
      </section>
    </div>
  );
}
