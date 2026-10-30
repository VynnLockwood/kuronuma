"use client";

import { useEffect, useRef } from "react";

const GRID = 64; // matches --space-8, the hero grid pitch

// Cursor-reactive layer for the landing hero: a grid that lights up around the pointer and a crosshair tick that
// snaps to the nearest grid node. The logo illustration is the only picture in the hero.
export function HeroField() {
  const field = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const hero = field.current?.parentElement;
    if (!hero) return;

    const reduce = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    const fine = window.matchMedia("(pointer: fine)").matches;
    if (reduce || !fine) return; // flat grid

    let moveRaf = 0;

    const onMove = (e: PointerEvent) => {
      if (moveRaf) return;
      moveRaf = requestAnimationFrame(() => {
        moveRaf = 0;
        const r = hero.getBoundingClientRect();
        const x = e.clientX - r.left;
        const y = e.clientY - r.top;
        hero.style.setProperty("--mx", `${x}px`);
        hero.style.setProperty("--my", `${y}px`);
        hero.style.setProperty("--tx", `${Math.round(x / GRID) * GRID}px`);
        hero.style.setProperty("--ty", `${Math.round(y / GRID) * GRID}px`);
        hero.style.setProperty("--near", "1");
      });
    };
    const onLeave = () => {
      hero.style.setProperty("--mx", "-9999px");
      hero.style.setProperty("--my", "-9999px");
      hero.style.setProperty("--near", "0");
    };

    hero.addEventListener("pointermove", onMove);
    hero.addEventListener("pointerleave", onLeave);

    return () => {
      cancelAnimationFrame(moveRaf);
      hero.removeEventListener("pointermove", onMove);
      hero.removeEventListener("pointerleave", onLeave);
    };
  }, []);

  return (
    <div className="hero-field" ref={field} aria-hidden="true">
      <div className="field-light" />
      <span className="field-tick" />
    </div>
  );
}
