"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { AppNav } from "@/components/AppNav";
import { Icon } from "@/components/icons";
import { Bow, Sparkle } from "@/components/motifs";
import { Spot } from "@/components/motifs/Spot";
import { SiteHeader } from "@/components/SiteHeader";
import { Alert } from "@/components/ui/alert";
import { Button, buttonVariants } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { fetchSession, signInUrl, signOut, type SessionState, type User } from "@/lib/api";
import { dashboardCopy } from "@/lib/dashboard-copy";
import type { Lang } from "@/lib/web";
import { cn } from "@/lib/utils";
import { DashContext } from "./context";
import { Overview } from "./Overview";
import { Memories } from "./Memories";
import { Spending } from "./Spending";

// One entry per tab. Later tabs are added here; the shell does not change.
const tabIds = ["overview", "spending", "tasks"] as const;
type TabId = (typeof tabIds)[number];
const isTab = (v: string | null): v is TabId => tabIds.includes(v as TabId);

const base = (lang: Lang) => (lang === "th" ? "/th/dashboard" : "/dashboard");

function Avatar({ user, size = 28 }: { user: User; size?: number }) {
  if (user.picture) {
    // Google avatar; no-referrer keeps our URL out of Google's logs.
    // eslint-disable-next-line @next/next/no-img-element
    return <img src={user.picture} alt="" width={size} height={size} referrerPolicy="no-referrer" className="rounded-full" style={{ width: size, height: size }} />;
  }
  return (
    <span aria-hidden="true" className="inline-flex items-center justify-center rounded-full bg-hairline text-xs font-semibold text-ink-2" style={{ width: size, height: size }}>
      {(user.name || user.email).charAt(0).toUpperCase()}
    </span>
  );
}

function SignInCard({ lang, error }: { lang: Lang; error: string | null }) {
  const c = dashboardCopy[lang].shell.signIn;
  return (
    <div className="flex min-h-[calc(100vh-53px)] items-center justify-center bg-muted px-6 py-16">
      <section className="w-full max-w-[440px] rounded-[28px] bg-background px-10 py-12 text-center shadow-[0_1px_2px_rgba(0,0,0,0.04),0_20px_60px_rgba(0,0,0,0.06)]">
        <div className="flex items-center justify-center gap-5">
          <Bow size={32} />
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src="/brand/kuronuma-logo.png" width={80} height={80} alt="" className="size-20" />
          <Sparkle size={32} />
        </div>
        <h1 className="mt-6 mb-2 text-[32px] font-bold tracking-[-0.025em]">{c.title}</h1>
        <p className="text-base leading-normal text-muted-foreground">{c.sub}</p>
        {error && (
          <Alert variant="destructive" className="mt-6">
            <Icon name="alert" size={18} />
            <div>
              <p className="font-semibold">{c.errorTitle}</p>
              <p className="mt-0.5 opacity-90">{error}</p>
            </div>
          </Alert>
        )}
        <a href={signInUrl(base(lang))} className={buttonVariants({ variant: "dark", size: "lg", className: "mt-8 h-13 w-full" })}>
          <Icon name="google" size={18} /> {c.button}
        </a>
        <p className="mt-5 text-xs leading-normal text-muted-foreground">{c.fine}</p>
        <div className="mt-7 border-t border-hairline pt-5">
          <Link href={lang === "th" ? "/th" : "/"} className="text-sm text-link hover:underline">
            {c.back}
          </Link>
        </div>
      </section>
    </div>
  );
}

function Loading({ label }: { label: string }) {
  return (
    <div aria-busy="true" aria-label={label} className="mx-auto max-w-[1120px] px-6 pt-16">
      <Skeleton className="h-4 w-32" />
      <Skeleton className="mt-3 h-10 w-72" />
      <Skeleton className="mt-10 h-56 w-full rounded-[32px] bg-muted" />
    </div>
  );
}

function Unreachable({ lang, onRetry }: { lang: Lang; onRetry: () => void }) {
  const c = dashboardCopy[lang].shell.unreachable;
  return (
    <div className="flex min-h-[calc(100vh-53px)] items-center justify-center px-6 py-16">
      <section className="max-w-sm text-center">
        <Spot motif="plane" size={64} />
        <h1 className="mt-5 mb-2 text-2xl font-bold tracking-[-0.02em]">{c.title}</h1>
        <p role="alert" className="text-[15px] leading-normal text-muted-foreground">
          {c.body}
        </p>
        <Button className="mt-6" onClick={onRetry}>
          {c.retry}
        </Button>
      </section>
    </div>
  );
}

