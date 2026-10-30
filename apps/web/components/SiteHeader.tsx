import { Logo } from "@/components/Logo";

// Header for the signed-in app pages (dashboard, connect).
export function SiteHeader({ children }: { children?: React.ReactNode }) {
  return (
    <header className="site-header">
      <div className="container header-row">
        <Logo />
        <div className="header-actions">{children}</div>
      </div>
    </header>
  );
}
