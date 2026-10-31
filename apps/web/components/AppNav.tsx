import Link from "next/link";
import { cn } from "@/lib/utils";

const links = [
  { href: "/dashboard", label: "Dashboard" },
  { href: "/connect", label: "Connect" },
];

// Navigation for the signed-in pages.
export function AppNav({ current }: { current: "/dashboard" | "/connect" }) {
  return (
    <nav aria-label="App" className="flex items-center gap-6 text-[13px]">
      {links.map((l) => (
        <Link
          key={l.href}
          href={l.href}
          aria-current={l.href === current ? "page" : undefined}
          className={cn("text-ink-2 hover:text-foreground", l.href === current && "font-semibold text-foreground")}
        >
          {l.label}
        </Link>
      ))}
    </nav>
  );
}
