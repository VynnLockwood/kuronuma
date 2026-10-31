import Link from "next/link";
import { AuthButton } from "@/components/AuthButton";
import { AuthErrorRedirect } from "@/components/AuthErrorRedirect";
import { Icon, type IconName } from "@/components/icons";
import { Logo } from "@/components/Logo";
import { SiteHeader } from "@/components/SiteHeader";
import { Accordion, AccordionItem } from "@/components/ui/accordion";
import { Badge } from "@/components/ui/badge";
import { buttonVariants } from "@/components/ui/button";
import { Card, CardDescription, CardTitle } from "@/components/ui/card";
import { BRAND, REPO_URL } from "@/lib/brand";
import { Cloud, Letter, MotifDivider, Petal, Sparkle, Sprout, Star } from "@/components/motifs";
import { cn } from "@/lib/utils";
import { RevealController } from "./Reveal";
import { copy, type Lang } from "./copy";

// Marks an element for the scroll reveal (see Reveal.tsx); `i` staggers siblings by 60 ms.
const rv = (i = 0) => ({ "data-reveal": "", style: { "--d": `${i * 60}ms` } as React.CSSProperties });

const home = (lang: Lang) => (lang === "th" ? "/th" : "/");

const stepIcons: IconName[] = ["plug", "layers", "chat"];

function LangSwitch({ lang, label }: { lang: Lang; label: string }) {
  const item = (active: boolean) =>
    cn(
      "rounded-full px-2.5 py-1 transition-colors",
      active ? "bg-background font-semibold text-foreground shadow-[0_1px_2px_rgba(0,0,0,0.08)]" : "text-muted-foreground hover:text-foreground",
    );
  return (
    <div role="group" aria-label={label} className="flex rounded-full bg-muted p-0.5 text-xs">
      <Link href="/" hrefLang="en" lang="en" aria-current={lang === "en" ? "true" : undefined} className={item(lang === "en")}>
        EN
      </Link>
      <Link href="/th" hrefLang="th" lang="th" aria-current={lang === "th" ? "true" : undefined} className={item(lang === "th")}>
        ไทย
      </Link>
    </div>
  );
}

function SectionHead({ title, sub, center = false }: { title: string; sub?: string; center?: boolean }) {
  return (
    <div className={cn("max-w-3xl", center && "mx-auto text-center")} {...rv()}>
      <h2 className="text-4xl leading-[1.08] font-bold tracking-[-0.03em] text-balance md:text-5xl">{title}</h2>
      {sub && <p className="mt-4 text-[19px] leading-normal text-muted-foreground">{sub}</p>}
    </div>
  );
}

function Section({ id, muted, children, className }: { id?: string; muted?: boolean; children: React.ReactNode; className?: string }) {
  return (
    <section id={id} className={cn("px-6 py-24 md:py-30", muted && "bg-muted", className)}>
      <div className="mx-auto max-w-[1120px]">{children}</div>
    </section>
  );
}

