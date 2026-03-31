"use strict";

const fs = require("node:fs");
const path = require("node:path");

const pkgRoot = path.dirname(require.resolve("@vocode/ui/package.json"));
const src = path.join(pkgRoot, "assets", "vocode_icon_black.svg");
const dest = path.join(__dirname, "..", "media", "vocode-icon.svg");

fs.mkdirSync(path.dirname(dest), { recursive: true });
fs.copyFileSync(src, dest);
