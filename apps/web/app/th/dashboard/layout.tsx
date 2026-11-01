import type { Metadata } from "next";

// The dashboard is per-user, so keep it out of search results.
export const metadata: Metadata = {
  title: "แดชบอร์ด",
  robots: { index: false, follow: false },
};

export default function DashboardLayoutTh({ children }: { children: React.ReactNode }) {
  return children;
}
