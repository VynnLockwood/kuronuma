import { Bird, Bow, Bunny, Balloon, Cat, Cloud, Cupcake, Daisy, Heart, Moon, Mushroom, Pencil, Plane, Rainbow, Strawberry, Teacup, Teddy, type MotifProps } from "@/components/motifs";
import { cn } from "@/lib/utils";

const MOTIFS = { bird: Bird, bow: Bow, bunny: Bunny, balloon: Balloon, cat: Cat, cloud: Cloud, cupcake: Cupcake, daisy: Daisy, heart: Heart, moon: Moon, mushroom: Mushroom, pencil: Pencil, plane: Plane, rainbow: Rainbow, strawberry: Strawberry, teacup: Teacup, teddy: Teddy } as const;
export type MotifName = keyof typeof MOTIFS;

export function Motif({ name, ...props }: MotifProps & { name: MotifName }) {
  const M = MOTIFS[name];
  return <M {...props} />;
}

// Motion is opt-in (`drift`, landing page only). A delay taken from the name starts each one mid-cycle, so they never move in step.
const drifting = (name: string): Pick<MotifProps, "className" | "style"> => ({
  className: "motif-float",
  style: { "--delay": `-${(name.length * 1.7) % 7}s`, "--dur": `${6 + (name.length % 4)}s` } as MotifProps["style"],
});

// One centred picture for a section, in the normal flow (never absolutely placed, so it can never sit on text).
// Two names make one spot: the first at `size`, the second one tier smaller beside it.
export function Spot({ motif, size = 96, className, drift = false }: { motif: MotifName | [MotifName, MotifName]; size?: 64 | 96; className?: string; drift?: boolean }) {
  const [a, b] = Array.isArray(motif) ? motif : [motif];
  return (
    <div data-spot aria-hidden="true" className={cn("pointer-events-none flex items-end justify-center gap-4", className)}>
      <Motif name={a} size={size} {...(drift ? drifting(a) : {})} />
      {b && <Motif name={b} size={64} {...(drift ? drifting(b) : {})} />}
    </div>
  );
}

// A mirrored pair that fills the empty lower corners of a roomy container. Only from `lg` up; below that there is no room.
export function Accents({ left, right, className, drift = false }: { left: MotifName; right: MotifName; className?: string; drift?: boolean }) {
  return (
    <div data-accents aria-hidden="true" className={cn("pointer-events-none absolute inset-x-8 bottom-8 hidden items-end justify-between lg:flex", className)}>
      <Motif name={left} size={32} {...(drift ? drifting(left) : {})} />
      <Motif name={right} size={32} {...(drift ? drifting(right) : {})} />
    </div>
  );
}
