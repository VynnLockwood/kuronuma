import Link from "next/link";
import { dashboardCopy } from "@/lib/dashboard-copy";
import type { Lang } from "@/lib/web";
import { cn } from "@/lib/utils";

// Navigation for the signed-in pages. On the Thai dashboard the labels are Thai and "Dashboard" stays on /th/dashboard.
export function AppNav({ current, lang = "en" }: { current: "/dashboard" | "/connect"; lang?: Lang }) {
  const nav = dashboardCopy[lang].shell.nav;
  const links = [
    { href: "/dashboard", to: lang === "th" ? "/th/dashboard" : "/dashboard", label: nav.dashboard },
    { href: "/connect", to: "/connect", label: nav.connect },
  ];
  return (
    <nav aria-label="App" lang={dashboardCopy[lang].htmlLang} className="flex items-center gap-3 text-[13px] sm:gap-6">
      {links.map((l) => (
        <Link
          key={l.href}
          href={l.to}
          aria-current={l.href === current ? "page" : undefined}
          className={cn("text-ink-2 hover:text-foreground", l.href === current && "font-semibold text-foreground")}
        >
          {l.label}
        </Link>
      ))}
    </nav>
  );
}
