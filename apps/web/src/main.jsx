import vocodeFavicon from "@vocode/ui/assets/vocode_icon_white.svg?url";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";

import App from "./App.jsx";

const favicon = document.querySelector("link[rel='icon']");
if (favicon) {
  favicon.href = vocodeFavicon;
}

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
