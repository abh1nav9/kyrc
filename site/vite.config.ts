import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

// Static SPA. `base` is "./" so the build works whether it's served from a
// domain root (Vercel/Netlify) or a subpath (GitHub Pages project site).
// Tailwind v4 runs as a Vite plugin — no tailwind.config.js or PostCSS needed.
export default defineConfig({
  plugins: [react(), tailwindcss()],
  base: "./",
});
