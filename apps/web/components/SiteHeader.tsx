import { Logo } from "@/components/Logo";

// Thin, translucent header shared by every page. `nav` sits in the middle, `children` on the right.
export function SiteHeader({
  home = "/",
  nav,
  children,
}: {
  home?: string;
  nav?: React.ReactNode;
  children?: React.ReactNode;
}) {
  return (
    <header className="sticky top-0 z-50 border-b border-hairline bg-background/80 backdrop-blur-xl backdrop-saturate-150">
      <div className="mx-auto flex h-13 max-w-[1120px] items-center justify-between gap-6 px-6">
        <Logo href={home} />
        {nav}
        <div className="flex items-center gap-3">{children}</div>
      </div>
    </header>
  );
}
