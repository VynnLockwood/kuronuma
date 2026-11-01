"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Spot } from "@/components/motifs/Spot";
import { SiteHeader } from "@/components/SiteHeader";
import { buttonVariants } from "@/components/ui/button";

const text = {
  en: { lang: "en", home: "/", title: "This page wandered off", body: "We could not find what you were looking for. The link may be wrong, or the page may have moved.", back: "Back to home" },
  th: { lang: "th", home: "/th", title: "ไม่พบหน้านี้", body: "ไม่พบหน้าที่คุณต้องการ ลิงก์อาจไม่ถูกต้อง หรือหน้านี้อาจถูกย้ายไปแล้ว", back: "กลับหน้าแรก" },
} as const;

// Thai when the address starts with /th, English otherwise.
export function NotFoundView() {
  const path = usePathname() ?? "/";
  const t = path === "/th" || path.startsWith("/th/") ? text.th : text.en;
  return (
    <div lang={t.lang}>
      <SiteHeader home={t.home} />
      <main className="flex min-h-[calc(100vh-53px)] flex-col items-center justify-center px-6 py-16 text-center">
        <Spot motif="bunny" />
        <h1 className="mt-6 text-[32px] leading-[1.1] font-bold tracking-[-0.03em] md:text-5xl">{t.title}</h1>
        <p className="mx-auto mt-4 max-w-md text-[17px] leading-normal text-muted-foreground">{t.body}</p>
        <Link href={t.home} className={buttonVariants({ size: "lg", className: "mt-8" })}>
          {t.back}
        </Link>
      </main>
    </div>
  );
}
