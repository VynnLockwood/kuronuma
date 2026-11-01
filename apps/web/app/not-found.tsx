import type { Metadata } from "next";
import { NotFoundView, notFoundText } from "@/components/NotFoundView";

export const metadata: Metadata = { title: notFoundText.en.tab, robots: { index: false, follow: false } };

export default function NotFound() {
  return <NotFoundView lang="en" />;
}