function LangSwitch({ lang, tab }: { lang: Lang; tab: TabId }) {
  const q = tab === "overview" ? "" : `?tab=${tab}`;
  const item = (active: boolean) =>
    cn("inline-flex min-h-10 items-center rounded-full px-3 transition-colors sm:min-h-0 sm:py-1", active ? "bg-background font-semibold text-foreground shadow-[0_1px_2px_rgba(0,0,0,0.08)]" : "text-muted-foreground hover:text-foreground");
  return (
    <div role="group" aria-label={dashboardCopy[lang].shell.langLabel} className="flex rounded-full bg-muted p-0.5 text-xs">
      <Link href={`/dashboard${q}`} hrefLang="en" lang="en" aria-current={lang === "en" ? "true" : undefined} className={item(lang === "en")}>
        EN
      </Link>
      <Link href={`/th/dashboard${q}`} hrefLang="th" lang="th" aria-current={lang === "th" ? "true" : undefined} className={item(lang === "th")}>
        ไทย
      </Link>
    </div>
  );
}

export function Shell({ lang }: { lang: Lang }) {
  const copy = dashboardCopy[lang];
  const [session, setSession] = useState<SessionState | null>(null);
  const [authError, setAuthError] = useState<string | null>(null);
  const [tab, setTab] = useState<TabId>("overview");

  const load = useCallback(() => {
    setSession(null);
    fetchSession().then(setSession);
  }, []);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    // Cerberus reports sign-in problems as ?auth_error=code. Show a message, then clean the URL.
    const code = params.get("auth_error");
    if (code) {
      setAuthError(copy.shell.authErrors[code] ?? copy.shell.authErrors.server_error);
      params.delete("auth_error");
      const rest = params.toString();
      window.history.replaceState(null, "", window.location.pathname + (rest ? `?${rest}` : ""));
    }
    const t = params.get("tab");
    if (isTab(t)) setTab(t);
    load();
  }, [load, copy]);

  const chooseTab = (id: string) => {
    if (!isTab(id)) return;
    setTab(id);
    window.history.replaceState(null, "", id === "overview" ? base(lang) : `${base(lang)}?tab=${id}`);
  };

  const onSignedOut = useCallback(() => setSession({ status: "signed-out" }), []);
  async function handleSignOut() {
    await signOut();
    setSession({ status: "signed-out" });
  }
  const user = session?.status === "signed-in" ? session.user : null;
  const labels: Record<TabId, string> = { overview: copy.tabs.overview, spending: copy.tabs.spending, tasks: copy.tabs.tasks };

  return (
    <div lang={copy.htmlLang} data-dashboard>
      <SiteHeader home={lang === "th" ? "/th" : "/"} lead={<AppNav current="/dashboard" lang={lang} />}>
        {user && <Avatar user={user} />}
      </SiteHeader>

      <main>
        {session === null && <Loading label={copy.shell.loading} />}
        {session?.status === "unreachable" && <Unreachable lang={lang} onRetry={load} />}
        {session?.status === "signed-out" && <SignInCard lang={lang} error={authError} />}

        {user && (
          <DashContext.Provider value={{ lang, copy, onSignedOut }}>
            <div className="mx-auto max-w-[1120px] px-6 pt-12 pb-24">
              <div className="flex flex-wrap items-end justify-between gap-6">
                <div className="min-w-0">
                  <p className="mb-1.5 text-[15px] text-muted-foreground">{copy.shell.welcome}</p>
                  <h1 className="text-4xl leading-[1.08] font-bold tracking-[-0.03em] break-words md:text-5xl">{user.name || user.email}</h1>
                </div>
                <div className="flex max-w-full flex-wrap items-center gap-3">
                  <LangSwitch lang={lang} tab={tab} />
                  <div className="flex max-w-full items-center gap-3.5 rounded-full bg-muted py-2 pr-2 pl-4">
                    <span className="truncate text-sm text-ink-2">{user.email}</span>
                    <Button variant="outline" size="sm" className="h-10 sm:h-8" onClick={handleSignOut}>
                      {copy.shell.signOut}
                    </Button>
                  </div>
                </div>
              </div>

              <Tabs value={tab} onValueChange={chooseTab} className="mt-10 gap-8">
                <TabsList aria-label={copy.shell.tabsLabel} className="w-full md:w-fit">
                  {tabIds.map((id) => (
                    <TabsTrigger key={id} value={id} className="h-auto min-h-10 min-w-0 flex-1 px-2 py-2 text-center leading-tight whitespace-normal md:flex-none md:px-5 md:whitespace-nowrap">
                      {labels[id]}
                    </TabsTrigger>
                  ))}
                </TabsList>
                <TabsContent value="overview">
                  <Overview />
                </TabsContent>
                <TabsContent value="spending">
                  <Spending />
                </TabsContent>
                <TabsContent value="tasks">
                  <Memories />
                </TabsContent>
              </Tabs>
            </div>
          </DashContext.Provider>
        )}
      </main>
    </div>
  );
}
