// Physics for the drifting background letters (components/LetterField.tsx). Pure functions, no imports, so they can
// be checked in Node without a browser. Positions are the top-left corner in px; angles are degrees.

export type LetterState = {
  x: number;
  y: number;
  vx: number;
  vy: number;
  /** Base drift in px per second: the velocity a letter relaxes back to. */
  bx: number;
  by: number;
  rot: number;
  size: number;
  opacity: number;
};
export type Pointer = { x: number; y: number } | null;

export const RADIUS = 120;
export const MAX_DT = 0.05;
const PUSH = 900; // px/s² at contact; falls off with the square of the distance to RADIUS
const RELAX = 2.5; // per second: how fast velocity returns to the base drift
const TILT = 14; // degrees of tilt away from the pointer at contact
const TILT_RATE = 4; // per second

// mulberry32: small seeded generator, so the same seed always gives the same layout.
function rng(seed: number): () => number {
  let a = seed >>> 0;
  return () => {
    a = (a + 0x6d2b79f5) >>> 0;
    let t = a;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

export function makeLetters(count: number, width: number, height: number, seed: number): LetterState[] {
  const r = rng(seed);
  const out: LetterState[] = [];
  for (let i = 0; i < count; i++) {
    const size = 22 + r() * 18;
    const speed = 6 + r() * 10;
    const angle = (r() - 0.5) * 2 * ((70 * Math.PI) / 180);
    const up = r() < 0.85 ? -1 : 1;
    const bx = speed * Math.sin(angle);
    const by = up * speed * Math.cos(angle);
    out.push({
      x: r() * Math.max(0, width - size),
      y: r() * Math.max(0, height - size),
      vx: bx,
      vy: by,
      bx,
      by,
      rot: (r() - 0.5) * 16,
      size,
      opacity: 0.35 + r() * 0.35,
    });
  }
  return out;
}

export function stepLetter(l: LetterState, dt: number, pointer: Pointer, width: number, height: number): LetterState {
  const h = Math.min(Math.max(dt, 0), MAX_DT);
  let ax = 0;
  let ay = 0;
  let rotTarget = 0;
  if (pointer) {
    const dx = l.x + l.size / 2 - pointer.x;
    const dy = l.y + l.size / 2 - pointer.y;
    const d = Math.hypot(dx, dy);
    if (d < RADIUS) {
      const k = 1 - d / RADIUS;
      const f = PUSH * k * k;
      if (d < 1e-3) ay = -f; // exactly on the centre: push up
      else {
        ax = (f * dx) / d;
        ay = (f * dy) / d;
      }
      rotTarget = (dx < 0 ? -1 : 1) * TILT * k;
    }
  }
  const vx = l.vx + (-RELAX * (l.vx - l.bx) + ax) * h;
  const vy = l.vy + (-RELAX * (l.vy - l.by) + ay) * h;
  let x = l.x + vx * h;
  let y = l.y + vy * h;
  const m = l.size;
  if (x < -m) x += width + 2 * m;
  else if (x > width + m) x -= width + 2 * m;
  if (y < -m) y += height + 2 * m;
  else if (y > height + m) y -= height + 2 * m;
  const rot = l.rot + (rotTarget - l.rot) * Math.min(1, TILT_RATE * h);
  return { ...l, x, y, vx, vy, rot };
}
