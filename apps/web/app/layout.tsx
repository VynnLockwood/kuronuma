import type { Metadata, Viewport } from "next";
import "@fontsource/prompt/thai-400.css";
import "@fontsource/prompt/thai-500.css";
import "@fontsource/prompt/thai-600.css";
import "@fontsource/prompt/thai-700.css";
import "@fontsource/barlow/latin-500.css";
import "@fontsource/barlow/latin-600.css";
import "@fontsource-variable/geist/index.css";
import "@fontsource-variable/geist-mono/index.css";
import "./tokens.css";
import "./globals.css";
import { copy } from "@/components/landing/copy";
import { SITE_URL } from "@/lib/brand";

const { title, description } = copy.en.meta;

// Social-preview image URLs must be absolute, so they are only added when NEXT_PUBLIC_SITE_URL is set at build time.
export const metadata: Metadata = {
  ...(SITE_URL ? { metadataBase: new URL(SITE_URL) } : {}),
  title: { default: title, template: "%s | Nemo" },
  description,
  applicationName: "Nemo",
  alternates: { canonical: "/", languages: { en: "/", th: "/th" } },
  openGraph: {
    type: "website",
    siteName: "Nemo",
    title,
    description,
    ...(SITE_URL ? { images: [{ url: "/og.png", width: 1200, height: 630, alt: "Nemo" }] } : {}),
  },
  twitter: { card: SITE_URL ? "summary_large_image" : "summary", title, description, ...(SITE_URL ? { images: ["/og.png"] } : {}) },
};

export const viewport: Viewport = { themeColor: "#ffffff" };

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
