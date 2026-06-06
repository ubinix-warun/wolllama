import { Link } from "react-router-dom";
import { motion, useScroll, useTransform } from "framer-motion";
import { useRef } from "react";
import wallyImg from "../assets/wally.webp";
import tatumLogo from "../assets/tatum-logo.svg";

// ── Stars ──────────────────────────────────────────────────────────────
const stars = [
  { left: "9%", top: "64%", w: 3, blur: 2, dur: 5, del: 0 },
  { left: "13%", top: "78%", w: 5, blur: 7, dur: 7, del: 2.4 },
  { left: "15%", top: "40%", w: 4, blur: 4, dur: 4.5, del: 1.1 },
  { left: "27%", top: "56%", w: 7, blur: 3, dur: 6, del: 3.6 },
  { left: "5%", top: "30%", w: 2, blur: 1, dur: 4, del: 0.5 },
  { left: "22%", top: "72%", w: 4, blur: 5, dur: 6.5, del: 4.8 },
  { left: "18%", top: "22%", w: 3, blur: 3, dur: 5.5, del: 2 },
  { left: "8%", top: "50%", w: 6, blur: 5, dur: 7, del: 1.5 },
  { left: "30%", top: "38%", w: 2, blur: 2, dur: 4.5, del: 3.2 },
  { left: "12%", top: "90%", w: 3, blur: 4, dur: 6, del: 0.3 },
  { left: "68%", top: "34%", w: 5, blur: 7, dur: 5.5, del: 0.8 },
  { left: "72%", top: "47%", w: 3, blur: 2, dur: 4, del: 4.2 },
  { left: "71%", top: "82%", w: 11, blur: 3, dur: 6.5, del: 1.8 },
  { left: "89%", top: "60%", w: 7, blur: 6, dur: 7.5, del: 3 },
  { left: "95%", top: "35%", w: 2, blur: 1, dur: 4.5, del: 1.2 },
  { left: "78%", top: "24%", w: 4, blur: 4, dur: 5, del: 3.8 },
  { left: "83%", top: "70%", w: 3, blur: 3, dur: 6, del: 0.6 },
  { left: "65%", top: "55%", w: 5, blur: 6, dur: 7, del: 2.8 },
  { left: "92%", top: "80%", w: 6, blur: 5, dur: 5.5, del: 4.5 },
  { left: "75%", top: "15%", w: 2, blur: 2, dur: 4, del: 1.8 },
];

// ── Headline split into characters for blur-reveal ─────────────────────
const headline = "Your models.\nDecentralized storage.\nNo limits.";
const headlineChars = headline.split("").map((char, i) => ({ char, key: i }));

// ── Terminal lines ─────────────────────────────────────────────────────
const terminalLines = [
  { text: "$ wolllama pull PPofW8S3LyWmU-EM..." },
  { text: "Fetching wolllama manifest PPofW8S3LyWmU-EM...... ✓" },
  { text: "[1/5] 62fbfd9ed093 (-zPlDPhJn7zhD6iD...)... ✓" },
  { text: "[2/5] ca7a9654b546 (OUpbCFd9dKawdh42...)... ✓" },
  { text: "[3/5] cfc7749b96f6 (q9V_hlrkIEy7ArhC...)... ✓" },
  { text: "[4/5] eb2c714d40d4 (2 chunks)..." },
  { text: "    chunk 1/2 Og_eZxoUW9FkMy20...... ✓ (47 MB)" },
  { text: "    chunk 2/2 xVcx6DGQA6aVw-iK...... ✓ (45 MB)" },
  { text: "✓ checksum verified" },
  { text: "[5/5] f590523c855b (c69cYycTMrrCHddK...)... ✓" },
  { text: "  ✓ smollm:135m pulled from Walrus" },
  { text: "  Model:  smollm:135m" },
  { text: "  Blobs:  5" },
  { text: "  Size:   92 MB" },
  { text: "  Restart Ollama to use the model:" },
  { text: "    ollama serve" },
  { text: "  Then run: ollama run smollm:135m" },
];

