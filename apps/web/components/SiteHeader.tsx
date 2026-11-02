import { Logo } from "@/components/Logo";
import { RepoLink } from "@/components/RepoLink";

// Thin, translucent header shared by every page. `nav` sits in the middle. The right-hand group reads, in order:
// `lead` (the app navigation on signed-in pages), the repository button, then `children` (language switch, sign-in, avatar).
export function SiteHeader({
  home = "/",
  nav,
  lead,
  children,
}: {
  home?: string;
  nav?: React.ReactNode;
  lead?: React.ReactNode;
  children?: React.ReactNode;
}) {
  return (
    <header className="sticky top-0 z-50 border-b border-hairline bg-background/80 backdrop-blur-xl backdrop-saturate-150">
      <div className="mx-auto flex h-13 max-w-[1120px] items-center justify-between gap-2 px-4 sm:gap-6 sm:px-6">
        <Logo href={home} />
        {nav}
        <div className="flex items-center gap-2 sm:gap-3">
          {lead}
          <RepoLink lang={home === "/th" ? "th" : "en"} />
          {children}
        </div>
      </div>
    </header>
  );
}
