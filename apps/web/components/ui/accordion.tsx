import * as React from "react";
import { Icon } from "@/components/icons";
import { cn } from "@/lib/utils";

// shadcn-style accordion on native <details>: keyboard and screen-reader support with no JavaScript.
function Accordion({ className, ...props }: React.ComponentProps<"div">) {
  return <div data-slot="accordion" className={cn("border-t border-border", className)} {...props} />;
}

function AccordionItem({
  question,
  children,
  className,
  ...props
}: React.ComponentProps<"details"> & { question: React.ReactNode }) {
  return (
    <details data-slot="accordion-item" className={cn("group border-b border-border", className)} {...props}>
      <summary className="flex cursor-pointer items-center justify-between gap-4 py-5.5 text-[19px] font-semibold">
        <span>{question}</span>
        <span className="chevron text-muted-foreground transition-transform duration-300">
          <Icon name="chevron" size={20} />
        </span>
      </summary>
      <div className="pb-5.5 text-base leading-relaxed text-muted-foreground">{children}</div>
    </details>
  );
}

export { Accordion, AccordionItem };
