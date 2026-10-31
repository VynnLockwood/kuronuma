import { cva, type VariantProps } from "class-variance-authority";
import * as React from "react";
import { cn } from "@/lib/utils";

const badgeVariants = cva("inline-flex w-fit items-center rounded-full px-2.5 py-1 text-xs font-medium whitespace-nowrap", {
  variants: {
    variant: {
      muted: "bg-muted text-muted-foreground",
      success: "bg-success-muted text-success",
      info: "bg-info-muted text-info",
      destructive: "bg-destructive-muted text-destructive",
      outline: "border border-border text-foreground",
      onMuted: "bg-background text-muted-foreground",
    },
  },
  defaultVariants: { variant: "muted" },
});

function Badge({ className, variant, ...props }: React.ComponentProps<"span"> & VariantProps<typeof badgeVariants>) {
  return <span data-slot="badge" className={cn(badgeVariants({ variant }), className)} {...props} />;
}

export { Badge, badgeVariants };
