/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./web/templates/**/*.html", "./web/static/**/*.js"],
  safelist: [
    "animate-glow-pulse",
    "animate-pulse",
    "border-l-2",
    "border-neon-blue",
    "border-neon-green",
    "border-neon-red",
    "border-border-dim",
    "bg-neon-blue/5",
    "bg-neon-red/5",
    "bg-neon-blue",
    "bg-neon-green",
    "bg-neon-red",
    "bg-dim",
  ],
  theme: {
    extend: {
      colors: {
        base: "#08080f",
        surface: "#0f0f1c",
        raised: "#161625",
        "neon-blue": "#00c2ff",
        "neon-violet": "#8b5cf6",
        "neon-green": "#00e887",
        "neon-red": "#ff4466",
        "neon-amber": "#f59e0b",
        dim: "#5a5a7a",
        "border-dim": "#1e1e35",
      },
      boxShadow: {
        "glow-blue": "0 0 14px rgba(0,194,255,0.2)",
        "glow-violet": "0 0 14px rgba(139,92,246,0.2)",
        "glow-green": "0 0 14px rgba(0,232,135,0.2)",
        "glow-red": "0 0 14px rgba(255,68,102,0.2)",
      },
      fontFamily: {
        mono: ["JetBrains Mono", "monospace"],
      },
      animation: {
        "fade-up": "fadeUp 0.2s ease-out",
        "glow-pulse": "glowPulse 2s ease-in-out infinite",
      },
      keyframes: {
        fadeUp: {
          "0%": { opacity: "0", transform: "translateY(6px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
        glowPulse: {
          "0%,100%": { boxShadow: "0 0 4px rgba(0,194,255,0.1)" },
          "50%": { boxShadow: "0 0 16px rgba(0,194,255,0.35)" },
        },
      },
    },
  },
}