// ── Shared animation variants ──────────────────────────────────────────
const fadeUp = {
  hidden: { opacity: 0, y: 30 },
  visible: { opacity: 1, y: 0 },
};

const stagger = {
  hidden: {},
  visible: { transition: { staggerChildren: 0.08, delayChildren: 0.5 } },
};

const blurReveal = {
  hidden: { opacity: 0, filter: "blur(20px)", y: 20 },
  visible: { opacity: 1, filter: "blur(0px)", y: 0 },
};

const terminalStagger = {
  hidden: {},
  visible: { transition: { staggerChildren: 0.04, delayChildren: 1.2 } },
};

const terminalLine = {
  hidden: { opacity: 0, x: -4 },
  visible: { opacity: 1, x: 0 },
};

const cardStagger = {
  hidden: {},
  visible: { transition: { staggerChildren: 0.15 } },
};

// ── Section reveal wrapper ─────────────────────────────────────────────
function SectionReveal({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <motion.section
      className={className}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true, margin: "-80px" }}
      variants={fadeUp}
      transition={{ duration: 0.6, ease: "easeOut" }}
    >
      {children}
    </motion.section>
  );
}

// ── Component ──────────────────────────────────────────────────────────
export default function LandingPage() {
  const heroRef = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({
    target: heroRef,
    offset: ["start start", "end start"],
  });
  const wallyY = useTransform(scrollYProgress, [0, 1], [0, 120]);

  return (
    <div>
      {/* ═══ Hero Section ═══ */}
      <section ref={heroRef} className="relative overflow-hidden pt-24 pb-16 md:pt-32 md:pb-24">
        {/* Aurora gradient background */}
        <div className="absolute inset-0 bg-gradient-to-b from-[#0a0a2e] via-[#0d1225] via-40% to-[#0d0d0d]">
          <div className="absolute inset-0 animate-aurora bg-[radial-gradient(ellipse_80%_60%_at_50%_0%,rgba(66,99,235,0.15),transparent_70%),radial-gradient(ellipse_60%_50%_at_80%_20%,rgba(99,50,200,0.1),transparent_70%),radial-gradient(ellipse_50%_50%_at_20%_10%,rgba(30,100,200,0.1),transparent_70%)]" />
        </div>

        {/* Twinkling stars */}
        <div className="absolute inset-0 overflow-hidden pointer-events-none" aria-hidden="true">
          {stars.map((s, i) => (
            <div
              key={i}
              className="animate-twinkle absolute rounded-full bg-white"
              style={{
                left: s.left,
                top: s.top,
                width: s.w,
                height: s.w,
                filter: `blur(${s.blur}px)`,
                ["--twinkle-dur" as never]: `${s.dur}s`,
                ["--twinkle-del" as never]: `${s.del}s`,
              } as React.CSSProperties}
            />
          ))}
        </div>

        {/* Hero content */}
        <div className="relative z-10 flex flex-col items-center px-6">
          {/* Character-level blur-reveal headline */}
          <motion.h1
            className="text-center text-4xl md:text-6xl lg:text-7xl font-bold text-white mb-6 leading-[0.9] tracking-tight max-w-4xl"
            initial="hidden"
            animate="visible"
            variants={stagger}
          >
            {headlineChars.map(({ char, key }) =>
              char === "\n" ? (
                <br key={key} />
              ) : (
                <motion.span
                  key={key}
                  variants={blurReveal}
                  transition={{ duration: 0.6, ease: "easeOut" }}
                  className="inline-block"
                  style={char === " " ? { width: "0.3em" } : undefined}
                >
                  {char === " " ? "\u00A0" : char}
                </motion.span>
              )
            )}
          </motion.h1>

          {/* Subtitle fade-up */}
          <motion.p
            className="text-lg md:text-xl text-gray-400 mb-8 max-w-xl md:max-w-2xl text-center leading-relaxed"
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 1.2, ease: "easeOut" }}
          >
            Push and pull Ollama models to decentralized storage on{" "}
            <strong className="text-blue-400 font-semibold">Walrus</strong>.
            No central registry. No gatekeepers. Just you and your models.
          </motion.p>

          {/* CTA pill button */}
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 1.5, ease: "easeOut" }}
          >
            <a
              href="https://github.com/wolllama-org/wolllama/releases"
              className="inline-flex items-center gap-3 rounded-[26px] border border-white/30 px-8 py-3 text-lg font-medium text-white tracking-[0.04em] transition-all duration-200 hover:border-white hover:bg-white/5"
            >
              Start building
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="9"
                height="9"
                viewBox="0 0 9 9"
                fill="none"
                className="shrink-0"
              >
                <path d="M9 0.00427246H0V1.00427H9V0.00427246Z" fill="currentColor" />
                <path d="M8.29674 9.13222e-06L0.00238037 8.29437L0.709487 9.00148L9.00385 0.707116L8.29674 9.13222e-06Z" fill="currentColor" />
                <path d="M9 0.00427246H8V9.00427H9V0.00427246Z" fill="currentColor" />
              </svg>
            </a>
          </motion.div>

          {/* Wally mascot — parallax fade-up */}
          <motion.div
            className="mt-8 md:mt-12"
            initial={{ opacity: 0, y: 80 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 1.0, ease: "easeOut" }}
            style={{ y: wallyY }}
          >
            <img
              src={wallyImg}
              alt="Wally the Walrus mascot"
              className="w-[300px] sm:w-[400px] md:w-[600px] h-auto drop-shadow-[0_0_80px_rgba(66,99,235,0.3)]"
            />
          </motion.div>
        </div>
      </section>

      {/* Terminal mockup */}
      <motion.div
        className="max-w-lg mx-auto px-6 -mt-8 relative z-20"
        initial="hidden"
        animate="visible"
        variants={terminalStagger}
      >
        <div className="bg-[#1a1a2e] rounded-lg overflow-hidden border border-white/10 shadow-2xl">
          <div className="flex items-center gap-2 px-4 py-2 bg-[#060a14] border-b border-white/10">
            <span className="w-3 h-3 rounded-full bg-red-500" />
            <span className="w-3 h-3 rounded-full bg-yellow-500" />
            <span className="w-3 h-3 rounded-full bg-green-500" />
            <span className="text-xs text-gray-500 ml-2">terminal</span>
          </div>
          <div className="p-4 text-left font-mono text-xs leading-relaxed">
            {terminalLines.map((line, i) => {
              const isGreen =
                line.text.includes("✓") || line.text.startsWith("$");
              const isDim = line.text.startsWith("    ") && !line.text.includes("✓");
              return (
                <motion.div
                  key={i}
                  className={`mb-1 ${
                    isGreen ? "text-green-400" : isDim ? "text-gray-600" : "text-gray-500"
                  }`}
                  variants={terminalLine}
                >
                  {line.text}
                </motion.div>
              );
            })}
          </div>
        </div>
      </motion.div>

      {/* ═══ How It Works ═══ */}
      <SectionReveal className="py-20 px-6 bg-[#0a0e1a]">
        <h2 className="text-3xl font-bold text-white text-center mb-12">How It Works</h2>
        <motion.div
          className="max-w-4xl mx-auto grid grid-cols-1 md:grid-cols-3 gap-8"
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true, margin: "-60px" }}
          variants={cardStagger}
        >
          {[
            {
              emoji: "📤",
              title: "Push",
              desc: "Upload models directly from your Ollama store to Walrus decentralized storage.",
              cmd: "wolllama push llama3.2:latest",
              cmdColor: "text-green-400",
            },
            {
              emoji: "🔗",
              title: "Share",
              desc: "Get a manifest object ID. Share it with anyone, or submit to the registry for discovery.",
              cmd: "O1ABCdef...xyz",
              cmdColor: "text-blue-400",
            },
            {
              emoji: "📥",
              title: "Pull",
              desc: "Download models from Walrus and load them into Ollama with a single command.",
              cmd: "wolllama pull O1ABCdef...xyz",
              cmdColor: "text-green-400",
            },
          ].map((card) => (
            <motion.div
              key={card.title}
              className="text-center p-6 rounded-xl border border-white/10 bg-[#060a14]"
              variants={fadeUp}
            >
              <div className="text-2xl mb-3">{card.emoji}</div>
              <h3 className="text-lg font-semibold text-white mb-2">{card.title}</h3>
              <p className="text-sm text-gray-400">{card.desc}</p>
              <code className={`block mt-3 text-xs ${card.cmdColor} bg-[#1a1a2e] rounded p-2`}>
                {card.cmd}
              </code>
            </motion.div>
          ))}
        </motion.div>
      </SectionReveal>

      {/* ═══ Roadmap ═══ */}
      <SectionReveal className="py-20 px-6 bg-[#0a0e1a]">
        <h2 className="text-3xl font-bold text-white text-center mb-4">Roadmap</h2>
        <p className="text-center text-gray-400 mb-12 max-w-xl mx-auto">
          Where we've been and where we're headed.
        </p>

        <div className="max-w-5xl mx-auto">
          {/* Horizontal timeline */}
          <div className="hidden md:flex items-start justify-between gap-2">
            {[
              {
                version: "v1.0",
                title: "Direct Walrus",
                desc: "CLI, API, site, manifest v2, SHA256 verification.",
                status: "done",
              },
              {
                version: "v1.2",
                title: "Multi-Provider",
                desc: "WalrusProvider (256 MB) + TatumProvider (45 MB), provider selection.",
                status: "done",
                logo: tatumLogo,
                logoAlt: "Tatum",
              },
              {
                version: "v1.3",
                title: "On-Chain Registry",
                desc: "Sui wallet login, Move smart contract, on-chain audit trail.",
                status: "in-progress",
              },
              {
                version: "v2.0",
                title: "Private Registry",
                desc: "Sign-in, public/private models, SealSDK encryption.",
                status: "planned",
              },
              {
                version: "v3.0",
                title: "Ecosystem",
                desc: "Search, versioning, analytics, team features.",
                status: "planned",
              },
            ].map((item, idx, arr) => (
              <div key={item.version} className="flex-1 flex flex-col items-center min-w-0">
                {/* Dot + connector bar */}
                <div className="w-full flex items-center">
                  {/* Left connector */}
                  <div
                    className={`flex-1 h-0.5 ${
                      idx === 0
                        ? "bg-transparent"
                        : item.status === "done"
                          ? "bg-green-500/50"
                          : "bg-white/10"
                    }`}
                  />
                  {/* Dot */}
                  <div
                    className={`flex-shrink-0 w-5 h-5 rounded-full border-2 flex items-center justify-center ${
                      item.status === "done"
                        ? "bg-green-500 border-green-400"
                        : item.status === "in-progress"
                          ? "bg-blue-500/30 border-blue-400"
                          : "bg-[#0a0e1a] border-white/20"
                    }`}
                  >
                    {item.status === "done" && (
                      <span className="text-[10px] text-white leading-none">✓</span>
                    )}
                    {item.status === "in-progress" && (
                      <span className="w-1.5 h-1.5 rounded-full bg-blue-400" />
                    )}
                  </div>
                  {/* Right connector */}
                  <div
                    className={`flex-1 h-0.5 ${
                      idx === arr.length - 1
                        ? "bg-transparent"
                        : item.status === "done"
                          ? "bg-green-500/50"
                          : "bg-white/10"
                    }`}
                  />
                </div>

                {/* Card below dot */}
                <div
                  className={`mt-4 w-full rounded-xl border p-4 text-center ${
                    item.status === "in-progress"
                      ? "border-blue-400/30 bg-blue-500/5"
                      : "border-white/10 bg-[#060a14]"
                  }`}
                >
                  <div className="flex items-center justify-center gap-1.5 mb-1">
                    <span
                      className={`text-[10px] font-bold px-1.5 py-0.5 rounded ${
                        item.status === "done"
                          ? "bg-green-500/20 text-green-400"
                          : item.status === "in-progress"
                            ? "bg-blue-500/20 text-blue-400"
                            : "bg-white/10 text-gray-400"
                      }`}
                    >
                      {item.version}
                    </span>
                    {"logo" in item && item.logo && (
                      <img
                        src={item.logo as string}
                        alt={(item as { logoAlt?: string }).logoAlt || ""}
                        className="h-4 opacity-60"
                      />
                    )}
                  </div>
                  <h3 className="text-sm font-semibold text-white mb-1">{item.title}</h3>
                  <p className="text-xs text-gray-400 leading-relaxed">{item.desc}</p>
                </div>
              </div>
            ))}
          </div>

          {/* Mobile: vertical stack */}
          <div className="md:hidden space-y-4">
            {[
              {
                version: "v1.0",
                title: "Direct Walrus",
                desc: "CLI, API, site, manifest v2, SHA256 verification. GoReleaser CI.",
                status: "done",
              },
              {
                version: "v1.2",
                title: "Multi-Provider Storage",
                desc: "WalrusProvider (256 MB) and TatumProvider (45 MB). Provider selection and quiltPatchId support.",
                status: "done",
                logo: tatumLogo,
                logoAlt: "Tatum",
              },
              {
                version: "v1.3",
                title: "Sui On-Chain Registry",
                desc: "Sui wallet login via zkLogin, Move smart contract registry, on-chain audit trail, decentralized identity.",
                status: "in-progress",
              },
              {
                version: "v2.0",
                title: "Private Registry",
                desc: "Authenticated sign-in, public/private model visibility, SealSDK encryption, verified publishing.",
                status: "planned",
              },
              {
                version: "v3.0",
                title: "Ecosystem",
                desc: "Full-text search, model versioning, download analytics, team collaboration features.",
                status: "planned",
              },
            ].map((item) => (
              <div
                key={item.version}
                className={`bg-[#060a14] border rounded-xl p-4 ${
                  item.status === "in-progress"
                    ? "border-blue-400/30"
                    : "border-white/10"
                }`}
              >
                <div className="flex items-center gap-2 mb-2">
                  <div
                    className={`w-3 h-3 rounded-full border-2 flex-shrink-0 ${
                      item.status === "done"
                        ? "bg-green-500 border-green-400"
                        : item.status === "in-progress"
                          ? "bg-blue-500/30 border-blue-400"
                          : "bg-[#0a0e1a] border-white/20"
                    }`}
                  />
                  <span
                    className={`text-xs font-bold px-1.5 py-0.5 rounded ${
                      item.status === "done"
                        ? "bg-green-500/20 text-green-400"
                        : item.status === "in-progress"
                          ? "bg-blue-500/20 text-blue-400"
                          : "bg-white/10 text-gray-400"
                    }`}
                  >
                    {item.version}
                  </span>
                  <h3 className="text-sm font-semibold text-white">{item.title}</h3>
                  {"logo" in item && item.logo && (
                    <img
                      src={item.logo as string}
                      alt={(item as { logoAlt?: string }).logoAlt || ""}
                      className="h-4 opacity-60"
                    />
                  )}
                </div>
                <p className="text-xs text-gray-400 leading-relaxed">{item.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </SectionReveal>

      {/* ═══ Featured Models ═══ */}
      <SectionReveal className="py-20 px-6">
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
      </SectionReveal>

      {/* ═══ CTA ═══ */}
      <SectionReveal className="py-20 px-6 bg-[#0a0e1a] text-center">
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
      </SectionReveal>

    </div>
  );
}
