"use client";

import { useEffect, useRef } from "react";

// Polygon points of the Nemo dragonfly mark (public/brand/nemo-mark.svg), drawn here as stroke-only line art.
const MARK = [
  "100.25,87.06 103.88,89.44 111.88,143.44 101.62,160.94 96.12,156.81 90,145.56 98,88.44",
  "105.12,78.19 107.62,77.06 133.75,89.19 175.75,112.19 175.75,114.81 162.38,114.94 117.38,103.94",
  "96.75,76.19 88.88,97.56 83.5,103.94 37.5,114.94 26.12,114.81 26.12,112.19 94.25,75.06",
  "195.88,88.56 193.25,89.94 113.38,75.94 109,71.44 172.5,47.06 181.88,58.44 195.88,84.44",
  "4,88.56 25,50.44 29.38,47.06 92.88,71.44 90.5,74.94 10.62,89.94",
  "101.12,38.94 107.75,43.19 110.88,50.56 100.25,67.94 90,47.44",
];

// Seeded start positions (percent of the hero) and sizes (px) so the static layout is stable.
const SPRITES = [
  { x: 8, y: 16, size: 40 },
  { x: 74, y: 9, size: 30 },
  { x: 48, y: 58, size: 34 },
  { x: 90, y: 72, size: 28 },
  { x: 26, y: 86, size: 44 },
];

const GRID = 64; // matches --space-8, the hero grid pitch
const REPEL_RADIUS = 140;

type Bug = { ox: number; oy: number; vx: number; vy: number; angle: number };

// Cursor-reactive layer for the landing hero: a grid that lights up around the pointer, a crosshair tick that
// snaps to the nearest grid node, and a few dragonfly sprites that wander and veer away from the pointer.
export function HeroField() {
  const field = useRef<HTMLDivElement>(null);
  const sprites = useRef<(SVGSVGElement | null)[]>([]);

  useEffect(() => {
    const el = field.current;
    const hero = el?.parentElement;
    if (!el || !hero) return;

    const reduce = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    const fine = window.matchMedia("(pointer: fine)").matches;
    el.dataset.active = "false";
    if (reduce || !fine) return; // static sprites, flat grid

    const pointer = { x: -9999, y: -9999, near: false };
    const bugs: Bug[] = SPRITES.map(() => ({ ox: 0, oy: 0, vx: (Math.random() - 0.5) * 0.8, vy: (Math.random() - 0.5) * 0.8, angle: Math.random() * 6.28 }));
    let raf = 0;
    let moveRaf = 0;
    let visible = true;
    let running = false;

    const place = (i: number) => {
      const b = bugs[i];
      const node = sprites.current[i];
      if (!node) return;
      const heading = Math.atan2(b.vy, b.vx) * (180 / Math.PI) + 90;
      node.style.transform = `translate(${b.ox}px, ${b.oy}px) rotate(${heading}deg)`;
    };

    const step = () => {
      const rect = hero.getBoundingClientRect();
      const w = rect.width;
      const h = rect.height;
      bugs.forEach((b, i) => {
        const s = SPRITES[i];
        const baseX = (s.x / 100) * w;
        const baseY = (s.y / 100) * h;
        // wander
        b.angle += (Math.random() - 0.5) * 0.3;
        b.vx += Math.cos(b.angle) * 0.03;
        b.vy += Math.sin(b.angle) * 0.03;
        // repel from the pointer
        const dx = baseX + b.ox + s.size / 2 - pointer.x;
        const dy = baseY + b.oy + s.size / 2 - pointer.y;
        const d = Math.hypot(dx, dy);
        if (pointer.near && d < REPEL_RADIUS) {
          const push = (1 - d / REPEL_RADIUS) * 0.7;
          const nx = d > 0.001 ? dx / d : Math.cos(b.angle);
          const ny = d > 0.001 ? dy / d : Math.sin(b.angle);
          b.vx += nx * push;
          b.vy += ny * push;
        }
        // stay inside the hero
        const ax = baseX + b.ox;
        const ay = baseY + b.oy;
        if (ax < 0) b.vx += 0.08;
        if (ax > w - s.size) b.vx -= 0.08;
        if (ay < 0) b.vy += 0.08;
        if (ay > h - s.size) b.vy -= 0.08;
        // limit speed and damp
        const sp = Math.hypot(b.vx, b.vy);
        const max = pointer.near && d < REPEL_RADIUS ? 3.2 : 0.9;
        if (sp > max) {
          b.vx = (b.vx / sp) * max;
          b.vy = (b.vy / sp) * max;
        }
        b.vx *= 0.985;
        b.vy *= 0.985;
        b.ox += b.vx;
        b.oy += b.vy;
        place(i);
      });
      raf = requestAnimationFrame(step);
    };

    const sync = () => {
      const should = visible && !document.hidden;
      el.dataset.active = should ? "true" : "false";
      if (should && !running) {
        running = true;
        raf = requestAnimationFrame(step);
      } else if (!should && running) {
        running = false;
        cancelAnimationFrame(raf);
      }
    };

    const onMove = (e: PointerEvent) => {
      if (moveRaf) return;
      moveRaf = requestAnimationFrame(() => {
        moveRaf = 0;
        const r = hero.getBoundingClientRect();
        const x = e.clientX - r.left;
        const y = e.clientY - r.top;
        pointer.x = x;
        pointer.y = y;
        pointer.near = true;
        hero.style.setProperty("--mx", `${x}px`);
        hero.style.setProperty("--my", `${y}px`);
        hero.style.setProperty("--tx", `${Math.round(x / GRID) * GRID}px`);
        hero.style.setProperty("--ty", `${Math.round(y / GRID) * GRID}px`);
        hero.style.setProperty("--near", "1");
      });
    };
    const onLeave = () => {
      pointer.near = false;
      hero.style.setProperty("--mx", "-9999px");
      hero.style.setProperty("--my", "-9999px");
      hero.style.setProperty("--near", "0");
    };

    const io = new IntersectionObserver(([entry]) => {
      visible = entry.isIntersecting;
      sync();
    });
    io.observe(hero);
    document.addEventListener("visibilitychange", sync);
    hero.addEventListener("pointermove", onMove);
    hero.addEventListener("pointerleave", onLeave);
    sync();

    return () => {
      cancelAnimationFrame(raf);
      cancelAnimationFrame(moveRaf);
      io.disconnect();
      document.removeEventListener("visibilitychange", sync);
      hero.removeEventListener("pointermove", onMove);
      hero.removeEventListener("pointerleave", onLeave);
    };
  }, []);

  return (
    <div className="hero-field" ref={field} data-active="false" aria-hidden="true">
      <div className="field-light" />
      <span className="field-tick" />
      {SPRITES.map((s, i) => (
        <svg
          key={i}
          ref={(n) => {
            sprites.current[i] = n;
          }}
          className="sprite"
          viewBox="0 0 199.88 199.88"
          width={s.size}
          height={s.size}
          style={{ left: `${s.x}%`, top: `${s.y}%` }}
          focusable="false"
        >
          {MARK.map((pts) => (
            <polygon key={pts} points={pts} />
          ))}
        </svg>
      ))}
    </div>
  );
}
