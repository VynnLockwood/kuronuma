import type { Metadata } from "next";

// The dashboard is per-user, so keep it out of search results.
export const metadata: Metadata = {
  title: "Dashboard",
  robots: { index: false, follow: false },
};

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return children;
}
