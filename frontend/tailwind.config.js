/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{svelte,js,ts,jsx,tsx}"],
  theme: {
    extend: {
      colors: {
        tokyo: {
          bgDark: "#16161e",
          bgBase: "#1a1b26",
          bgSurface: "#24283b",
          bgHighlight: "#292e42",
          bgHover: "#2e344f",

          borderSubtle: "#23283d",
          border: "#2e344f",
          borderFocus: "#4fd6be",

          textMain: "#c0caf5",
          textSub: "#a9b1d6",
          textMuted: "#565f89",
          textDark: "#414868",

          teal: "#4fd6be",
          tealSubtle: "rgba(79, 214, 190, 0.12)",
          cyan: "#7dcfff",
          cyanSubtle: "rgba(125, 207, 255, 0.15)",
          blue: "#7aa2f7",
          blueSubtle: "rgba(122, 162, 247, 0.12)",
          green: "#9ece6a",
          greenSubtle: "rgba(158, 206, 106, 0.12)",
          orange: "#ff9e64",
          orangeSubtle: "rgba(255, 158, 100, 0.12)",
          purple: "#bb9af7",
          purpleSubtle: "rgba(187, 154, 247, 0.12)",
          yellow: "#e0af68",
          yellowSubtle: "rgba(224, 175, 104, 0.15)",
          red: "#f7768e",
          redSubtle: "rgba(247, 118, 142, 0.12)",
        },
      },
      fontFamily: {
        sans: ["Inter", "system-ui", "sans-serif"],
        mono: ["JetBrains mono", "monospace"],
      },
    },
  },
  plugins: [],
};
