import * as React from "react";
import { cn } from "@/lib/utils";

// Store-style card: large radius, no shadow. `bg-muted` cards sit on white, white cards sit on `bg-muted` sections.
function Card({ className, ...props }: React.ComponentProps<"div">) {
  return <div data-slot="card" className={cn("rounded-3xl bg-card p-8 text-card-foreground", className)} {...props} />;
}

function CardTitle({ className, ...props }: React.ComponentProps<"h3">) {
  return <h3 data-slot="card-title" className={cn("text-[21px] font-semibold tracking-[-0.015em]", className)} {...props} />;
}

function CardDescription({ className, ...props }: React.ComponentProps<"p">) {
  return <p data-slot="card-description" className={cn("text-base leading-relaxed text-muted-foreground", className)} {...props} />;
}

export { Card, CardTitle, CardDescription };
