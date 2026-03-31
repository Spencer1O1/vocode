import { access } from "node:fs/promises";
import { join } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(fileURLToPath(new URL("..", import.meta.url)));
const expected = [
  "assets/vocode_icon_color.svg",
  "assets/vocode_icon_black.svg",
  "assets/vocode_icon_white.svg",
];

for (const rel of expected) {
  await access(join(root, rel));
}
