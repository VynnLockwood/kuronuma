// Static conformance check: globals.css must use design tokens only (see docs/superpowers/specs/2026-10-30-web-design-system-design.md).
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

const file = join(dirname(fileURLToPath(import.meta.url)), "..", "app", "globals.css");
const lines = readFileSync(file, "utf8").split("\n");
const problems = [];

// Track the current selector so gradients and 50% radii can be allowed in named rules only.
let selector = "";
const GRADIENT_OK = /\.hero::before|^\.hero-field|\.bx-hero__grid/;
const ROUND_OK = /dot|avatar/;
const RADIUS_OK = /^(0|1px|2px|4px|var\(--radius-[a-z]+\)|inherit)$/;

lines.forEach((raw, i) => {
  const line = raw.replace(/\/\*.*?\*\//g, "");
  const n = i + 1;
  if (/\{\s*$/.test(line) && !/^\s*@/.test(line)) selector = line.replace(/\{\s*$/, "").trim();
  if (/#[0-9a-fA-F]{3,8}\b/.test(line)) problems.push(`${n}: hex colour literal: ${raw.trim()}`);
  const shadow = line.match(/box-shadow\s*:\s*([^;]+);/);
  if (shadow && !/^(none|var\(--shadow-overlay\))$/.test(shadow[1].trim())) problems.push(`${n}: box-shadow: ${raw.trim()}`);
  if (/(linear|radial|conic)-gradient/.test(line) && !GRADIENT_OK.test(selector)) problems.push(`${n}: gradient outside hero grid: ${raw.trim()}`);
  const radius = line.match(/border-radius\s*:\s*([^;]+);/);
  if (radius) {
    const v = radius[1].trim();
    const ok = v.split(/\s+/).every((p) => RADIUS_OK.test(p)) || (v === "50%" && ROUND_OK.test(selector));
    if (!ok) problems.push(`${n}: border-radius ${v}: ${raw.trim()}`);
  }
});

if (problems.length) {
  console.error(problems.join("\n"));
  console.error(`\n${problems.length} design-system violation(s) in app/globals.css`);
  process.exit(1);
}
console.log("check:design ok");
