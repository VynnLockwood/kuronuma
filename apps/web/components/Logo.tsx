import Link from "next/link";
import { BRAND } from "@/lib/brand";
import { cn } from "@/lib/utils";

// Logo illustration plus the wordmark, sized for the 52 px store-style header.
export function Logo({ href = "/", className }: { href?: string; className?: string }) {
  return (
    <Link
      href={href}
      aria-label={`${BRAND.name} home`}
      className={cn("flex items-center gap-2 text-[17px] font-semibold tracking-[-0.01em] text-foreground", className)}
    >
      <img src="/brand/nemo-logo.png" width={28} height={28} alt="" className="size-7" />
      <span>{BRAND.name}</span>
    </Link>
  );
}
