import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

const __dirname = dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  plugins: [react()],
  root: __dirname,
  build: {
    outDir: resolve(__dirname, "../dist/webview"),
    emptyOutDir: true,
    cssCodeSplit: false,
    rollupOptions: {
      input: resolve(__dirname, "index.html"),
      output: {
        entryFileNames: "main-panel.js",
        assetFileNames: (info) => {
          const names = info.names ?? [];
          if (names.some((n) => n.endsWith(".css"))) {
            return "main-panel.css";
          }
          return "[name][extname]";
        },
      },
    },
  },
});