export function Landing({ lang }: { lang: Lang }) {
  const c = copy[lang];

  return (
    <div lang={c.htmlLang}>
      <RevealController />
      <AuthErrorRedirect />
      <a
        href="#main"
        className="sr-only focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus:z-[60] focus:rounded-full focus:bg-foreground focus:px-4 focus:py-2 focus:text-background"
      >
        Skip to content
      </a>

      <SiteHeader
        home={home(lang)}
        nav={
          <nav aria-label="Primary" className="hidden gap-7 text-[13px] lg:flex">
            {c.nav.map((n) => (
              <a key={n.id} href={`#${n.id}`} className="text-foreground/85 hover:text-foreground">
                {n.label}
              </a>
            ))}
          </nav>
        }
      >
        <LangSwitch lang={lang} label={c.switchLabel} />
        <AuthButton signIn={c.signIn} openApp={c.openApp} className={cn(buttonVariants({ size: "sm" }), "max-[419px]:hidden")} />
      </SiteHeader>

      <main id="main">
        {/* Hero */}
        <section className="relative overflow-x-clip px-6 pt-20 text-center md:pt-24">
          <div aria-hidden="true" data-motif-layer className="pointer-events-none absolute inset-x-0 top-0 h-20 md:h-24">
            <Sparkle size={36} className="absolute top-5 left-[7%]" />
            <Star size={30} className="absolute top-7 right-[9%] md:size-9" />
          </div>
          <p className="mb-3 inline-flex items-center gap-1.5 text-[17px] font-semibold text-link">
            <Icon name="sparkle" size={16} /> {c.hero.eyebrow}
          </p>
          <h1 className="mx-auto max-w-4xl text-[44px] leading-[1.05] font-bold tracking-[-0.035em] text-balance sm:text-6xl md:text-[80px]">
            {c.hero.title[0]} <span className="text-subtle">{c.hero.title[1]}</span>
          </h1>
          <p className="mx-auto mt-6 max-w-2xl text-lg leading-normal text-muted-foreground md:text-[21px]">{c.hero.sub}</p>
          <div className="mt-8 flex flex-wrap items-center justify-center gap-6">
            <AuthButton signIn={c.hero.primary} openApp={c.openApp} className={buttonVariants({ size: "lg" })} />
            <a href="#how" className={buttonVariants({ variant: "link", className: "text-[17px]" })}>
              {c.hero.secondary} ›
            </a>
          </div>
          <ul aria-label="Highlights" className="mx-auto mt-7 flex max-w-3xl flex-wrap justify-center gap-2">
            {c.hero.chips.map((chip) => (
              <li key={chip} className="rounded-full bg-muted px-3.5 py-1.5 text-[13px] text-ink-2">
                {chip}
              </li>
            ))}
          </ul>
          <p className="mt-5 text-sm text-muted-foreground">{c.hero.worksWith}</p>

          {/* Product shot: an example chat and the memory panel */}
          <div className="relative mx-auto mt-16 max-w-[1120px] overflow-hidden rounded-t-[32px] bg-muted px-4 pt-10 sm:px-6 md:pt-14">
            <div aria-hidden="true" data-motif-layer className="pointer-events-none absolute inset-x-0 top-0 h-10 md:h-14">
              <Petal size={32} className="absolute top-1 left-4 md:top-2 md:left-8 md:size-11" />
              <Sprout size={32} className="absolute top-1 right-4 md:top-2 md:right-8 md:size-11" />
            </div>
            <div className="mx-auto grid max-w-[880px] items-end gap-5 text-left md:grid-cols-2">
              <div role="img" aria-label={c.mock.example} className="rounded-t-3xl bg-background p-6 shadow-[0_20px_60px_rgba(0,0,0,0.08)]">
                <div className="flex items-center gap-2 border-b border-hairline pb-4">
                  <span className="size-2 rounded-full bg-live" />
                  <strong className="text-sm">{c.mock.assistant}</strong>
                  <Badge className="ml-auto text-ink-2">{c.mock.connected}</Badge>
                </div>
                <div className="mt-4 flex justify-end">
                  <span className="rounded-[18px] bg-primary px-3.5 py-2.5 text-sm text-primary-foreground">{c.mock.youSaid}</span>
                </div>
                <div className="mt-3 rounded-[18px] bg-muted p-3.5 text-sm">
                  <p className="mb-2.5">{c.mock.assistantSaid}</p>
                  <dl className="grid grid-cols-3 gap-2">
                    {[
                      [c.mock.amount, "฿450.00"],
                      [c.mock.payee, "Grab"],
                      [c.mock.date, "28 Oct"],
                    ].map(([k, v]) => (
                      <div key={k} className="rounded-xl bg-background p-2.5">
                        <dt className="text-[11px] text-muted-foreground">{k}</dt>
                        <dd className="mt-0.5 font-semibold">{v}</dd>
                      </div>
                    ))}
                  </dl>
                  <div className="mt-3 flex gap-2">
                    <span className="rounded-full bg-foreground px-3.5 py-1.5 text-[13px] text-background">{c.mock.save}</span>
                    <span className="rounded-full border border-border bg-background px-3.5 py-1.5 text-[13px]">{c.mock.edit}</span>
                  </div>
                </div>
                <p className="mt-3.5 flex items-center gap-1.5 text-[13px] text-success">
                  <Icon name="check" size={14} /> {c.mock.saved}
                </p>
              </div>

              <div role="img" aria-label={c.mock.example} className="rounded-t-3xl bg-background p-6 shadow-[0_20px_60px_rgba(0,0,0,0.08)]">
                <p className="mb-3.5 text-xs font-semibold tracking-[0.04em] text-muted-foreground uppercase">{c.hero.panelTitle}</p>
                <ul className="flex flex-col">
                  {c.modules.items.map((m) => (
                    <li key={m.god} className="flex items-center justify-between gap-3 border-b border-hairline py-3 last:border-0">
                      <span className={cn("text-[15px]", !m.live && "text-muted-foreground")}>
                        {m.area} <span className="text-muted-foreground">· {m.god}</span>
                      </span>
                      <Badge variant={m.live ? "success" : "muted"}>{m.live ? c.modules.available : c.modules.soon}</Badge>
                    </li>
                  ))}
                </ul>
                <p className="mt-3 text-[13px] text-muted-foreground">{c.hero.panelNote}</p>
              </div>
            </div>
          </div>
        </section>

        {/* Demo */}
        <Section id="demo">
          <div className="grid items-center gap-16 md:grid-cols-2">
            <div {...rv()}>
              <p className="mb-2 text-[17px] font-semibold text-muted-foreground">{c.demo.eyebrow}</p>
              <h2 className="text-4xl leading-[1.08] font-bold tracking-[-0.03em] md:text-5xl">{c.demo.title}</h2>
              <p className="mt-5 text-[19px] leading-normal text-muted-foreground">{c.demo.sub}</p>
            </div>
            <ol className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              {c.demo.points.map((t, i) => (
                <li key={t} className="rounded-3xl bg-muted p-7" {...rv(i)}>
                  <span className="text-[13px] font-semibold text-link">{i + 1}</span>
                  <p className="mt-2 text-[19px] font-semibold tracking-[-0.01em]">{t}</p>
                </li>
              ))}
            </ol>
          </div>
        </Section>

        {/* How it works */}
        <Section id="how" muted>
          <SectionHead title={c.how.title} sub={c.how.sub} />
          <ol className="mt-12 grid gap-5 md:grid-cols-3">
            {c.how.steps.map((s, i) => (
              <li key={s.title} {...rv(i)}>
                <Card className="flex h-full min-h-[220px] flex-col gap-3 rounded-[28px] p-9">
                  <span className="inline-flex size-11 items-center justify-center rounded-full bg-muted" aria-hidden="true">
                    <Icon name={stepIcons[i] ?? "check"} size={22} />
                  </span>
                  <CardTitle className="mt-2 text-2xl tracking-[-0.02em]">{s.title}</CardTitle>
                  <CardDescription>{s.body}</CardDescription>
                </Card>
              </li>
            ))}
          </ol>
        </Section>

        {/* Features */}
        <Section id="features">
          <SectionHead title={c.features.title} sub={c.features.sub} />
          <ul className="mt-12 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
            {c.features.items.map((f, i) => (
              <li key={f.title} {...rv(i % 3)}>
                <Card className="h-full min-h-[200px] rounded-[28px] bg-muted p-9">
                  <Icon name={f.icon} size={28} />
                  <CardTitle className="mt-5 mb-2">{f.title}</CardTitle>
                  <CardDescription>{f.body}</CardDescription>
                </Card>
              </li>
            ))}
          </ul>
        </Section>

        {/* Modules */}
        <Section id="modules" muted>
          <SectionHead title={c.modules.title} sub={c.modules.sub} />
          <ul className="mt-12 grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
            {c.modules.items.map((m, i) => (
              <li key={m.god} {...rv(i)}>
                <Card className="flex h-full min-h-[220px] flex-col rounded-[28px]">
                  <Badge variant={m.live ? "success" : "muted"}>{m.live ? c.modules.available : c.modules.soon}</Badge>
                  <h3 className={cn("mt-auto pt-12 text-[28px] font-bold tracking-[-0.02em]", !m.live && "text-muted-foreground")}>{m.god}</h3>
                  <p className={cn("text-[15px] font-semibold", !m.live && "text-muted-foreground")}>{m.area}</p>
                  <p className="mt-1.5 text-[15px] text-muted-foreground">{m.body}</p>
                </Card>
              </li>
            ))}
          </ul>
        </Section>

        {/* Mission */}
        <Section id="mission">
          <div className="grid gap-16 md:grid-cols-2">
            <div {...rv()}>
              <h2 className="text-4xl leading-[1.08] font-bold tracking-[-0.03em] md:text-5xl">{c.mission.title}</h2>
              <p className="mt-5 text-[19px] leading-normal text-muted-foreground">{c.mission.sub}</p>
              <p className="mt-5 text-sm text-muted-foreground">{c.mission.note}</p>
            </div>
            <ol className="flex flex-col">
              {c.mission.items.map((m, i) => (
                <li key={m.title} className="border-t border-border py-6 last:border-b" {...rv(i)}>
                  <h3 className="text-[21px] font-semibold">{m.title}</h3>
                  <p className="mt-1.5 text-base text-muted-foreground">{m.body}</p>
                </li>
              ))}
            </ol>
          </div>
        </Section>

        {/* Privacy */}
        <Section id="privacy" muted>
          <SectionHead title={c.privacy.title} sub={c.privacy.sub} />
          <div className="mt-12 grid gap-5 md:grid-cols-2">
            {[
              { title: c.privacy.storesTitle, items: c.privacy.stores, icon: "check" as const, tone: "text-success" },
              { title: c.privacy.notTitle, items: c.privacy.nots, icon: "cross" as const, tone: "text-subtle" },
            ].map((group) => (
              <Card key={group.title} className="rounded-[28px] p-10" {...rv()}>
                <CardTitle className="mb-5">{group.title}</CardTitle>
                <ul className="flex flex-col gap-3.5 text-base leading-snug">
                  {group.items.map((t) => (
                    <li key={t} className="flex gap-3">
                      <Icon name={group.icon} size={20} className={cn("mt-0.5 shrink-0", group.tone)} />
                      {t}
                    </li>
                  ))}
                </ul>
              </Card>
            ))}
          </div>
          <p className="mt-6 flex max-w-3xl items-start gap-2.5 text-sm leading-normal text-muted-foreground">
            <Icon name="lock" size={18} className="mt-0.5 shrink-0" /> <span>{c.privacy.note}</span>
          </p>
        </Section>

        {/* FAQ */}
        <MotifDivider />

        <section id="faq" className="px-6 py-24 md:py-30">
          <div className="mx-auto max-w-[760px]">
            <h2 className="mb-10 text-center text-4xl leading-[1.08] font-bold tracking-[-0.03em] md:text-5xl" {...rv()}>
              {c.faq.title}
            </h2>
            <Accordion>
              {c.faq.items.map((f) => (
                <AccordionItem key={f.q} question={f.q}>
                  {f.a}
                </AccordionItem>
              ))}
            </Accordion>
          </div>
        </section>

        {/* Final call to action */}
        <MotifDivider />

        <section className="px-6 pb-24 md:pb-30">
          <div className="mx-auto max-w-[1120px] rounded-[32px] bg-muted px-6 py-22 text-center" {...rv()}>
            <div className="flex items-center justify-center gap-6">
              <Letter size={40} />
              <img src="/brand/kuronuma-logo.png" width={80} height={80} alt="" className="size-20" />
              <Sparkle size={34} />
            </div>
            <h2 className="mt-6 text-4xl leading-[1.08] font-bold tracking-[-0.03em] md:text-5xl">{c.cta.title}</h2>
            <p className="mx-auto mt-4 max-w-xl text-[19px] text-muted-foreground">{c.cta.sub}</p>
            <div className="mt-8 flex flex-wrap items-center justify-center gap-6">
              <AuthButton signIn={c.signIn} openApp={c.openApp} className={buttonVariants({ size: "lg" })} />
              <Link href="/connect" className={buttonVariants({ variant: "link", className: "text-[17px]" })}>
                {c.cta.guide} ›
              </Link>
            </div>
          </div>
        </section>
      </main>

      <footer className="relative bg-muted px-6 pt-10 pb-6 text-xs text-muted-foreground">
        <Cloud size={56} className="absolute top-0 right-6 -translate-y-1/2 opacity-90" />
        <div className="mx-auto max-w-[1120px]">
          <div className="grid gap-8 border-b border-border pb-7 sm:grid-cols-2 lg:grid-cols-4">
            <div>
              <Logo href={home(lang)} className="text-[15px]" />
              <p className="mt-3 leading-relaxed">{c.footer.blurb}</p>
            </div>
            <nav aria-label={c.footer.product}>
              <h4 className="mb-2 font-semibold text-foreground">{c.footer.product}</h4>
              <ul className="flex flex-col gap-2">
                {c.nav.slice(0, 4).map((n) => (
                  <li key={n.id}>
                    <a href={`#${n.id}`} className="text-ink-2 hover:text-foreground hover:underline">
                      {n.label}
                    </a>
                  </li>
                ))}
              </ul>
            </nav>
            <nav aria-label={c.footer.resources}>
              <h4 className="mb-2 font-semibold text-foreground">{c.footer.resources}</h4>
              <ul className="flex flex-col gap-2">
                <li>
                  <Link href="/connect" className="text-ink-2 hover:text-foreground hover:underline">
                    {c.footer.guide}
                  </Link>
                </li>
                {REPO_URL && (
                  <li>
                    <a href={REPO_URL} rel="noopener noreferrer" className="text-ink-2 hover:text-foreground hover:underline">
                      {c.footer.source}
                    </a>
                  </li>
                )}
                <li>{c.footer.license}</li>
              </ul>
            </nav>
            <div>
              <h4 className="mb-2 font-semibold text-foreground">{c.switchLabel}</h4>
              <LangSwitch lang={lang} label={c.switchLabel} />
            </div>
          </div>
          <div className="flex flex-wrap justify-between gap-2 pt-4">
            <span>
              © {new Date().getFullYear()} {BRAND.name} ({BRAND.thai})
            </span>
            <span>{c.footer.built}</span>
          </div>
        </div>
      </footer>
    </div>
  );
}
