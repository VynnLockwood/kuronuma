// Static conformance check for the shadcn/ui theme:
//  - hex colours live only in the :root token block of app/globals.css; components use Tailwind token utilities;
//  - no gradients anywhere (flat white and grey surfaces only).
import { readFileSync, readdirSync, statSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join, relative } from "node:path";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const problems = [];

// 1. globals.css: hex only inside `:root { ... }`.
const css = readFileSync(join(root, "app", "globals.css"), "utf8").split("\n");
let inRoot = false;
css.forEach((raw, i) => {
  const line = raw.replace(/\/\*.*?\*\//g, "");
  if (/^:root\s*\{/.test(line)) inRoot = true;
  if (!inRoot && /#[0-9a-fA-F]{3,8}\b/.test(line)) problems.push(`app/globals.css:${i + 1}: hex colour outside :root: ${raw.trim()}`);
  if (/(linear|radial|conic)-gradient/.test(line)) problems.push(`app/globals.css:${i + 1}: gradient: ${raw.trim()}`);
  if (inRoot && /^\}/.test(line)) inRoot = false;
});

// 2. Components and pages: no hex colours and no gradients.
function walk(dir) {
  for (const name of readdirSync(dir)) {
    const p = join(dir, name);
    if (statSync(p).isDirectory()) walk(p);
    else if (/\.(tsx|ts)$/.test(name)) {
      readFileSync(p, "utf8")
        .split("\n")
        .forEach((raw, i) => {
          const where = `${relative(root, p)}:${i + 1}`;
          if (!/themeColor/.test(raw) && /["'`\s(\[]#[0-9a-fA-F]{3,8}\b/.test(raw)) problems.push(`${where}: hex colour (use a theme token): ${raw.trim()}`);
          if (/(linear|radial|conic)-gradient|bg-gradient-|bg-linear-|bg-radial-/.test(raw)) problems.push(`${where}: gradient: ${raw.trim()}`);
        });
    }
  }
}
walk(join(root, "app"));
walk(join(root, "components"));

if (problems.length) {
  console.error(problems.join("\n"));
  console.error(`\n${problems.length} design-system violation(s)`);
  process.exit(1);
}
console.log("check:design ok");
