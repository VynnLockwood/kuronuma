import type { Metadata } from "next";
import { NotFoundView, notFoundText } from "@/components/NotFoundView";

export const metadata: Metadata = { title: notFoundText.th.tab, robots: { index: false, follow: false } };

export default function NotFoundTh() {
  return <NotFoundView lang="th" />;
}
