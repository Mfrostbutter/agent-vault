import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
  ],
  base: "/",
  build: {
    outDir: "../internal/server/webdist",
    emptyOutDir: true,
  },
  server: {
    proxy: {
      "/v1": "http://localhost:14321",
      "/proxy": "http://localhost:14321",
      "/discover": "http://localhost:14321",
      "/health": "http://localhost:14321",
      "/invite": "http://localhost:14321",
    },
  },
});
