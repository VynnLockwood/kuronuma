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

// 3. Blush palette: pinned values and WCAG contrast. The tokens are the `--name: #hex;` lines of :root.
const rootBlock = readFileSync(join(root, "app", "globals.css"), "utf8").match(/:root\s*\{([\s\S]*?)\n\}/)?.[1] ?? "";
const tokens = Object.fromEntries([...rootBlock.matchAll(/--([a-z0-9-]+):\s*(#[0-9a-fA-F]{6})\s*;/g)].map((m) => [m[1], m[2].toLowerCase()]));
const luminance = (hex) => {
  const [r, g, b] = [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16) / 255).map((c) => (c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4));
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
};
const contrast = (a, b) => {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (hi + 0.05) / (lo + 0.05);
};
for (const [name, value] of Object.entries({ background: "#fffafb", primary: "#c23b69" })) {
  if (tokens[name] !== value) problems.push(`app/globals.css: --${name} must be ${value} (Blush), found ${tokens[name] ?? "nothing"}`);
}
const pairs = [
  ["foreground", ["background", "muted", "card"], 7],
  ["muted-foreground", ["background", "muted"], 4.5],
  ["ink-2", ["background", "muted"], 4.5],
  ["link", ["background", "muted"], 4.5],
  ["link-hover", ["background", "muted"], 4.5],
  ["primary-foreground", ["primary"], 4.5],
  ["success", ["success-muted"], 4.5],
  ["info", ["info-muted"], 4.5],
  ["destructive", ["destructive-muted"], 4.5],
  ["subtle", ["background", "muted"], 3], // large text only
  ["ring", ["background"], 3],
  ["primary", ["background"], 3],
];
for (const [fg, backs, min] of pairs) {
  for (const bg of backs) {
    if (!tokens[fg] || !tokens[bg]) problems.push(`contrast: token --${fg} or --${bg} is missing`);
    else if (contrast(tokens[fg], tokens[bg]) < min) problems.push(`contrast: --${fg} on --${bg} is ${contrast(tokens[fg], tokens[bg]).toFixed(2)}, needs ${min}`);
  }
}

if (problems.length) {
  console.error(problems.join("\n"));
  console.error(`\n${problems.length} design-system violation(s)`);
  process.exit(1);
}
console.log("check:design ok");
