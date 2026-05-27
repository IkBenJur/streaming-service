import { reactRouter } from "@react-router/dev/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [
    tailwindcss(),
    reactRouter(),
    {
      name: "chrome-devtools-probe",
      configureServer(server) {
        server.middlewares.use("/.well-known/appspecific/com.chrome.devtools.json", (_req, res) => {
          res.setHeader("Content-Type", "application/json");
          res.end("{}");
        });
      },
    },
  ],
  resolve: {
    tsconfigPaths: true,
  },
});
