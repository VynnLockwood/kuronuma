import type { CSSProperties, ReactNode } from "react";
import { cn } from "@/lib/utils";

// Six small original SVG motifs. They are decoration only: hidden from assistive technology, never focusable,
// never take a click. Colours come from the --motif-* tokens in globals.css.
export type MotifProps = { size?: number; className?: string; style?: CSSProperties };

function Frame({ name, size = 24, className, style, children }: MotifProps & { name: string; children: ReactNode }) {
  return (
    <svg
      data-motif={name}
      width={size}
      height={size}
      viewBox="0 0 48 48"
      aria-hidden="true"
      focusable="false"
      className={cn("pointer-events-none shrink-0", className)}
      style={style}
    >
      {children}
    </svg>
  );
}

const paint = (fill: string, line: string, width = 1.5): CSSProperties => ({
  fill: `var(--motif-${fill})`,
  stroke: `var(--motif-${line}-line)`,
  strokeWidth: width,
  strokeLinejoin: "round",
  strokeLinecap: "round",
});

export function Sparkle(props: MotifProps) {
  return (
    <Frame name="sparkle" {...props}>
      <path d="M24 4 C25.5 16 32 22.5 44 24 C32 25.5 25.5 32 24 44 C22.5 32 16 25.5 4 24 C16 22.5 22.5 16 24 4Z" style={paint("gold", "gold")} />
    </Frame>
  );
}

export function Sprout(props: MotifProps) {
  return (
    <Frame name="sprout" {...props}>
      <path d="M24 44 V22" style={{ fill: "none", stroke: "var(--motif-leaf-line)", strokeWidth: 3, strokeLinecap: "round" }} />
      <path d="M24 26 C12 26 8 16 10 10 C20 10 25 16 24 26Z" style={paint("leaf", "leaf")} />
      <path d="M24 22 C24 14 30 8 40 8 C41 18 35 24 24 22Z" style={paint("leaf", "leaf")} />
    </Frame>
  );
}

export function Cloud(props: MotifProps) {
  return (
    <Frame name="cloud" {...props}>
      <path d="M13 34 C6 34 5 24 12 23 C12 15 24 12 28 19 C35 15 43 21 39 28 C45 30 43 38 36 36 L14 36Z" style={paint("lilac", "lilac")} />
    </Frame>
  );
}

export function Letter(props: MotifProps) {
  return (
    <Frame name="letter" {...props}>
      <rect x="6" y="12" width="36" height="25" rx="4" style={{ ...paint("card", "rose"), fill: "var(--card)" }} />
      <path d="M7 14 L24 27 L41 14" style={{ ...paint("card", "rose"), fill: "none" }} />
      <path
        d="M24 35 C20 31 19 28 21.5 27 C23 26.5 24 28 24 28 C24 28 25 26.5 26.5 27 C29 28 28 31 24 35Z"
        transform="translate(0 -3)"
        style={paint("rose", "rose", 1)}
      />
    </Frame>
  );
}

export function Petal(props: MotifProps) {
  return (
    <Frame name="petal" {...props}>
      <path d="M24 40 C10 36 6 22 14 12 C22 14 28 20 24 40Z" style={paint("rose", "rose", 2)} />
      <path d="M24 40 C30 28 30 20 24 10" style={{ fill: "none", stroke: "var(--motif-rose-line)", strokeWidth: 1.2, strokeLinecap: "round" }} />
    </Frame>
  );
}

export function Star(props: MotifProps) {
  return (
    <Frame name="star" {...props}>
      <path d="M24 7 L28.5 18 L40.5 19 L31.3 26.8 L34.2 38.5 L24 32.2 L13.8 38.5 L16.7 26.8 L7.5 19 L19.5 18Z" style={paint("lilac", "lilac")} />
    </Frame>
  );
}

// A quiet divider between sections.
export function MotifDivider() {
  return (
    <div aria-hidden="true" className="flex items-center justify-center gap-4 py-2">
      <Sparkle size={18} className="motif-twinkle" style={{ "--delay": "-0.4s" } as CSSProperties} />
      <Sprout size={18} className="motif-float" style={{ "--delay": "-2.2s" } as CSSProperties} />
      <Sparkle size={18} className="motif-twinkle" style={{ "--delay": "-1.6s" } as CSSProperties} />
    </div>
  );
}
