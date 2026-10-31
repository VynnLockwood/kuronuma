"use client";

import { useEffect, useRef, useState } from "react";
import { Letter } from "@/components/motifs";
import { makeLetters, stepLetter, type LetterState, type Pointer } from "@/lib/letter-physics";

// Faint paper letters drift behind the page and ease away from a mouse pointer. Decoration only: the layer is
// aria-hidden, takes no pointer events and sits behind all content. One requestAnimationFrame loop writes only
// `transform`; it stops while the tab is hidden, and under reduced motion the letters are drawn once and never move.
const place = (l: LetterState) => `translate3d(${l.x.toFixed(2)}px, ${l.y.toFixed(2)}px, 0) rotate(${l.rot.toFixed(2)}deg)`;

export function LetterField() {
  const [letters, setLetters] = useState<LetterState[]>([]);
  const els = useRef<(HTMLDivElement | null)[]>([]);

  // Created after mount (the window size is needed), so the server renders only the empty layer.
  useEffect(() => {
    setLetters(makeLetters(window.innerWidth < 640 ? 8 : 14, window.innerWidth, window.innerHeight, 7));
  }, []);

  useEffect(() => {
    if (letters.length === 0) return;
    const state = letters.slice();
    let width = window.innerWidth;
    let height = window.innerHeight;
    let pointer: Pointer = null;
    const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
    let reduced = mq.matches;
    let raf = 0;
    let last = 0;
    let running = false;

    const tick = (now: number) => {
      const dt = last ? (now - last) / 1000 : 0;
      last = now;
      for (let i = 0; i < state.length; i++) {
        state[i] = stepLetter(state[i], dt, pointer, width, height);
        const el = els.current[i];
        if (el) el.style.transform = place(state[i]);
      }
      raf = requestAnimationFrame(tick);
    };
    const start = () => {
      if (running || reduced || document.hidden) return;
      running = true;
      last = 0;
      raf = requestAnimationFrame(tick);
    };
    const stop = () => {
      running = false;
      cancelAnimationFrame(raf);
    };

    const onMove = (e: PointerEvent) => {
      if (reduced || e.pointerType === "touch") return;
      pointer = { x: e.clientX, y: e.clientY };
    };
    const clearPointer = () => {
      pointer = null;
    };
    const onVisibility = () => (document.hidden ? stop() : start());
    const onResize = () => {
      width = window.innerWidth;
      height = window.innerHeight;
    };
    const onMotion = (e: MediaQueryListEvent) => {
      reduced = e.matches;
      if (reduced) {
        stop();
        clearPointer();
      } else start();
    };

    window.addEventListener("pointermove", onMove, { passive: true });
    document.documentElement.addEventListener("pointerleave", clearPointer);
    window.addEventListener("blur", clearPointer);
    window.addEventListener("resize", onResize);
    document.addEventListener("visibilitychange", onVisibility);
    mq.addEventListener("change", onMotion);
    start();
    return () => {
      stop();
      window.removeEventListener("pointermove", onMove);
      document.documentElement.removeEventListener("pointerleave", clearPointer);
      window.removeEventListener("blur", clearPointer);
      window.removeEventListener("resize", onResize);
      document.removeEventListener("visibilitychange", onVisibility);
      mq.removeEventListener("change", onMotion);
    };
  }, [letters]);

  return (
    <div aria-hidden="true" data-letter-field className="pointer-events-none fixed inset-0 -z-10 overflow-hidden">
      {letters.map((l, i) => (
        <div
          key={i}
          data-letter
          ref={(el) => {
            els.current[i] = el;
          }}
          className="absolute top-0 left-0 will-change-transform"
          style={{ width: l.size, height: l.size, opacity: l.opacity, transform: place(l) }}
        >
          <Letter size={l.size} />
        </div>
      ))}
    </div>
  );
}
