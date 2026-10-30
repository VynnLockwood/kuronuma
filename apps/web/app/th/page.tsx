import type { Metadata } from "next";
import { Landing } from "@/components/landing/Landing";
import { copy } from "@/components/landing/copy";

export const metadata: Metadata = {
  title: { absolute: copy.th.meta.title },
  description: copy.th.meta.description,
  alternates: { canonical: "/th", languages: { en: "/", th: "/th" } },
  openGraph: { title: copy.th.meta.title, description: copy.th.meta.description, locale: "th_TH" },
};

export default function HomePageTh() {
  return <Landing lang="th" />;
}
