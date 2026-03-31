import { createRequire } from "node:module";
import { resolve } from "node:path";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

const require = createRequire(import.meta.url);
const uiAssets = resolve(
  require.resolve("@vocode/ui/package.json"),
  "..",
  "assets",
);

// https://vite.dev/config/
export default defineConfig({
  plugins: [tailwindcss(), react()],
  resolve: {
    alias: {
      "@vocode/ui/assets": uiAssets,
    },
  },
});
