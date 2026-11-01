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

// ---- Sixteen more, drawn on the same 48 by 48 grid with the same flat fill and round 1.5 outline. ----

// A sitting teddy bear with a bow tie.
export function Teddy(props: MotifProps) {
  return (
    <Frame name="teddy" {...props}>
      <g transform="translate(0 -1.2)">
      <circle cx="13.5" cy="9.5" r="5.5" fill="var(--motif-honey)" stroke="var(--motif-honey-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" /><circle cx="34.5" cy="9.5" r="5.5" fill="var(--motif-honey)" stroke="var(--motif-honey-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <circle cx="13.5" cy="10" r="2.6" fill="var(--motif-rose)" stroke="none"/><circle cx="34.5" cy="10" r="2.6" fill="var(--motif-rose)" stroke="none"/>
      <ellipse cx="24" cy="37" rx="12" ry="9.5" fill="var(--motif-honey)" stroke="var(--motif-honey-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <ellipse cx="14.5" cy="44.5" rx="5.5" ry="3" fill="var(--motif-honey)" stroke="var(--motif-honey-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" /><ellipse cx="33.5" cy="44.5" rx="5.5" ry="3" fill="var(--motif-honey)" stroke="var(--motif-honey-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <ellipse cx="11" cy="34" rx="4" ry="6.5" transform="rotate(22 11 34)" fill="var(--motif-honey)" stroke="var(--motif-honey-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" /><ellipse cx="37" cy="34" rx="4" ry="6.5" transform="rotate(-22 37 34)" fill="var(--motif-honey)" stroke="var(--motif-honey-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <ellipse cx="24" cy="39" rx="6.5" ry="5.8" fill="var(--card)" stroke="var(--motif-honey-line)" strokeWidth="1.1" strokeLinejoin="round" strokeLinecap="round" />
      <circle cx="24" cy="19" r="12.5" fill="var(--motif-honey)" stroke="var(--motif-honey-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <ellipse cx="24" cy="23.2" rx="6" ry="4.6" fill="var(--card)" stroke="var(--motif-honey-line)" strokeWidth="1.1" strokeLinejoin="round" strokeLinecap="round" />
      <ellipse cx="24" cy="21.2" rx="2.2" ry="1.5" fill="var(--motif-honey-line)" stroke="none"/>
      <path d="M24 22.7 V24.6 M21.4 25.6 Q24 27.4 26.6 25.6" fill="none" stroke="var(--motif-honey-line)" strokeWidth="1.2" strokeLinecap="round"/>
      <circle cx="19" cy="16.3" r="1.5" fill="var(--motif-honey-line)" stroke="none"/><circle cx="29" cy="16.3" r="1.5" fill="var(--motif-honey-line)" stroke="none"/>
      <circle cx="15" cy="21.5" r="2.1" fill="var(--motif-rose)" stroke="none" opacity=".9"/><circle cx="33" cy="21.5" r="2.1" fill="var(--motif-rose)" stroke="none" opacity=".9"/>
      <path d="M24 31.6 L18.8 28.6 V34.6Z M24 31.6 L29.2 28.6 V34.6Z" fill="var(--motif-rose)" stroke="var(--motif-rose-line)" strokeWidth="1.2" strokeLinejoin="round" strokeLinecap="round" /><circle cx="24" cy="31.6" r="1.7" fill="var(--motif-rose)" stroke="var(--motif-rose-line)" strokeWidth="1.2" strokeLinejoin="round" strokeLinecap="round" /></g>
    </Frame>
  );
}

// A folded paper plane with a dashed flight trail.
export function Plane(props: MotifProps) {
  return (
    <Frame name="plane" {...props}>
      <path d="M3 39 C8 33 12 41 17 36 S25 41 29 38" fill="none" stroke="var(--motif-sky-line)" strokeWidth="1.5" strokeLinecap="round" strokeDasharray="1.5 3.2"/>
      <path d="M5 22 L43 6 L24 29Z" fill="var(--card)" stroke="var(--motif-sky-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" /><path d="M43 6 L24 29 L31 43Z" fill="var(--motif-sky)" stroke="var(--motif-sky-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M24 29 L26.6 36" fill="none" stroke="var(--motif-sky-line)" strokeWidth="1.2" strokeLinecap="round"/>
    </Frame>
  );
}

// A glossy heart.
export function Heart(props: MotifProps) {
  return (
    <Frame name="heart" {...props}>
      <path d="M24 41 C8 29 6 18 13 13 C18 9.5 22 12 24 16 C26 12 30 9.5 35 13 C42 18 40 29 24 41Z" fill="var(--motif-rose)" stroke="var(--motif-rose-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M14 18 Q15.5 14.5 19 14.3" fill="none" stroke="var(--card)" strokeWidth="2" strokeLinecap="round"/>
    </Frame>
  );
}

// A crescent moon with a sparkle.
export function Moon(props: MotifProps) {
  return (
    <Frame name="moon" {...props}>
      <path d="M31 6 A18 18 0 1 0 42 33 A14 14 0 0 1 31 6Z" fill="var(--motif-gold)" stroke="var(--motif-gold-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M38 10 L39.3 13.2 L42.5 14.5 L39.3 15.8 L38 19 L36.7 15.8 L33.5 14.5 L36.7 13.2Z" fill="var(--card)" stroke="var(--motif-gold-line)" strokeWidth="1.1" strokeLinejoin="round" strokeLinecap="round" />
      <circle cx="20" cy="26" r="1.3" fill="var(--motif-gold-line)" stroke="none"/><circle cx="14" cy="32" r="1.8" fill="var(--motif-gold-line)" stroke="none" opacity=".55"/>
    </Frame>
  );
}

// A bunny face.
export function Bunny(props: MotifProps) {
  return (
    <Frame name="bunny" {...props}>
      <ellipse cx="16.5" cy="12" rx="4.2" ry="10" transform="rotate(-8 16.5 12)" fill="var(--card)" stroke="var(--motif-rose-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" /><ellipse cx="31.5" cy="12" rx="4.2" ry="10" transform="rotate(8 31.5 12)" fill="var(--card)" stroke="var(--motif-rose-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <ellipse cx="16.5" cy="13" rx="2" ry="6.5" transform="rotate(-8 16.5 13)" fill="var(--motif-rose)" stroke="none"/><ellipse cx="31.5" cy="13" rx="2" ry="6.5" transform="rotate(8 31.5 13)" fill="var(--motif-rose)" stroke="none"/>
      <ellipse cx="24" cy="31" rx="13.5" ry="11.5" fill="var(--card)" stroke="var(--motif-rose-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <circle cx="19" cy="29.5" r="1.6" fill="var(--foreground)" stroke="none"/><circle cx="29" cy="29.5" r="1.6" fill="var(--foreground)" stroke="none"/>
      <ellipse cx="24" cy="33" rx="1.9" ry="1.3" fill="var(--motif-rose)" stroke="var(--motif-rose-line)" strokeWidth="1" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M24 34.2 V35.8 M21.8 36.6 Q24 38.2 26.2 36.6" fill="none" stroke="var(--motif-rose-line)" strokeWidth="1.2" strokeLinecap="round"/>
      <circle cx="15" cy="34" r="2.6" fill="var(--motif-rose)" stroke="none"/><circle cx="33" cy="34" r="2.6" fill="var(--motif-rose)" stroke="none"/>
    </Frame>
  );
}

// A cat face.
export function Cat(props: MotifProps) {
  return (
    <Frame name="cat" {...props}>
      <path d="M9 22 L10 7 L20 13 Q24 12 28 13 L38 7 L39 22 Q41 37 24 41 Q7 37 9 22Z" fill="var(--motif-lilac)" stroke="var(--motif-lilac-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M12.3 12 L17.4 15.4 L12.6 19Z M35.7 12 L30.6 15.4 L35.4 19Z" fill="var(--motif-rose)" stroke="none"/>
      <ellipse cx="18" cy="25" rx="1.7" ry="2.3" fill="var(--foreground)" stroke="none"/><ellipse cx="30" cy="25" rx="1.7" ry="2.3" fill="var(--foreground)" stroke="none"/>
      <circle cx="18.6" cy="24" r=".6" fill="var(--card)" stroke="none"/><circle cx="30.6" cy="24" r=".6" fill="var(--card)" stroke="none"/>
      <path d="M22.3 29.4 H25.7 L24 31.2Z" fill="var(--motif-rose)" stroke="var(--motif-rose-line)" strokeWidth="1" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M24 31.2 Q22 34 19.8 32.8 M24 31.2 Q26 34 28.2 32.8" fill="none" stroke="var(--motif-lilac-line)" strokeWidth="1.2" strokeLinecap="round"/>
      <path d="M6.5 28 L14.5 29.5 M6.5 33.5 L14.5 32.2 M41.5 28 L33.5 29.5 M41.5 33.5 L33.5 32.2" fill="none" stroke="var(--motif-lilac-line)" strokeWidth="1.1" strokeLinecap="round"/>
      <circle cx="14.2" cy="31" r="2.3" fill="var(--motif-rose)" stroke="none" opacity=".85"/><circle cx="33.8" cy="31" r="2.3" fill="var(--motif-rose)" stroke="none" opacity=".85"/>
    </Frame>
  );
}

// A cupcake with a cherry and sprinkles.
export function Cupcake(props: MotifProps) {
  return (
    <Frame name="cupcake" {...props}>
      <path d="M12 28 H36 L33 43 Q32.8 44.2 31.6 44.2 H16.4 Q15.2 44.2 15 43Z" fill="var(--motif-honey)" stroke="var(--motif-honey-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M19 28.6 L20.4 43.6 M24 28.6 V43.6 M29 28.6 L27.6 43.6" fill="none" stroke="var(--motif-honey-line)" strokeWidth="1" opacity=".7" strokeLinecap="round"/>
      <path d="M10 28 Q8 21 14 19.5 Q14 12.5 22 12.5 Q25 8 29.5 11.5 Q36 11.5 35 18.5 Q40.5 20.5 38 28Z" fill="var(--motif-rose)" stroke="var(--motif-rose-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M15 24.2 Q24 27.4 33 24.2" fill="none" stroke="var(--card)" strokeWidth="1.7" strokeLinecap="round" opacity=".9"/>
      <circle cx="26" cy="8.6" r="3.2" fill="var(--motif-rose)" stroke="var(--motif-rose-line)" strokeWidth="1.2" strokeLinejoin="round" strokeLinecap="round" /><path d="M26 5.6 Q27 2.6 30 2.8" fill="none" stroke="var(--motif-leaf-line)" strokeWidth="1.3" strokeLinecap="round"/>
      <path d="M17 17 l2 -1 M28 16 l2 1.2 M22 20 l2 .2 M32 21.5 l1.6 -1.4" fill="none" stroke="var(--motif-gold-line)" strokeWidth="1.5" strokeLinecap="round"/>
    </Frame>
  );
}

// A strawberry.
export function Strawberry(props: MotifProps) {
  return (
    <Frame name="strawberry" {...props}>
      <path d="M24 43 C10 36 7.5 24 11 18 C14 13 20 14 24 16.5 C28 14 34 13 37 18 C40.5 24 38 36 24 43Z" fill="var(--motif-rose)" stroke="var(--motif-rose-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M24 16.5 V9.5" fill="none" stroke="var(--motif-leaf-line)" strokeWidth="2.4" strokeLinecap="round"/>
      <path d="M24 16.5 L17.5 11.5 L21 15.8 L15.5 15.3 L22 19.2Z M24 16.5 L30.5 11.5 L27 15.8 L32.5 15.3 L26 19.2Z" fill="var(--motif-leaf)" stroke="var(--motif-leaf-line)" strokeWidth="1.2" strokeLinejoin="round" strokeLinecap="round" />
      <g fill="var(--motif-gold)" stroke="var(--motif-gold-line)" strokeWidth=".8"><ellipse cx="17" cy="25" rx=".9" ry="1.4"/><ellipse cx="24" cy="24" rx=".9" ry="1.4"/><ellipse cx="31" cy="25" rx=".9" ry="1.4"/><ellipse cx="20.5" cy="31" rx=".9" ry="1.4"/><ellipse cx="27.5" cy="31" rx=".9" ry="1.4"/><ellipse cx="24" cy="37" rx=".9" ry="1.4"/><ellipse cx="15" cy="30" rx=".9" ry="1.4"/><ellipse cx="33" cy="30" rx=".9" ry="1.4"/></g>
    </Frame>
  );
}

// A balloon on a string.
export function Balloon(props: MotifProps) {
  return (
    <Frame name="balloon" {...props}>
      <path d="M24 4 C34.5 4 38.5 12 36.5 20 C34.5 27.5 28.5 31.5 24 32.5 C19.5 31.5 13.5 27.5 11.5 20 C9.5 12 13.5 4 24 4Z" fill="var(--motif-rose)" stroke="var(--motif-rose-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M16.5 12 Q17.6 8.4 21 7.4" fill="none" stroke="var(--card)" strokeWidth="2.2" strokeLinecap="round"/>
      <path d="M24 32.5 L21.4 36 H26.6Z" fill="var(--motif-rose)" stroke="var(--motif-rose-line)" strokeWidth="1.2" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M24 36 C19.5 39 28.5 41.5 24 45" fill="none" stroke="var(--motif-rose-line)" strokeWidth="1.2" strokeLinecap="round"/>
    </Frame>
  );
}

// A rainbow resting on two clouds.
export function Rainbow(props: MotifProps) {
  return (
    <Frame name="rainbow" {...props}>
      <g transform="translate(24 24) scale(.9) translate(-24 -24)">
      <path d="M5 37 A19 19 0 0 1 43 37" fill="none" stroke="var(--motif-rose-line)" strokeWidth="3.4" strokeLinecap="round"/>
      <path d="M9.6 37 A14.4 14.4 0 0 1 38.4 37" fill="none" stroke="var(--motif-gold)" strokeWidth="3.4" strokeLinecap="round"/>
      <path d="M14.2 37 A9.8 9.8 0 0 1 33.8 37" fill="none" stroke="var(--motif-leaf)" strokeWidth="3.4" strokeLinecap="round"/>
      <path d="M18.8 37 A5.2 5.2 0 0 1 29.2 37" fill="none" stroke="var(--motif-sky)" strokeWidth="3.4" strokeLinecap="round"/>
      <path d="M2.5 43 a3.6 3.6 0 0 1 1.2 -7 a5 5 0 0 1 9.4 -.8 a4 4 0 0 1 3 7.8Z" fill="var(--card)" stroke="var(--motif-sky-line)" strokeWidth="1.3" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M31.5 43 a3.6 3.6 0 0 1 1.2 -7 a5 5 0 0 1 9.4 -.8 a4 4 0 0 1 3 7.8Z" fill="var(--card)" stroke="var(--motif-sky-line)" strokeWidth="1.3" strokeLinejoin="round" strokeLinecap="round" /></g>
    </Frame>
  );
}

// A daisy.
export function Daisy(props: MotifProps) {
  return (
    <Frame name="daisy" {...props}>
      <ellipse cx="24" cy="11.5" rx="4.7" ry="8" transform="rotate(0 24 24)" fill="var(--card)" stroke="var(--motif-rose-line)" strokeWidth="1.3" strokeLinejoin="round" strokeLinecap="round" /><ellipse cx="24" cy="11.5" rx="4.7" ry="8" transform="rotate(45 24 24)" fill="var(--card)" stroke="var(--motif-rose-line)" strokeWidth="1.3" strokeLinejoin="round" strokeLinecap="round" /><ellipse cx="24" cy="11.5" rx="4.7" ry="8" transform="rotate(90 24 24)" fill="var(--card)" stroke="var(--motif-rose-line)" strokeWidth="1.3" strokeLinejoin="round" strokeLinecap="round" /><ellipse cx="24" cy="11.5" rx="4.7" ry="8" transform="rotate(135 24 24)" fill="var(--card)" stroke="var(--motif-rose-line)" strokeWidth="1.3" strokeLinejoin="round" strokeLinecap="round" /><ellipse cx="24" cy="11.5" rx="4.7" ry="8" transform="rotate(180 24 24)" fill="var(--card)" stroke="var(--motif-rose-line)" strokeWidth="1.3" strokeLinejoin="round" strokeLinecap="round" /><ellipse cx="24" cy="11.5" rx="4.7" ry="8" transform="rotate(225 24 24)" fill="var(--card)" stroke="var(--motif-rose-line)" strokeWidth="1.3" strokeLinejoin="round" strokeLinecap="round" /><ellipse cx="24" cy="11.5" rx="4.7" ry="8" transform="rotate(270 24 24)" fill="var(--card)" stroke="var(--motif-rose-line)" strokeWidth="1.3" strokeLinejoin="round" strokeLinecap="round" /><ellipse cx="24" cy="11.5" rx="4.7" ry="8" transform="rotate(315 24 24)" fill="var(--card)" stroke="var(--motif-rose-line)" strokeWidth="1.3" strokeLinejoin="round" strokeLinecap="round" />
      <circle cx="24" cy="24" r="6.2" fill="var(--motif-gold)" stroke="var(--motif-gold-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" /><circle cx="22" cy="22" r="1.4" fill="var(--card)" stroke="none" opacity=".8"/>
    </Frame>
  );
}

// A chubby songbird.
export function Bird(props: MotifProps) {
  return (
    <Frame name="bird" {...props}>
      <path d="M11 31 L3 30 L5 36.5Z" fill="var(--motif-sky)" stroke="var(--motif-sky-line)" strokeWidth="1.3" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M10 30 C8 18 18 12 24 14 C26 9 34 8 38 13 C40 16 39 19 37 21 C39 31 32 40 22 40 C16 40 11 36 10 30Z" fill="var(--motif-sky)" stroke="var(--motif-sky-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M14 28 Q20 21.5 27.5 27.5 Q22 36 14 28Z" fill="var(--card)" stroke="var(--motif-sky-line)" strokeWidth="1.3" strokeLinejoin="round" strokeLinecap="round" />
      <circle cx="33.6" cy="15" r="1.5" fill="var(--foreground)" stroke="none"/><path d="M38 14 L44 16.6 L38 19.2Z" fill="var(--motif-gold)" stroke="var(--motif-gold-line)" strokeWidth="1.2" strokeLinejoin="round" strokeLinecap="round" />
      <circle cx="35" cy="19.4" r="2" fill="var(--motif-rose)" stroke="none" opacity=".9"/>
      <path d="M19 40 V44.5 M25 40 V44.5" fill="none" stroke="var(--motif-gold-line)" strokeWidth="1.5" strokeLinecap="round"/>
    </Frame>
  );
}

// A steaming teacup with a heart.
export function Teacup(props: MotifProps) {
  return (
    <Frame name="teacup" {...props}>
      <path d="M19 6 Q16 9 19 12 Q22 15 19 18 M27 5 Q24 8 27 11 Q30 14 27 17" fill="none" stroke="var(--motif-sky-line)" strokeWidth="1.4" strokeLinecap="round" opacity=".7"/>
      <ellipse cx="24" cy="40" rx="16.5" ry="3.4" fill="var(--card)" stroke="var(--motif-lilac-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M36 25 C44 24 44 34.5 35 33" fill="none" stroke="var(--motif-lilac-line)" strokeWidth="2.4" strokeLinecap="round"/>
      <path d="M10 21 H37 V28 C37 35.5 31.5 38.5 23.5 38.5 C15.5 38.5 10 35.5 10 28Z" fill="var(--card)" stroke="var(--motif-lilac-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M10.8 25 H36.2" fill="none" stroke="var(--motif-lilac)" strokeWidth="3.4"/>
      <ellipse cx="23.5" cy="21" rx="13.5" ry="2.4" fill="var(--motif-honey)" stroke="var(--motif-honey-line)" strokeWidth="1.2" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M23.5 35 C18 31.5 19 29 21 29 C22.4 29 23.2 30 23.5 30.8 C23.8 30 24.6 29 26 29 C28 29 29 31.5 23.5 35Z" fill="var(--motif-rose)" stroke="var(--motif-rose-line)" strokeWidth="1" strokeLinejoin="round" strokeLinecap="round" />
    </Frame>
  );
}

// A smiling mushroom.
export function Mushroom(props: MotifProps) {
  return (
    <Frame name="mushroom" {...props}>
      <path d="M17.5 28 Q16.5 40 19.5 43.5 H28.5 Q31.5 40 30.5 28Z" fill="var(--card)" stroke="var(--motif-rose-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M4.5 28.5 C4.5 14.5 13.5 6 24 6 C34.5 6 43.5 14.5 43.5 28.5 Q24 34 4.5 28.5Z" fill="var(--motif-rose)" stroke="var(--motif-rose-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <circle cx="15" cy="19" r="3.1" fill="var(--card)" stroke="none"/><circle cx="26.5" cy="13.5" r="2.4" fill="var(--card)" stroke="none"/><circle cx="34" cy="22" r="3.3" fill="var(--card)" stroke="none"/><circle cx="22" cy="25" r="1.9" fill="var(--card)" stroke="none"/>
      <circle cx="21.2" cy="35" r="1.2" fill="var(--foreground)" stroke="none"/><circle cx="26.8" cy="35" r="1.2" fill="var(--foreground)" stroke="none"/>
      <path d="M22.6 38 Q24 39.4 25.4 38" fill="none" stroke="var(--motif-rose-line)" strokeWidth="1.1" strokeLinecap="round"/>
      <circle cx="19.3" cy="37" r="1.4" fill="var(--motif-rose)" stroke="none"/><circle cx="28.7" cy="37" r="1.4" fill="var(--motif-rose)" stroke="none"/>
    </Frame>
  );
}

// A ribbon bow.
export function Bow(props: MotifProps) {
  return (
    <Frame name="bow" {...props}>
      <path d="M23 27 L14.5 42.5 L21 40.2 L25.5 28Z M25 27 L33.5 42.5 L27 40.2 L22.5 28Z" fill="var(--motif-rose)" stroke="var(--motif-rose-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M24 24 C16 11.5 3.5 13.5 5.5 24 C3.5 34.5 16 36.5 24 24Z" fill="var(--motif-rose)" stroke="var(--motif-rose-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M24 24 C32 11.5 44.5 13.5 42.5 24 C44.5 34.5 32 36.5 24 24Z" fill="var(--motif-rose)" stroke="var(--motif-rose-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M9 22.5 Q14 20.5 19 23 M39 22.5 Q34 20.5 29 23" fill="none" stroke="var(--motif-rose-line)" strokeWidth=".9" strokeLinecap="round" opacity=".8"/>
      <ellipse cx="24" cy="24.2" rx="3.7" ry="4.5" fill="var(--motif-rose)" stroke="var(--motif-rose-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
    </Frame>
  );
}

// A pencil.
export function Pencil(props: MotifProps) {
  return (
    <Frame name="pencil" {...props}>
      <g transform="rotate(38 24 24)">
      <rect x="18.5" y="2" width="11" height="7.5" rx="2.4" fill="var(--motif-rose)" stroke="var(--motif-rose-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <rect x="18.5" y="9.5" width="11" height="4" fill="var(--card)" stroke="var(--motif-gold-line)" strokeWidth="1.2" strokeLinejoin="round" strokeLinecap="round" />
      <rect x="18.5" y="13.5" width="11" height="19" fill="var(--motif-gold)" stroke="var(--motif-gold-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M24 13.8 V32.2" fill="none" stroke="var(--motif-gold-line)" strokeWidth=".9" opacity=".6"/>
      <path d="M18.5 32.5 L24 45 L29.5 32.5Z" fill="var(--motif-honey)" stroke="var(--motif-honey-line)" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
      <path d="M21.7 39.2 L24 45 L26.3 39.2Z" fill="var(--foreground)" stroke="none"/>
      </g>
    </Frame>
  );
}

export function MotifDivider() {
  return (
    <div aria-hidden="true" className="flex items-center justify-center gap-4 py-2">
      <Sparkle size={18} className="motif-twinkle" style={{ "--delay": "-0.4s" } as CSSProperties} />
      <Sprout size={18} className="motif-float" style={{ "--delay": "-2.2s" } as CSSProperties} />
      <Sparkle size={18} className="motif-twinkle" style={{ "--delay": "-1.6s" } as CSSProperties} />
    </div>
  );
}
