import { cva, type VariantProps } from "class-variance-authority";
import * as React from "react";
import { cn } from "@/lib/utils";

const alertVariants = cva("flex w-full gap-2.5 rounded-2xl px-4 py-3.5 text-left text-sm [&>svg]:mt-0.5 [&>svg]:shrink-0", {
  variants: {
    variant: {
      default: "bg-muted text-foreground",
      success: "bg-success-muted text-success font-medium",
      destructive: "border border-destructive/20 bg-destructive-muted text-destructive",
    },
  },
  defaultVariants: { variant: "default" },
});

function Alert({ className, variant, ...props }: React.ComponentProps<"div"> & VariantProps<typeof alertVariants>) {
  return <div data-slot="alert" role="alert" className={cn(alertVariants({ variant }), className)} {...props} />;
}

export { Alert };
