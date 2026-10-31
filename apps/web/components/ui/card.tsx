import * as React from "react";
import { Sparkle } from "@/components/motifs";
import { cn } from "@/lib/utils";

// Store-style card: large radius, no shadow. `bg-muted` cards sit on white, white cards sit on `bg-muted` sections.
// `cute` adds a small lift and a sparkle that fades in on hover (only on devices that can hover).
function Card({ className, cute = false, children, ...props }: React.ComponentProps<"div"> & { cute?: boolean }) {
  return (
    <div
      data-slot="card"
      className={cn(
        "rounded-3xl bg-card p-8 text-card-foreground",
        cute && "group relative transition-transform duration-[var(--dur-base)] ease-[var(--ease-soft)] hover:-translate-y-1",
        className,
      )}
      {...props}
    >
      {children}
      {cute && (
        <Sparkle
          size={20}
          className="absolute top-5 right-5 opacity-0 transition-opacity duration-[var(--dur-base)] group-hover:opacity-100"
        />
      )}
    </div>
  );
}

function CardTitle({ className, ...props }: React.ComponentProps<"h3">) {
  return <h3 data-slot="card-title" className={cn("text-[21px] font-semibold tracking-[-0.015em]", className)} {...props} />;
}

function CardDescription({ className, ...props }: React.ComponentProps<"p">) {
  return <p data-slot="card-description" className={cn("text-base leading-relaxed text-muted-foreground", className)} {...props} />;
}

export { Card, CardTitle, CardDescription };
