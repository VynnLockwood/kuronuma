"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { AppNav } from "@/components/AppNav";
import { Icon, type IconName } from "@/components/icons";
import { SiteHeader } from "@/components/SiteHeader";
import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button, buttonVariants } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { authErrorMessages, fetchSession, signInUrl, signOut, type SessionState, type User } from "@/lib/api";
import { cn } from "@/lib/utils";

const modules: { god: string; area: string; note: string; live: boolean; icon: IconName }[] = [
  { god: "Plutus", area: "Spending", note: "Bank slips, totals and reports", live: true, icon: "receipt" },
  { god: "Atlas", area: "Tasks and notes", note: "Tasks, to-dos and notes", live: true, icon: "check" },
  { god: "Horae", area: "Routines", note: "Daily and weekly routines", live: false, icon: "clock" },
  { god: "Hephaestus", area: "Work data", note: "Notes and data from your work", live: false, icon: "code" },
];

const prompts = ["“Save this slip to Kuronuma.”", "“How much did I spend on food this month?”", "“Add a to-do for Friday: send the report.”"];

function Avatar({ user, size = 28 }: { user: User; size?: number }) {
  if (user.picture) {
    // Google avatar; no-referrer keeps our URL out of Google's logs.
    // eslint-disable-next-line @next/next/no-img-element
    return <img src={user.picture} alt="" width={size} height={size} referrerPolicy="no-referrer" className="rounded-full" style={{ width: size, height: size }} />;
  }
  return (
    <span
      aria-hidden="true"
      className="inline-flex items-center justify-center rounded-full bg-hairline text-xs font-semibold text-ink-2"
      style={{ width: size, height: size }}
    >
      {(user.name || user.email).charAt(0).toUpperCase()}
    </span>
  );
}

function SignInCard({ error }: { error: string | null }) {
  return (
    <div className="flex min-h-[calc(100vh-53px)] items-center justify-center bg-muted px-6 py-16">
      <section className="w-full max-w-[440px] rounded-[28px] bg-background px-10 py-12 text-center shadow-[0_1px_2px_rgba(0,0,0,0.04),0_20px_60px_rgba(0,0,0,0.06)]">
        <img src="/brand/kuronuma-logo.png" width={80} height={80} alt="" className="mx-auto size-20" />
        <h1 className="mt-6 mb-2 text-[32px] font-bold tracking-[-0.025em]">Sign in to Kuronuma</h1>
        <p className="text-base leading-normal text-muted-foreground">Use your Google account to connect your assistant to your own data.</p>
        {error && (
          <Alert variant="destructive" className="mt-6">
            <Icon name="alert" size={18} />
            <div>
              <p className="font-semibold">Sign-in didn’t finish</p>
              <p className="mt-0.5 opacity-90">{error}</p>
            </div>
          </Alert>
        )}
        <a href={signInUrl("/dashboard")} className={buttonVariants({ variant: "dark", size: "lg", className: "mt-8 h-13 w-full" })}>
          <Icon name="google" size={18} /> Continue with Google
        </a>
        <p className="mt-5 text-xs leading-normal text-muted-foreground">Kuronuma stores your Google name, email and photo to sign you in. Never your password.</p>
        <div className="mt-7 border-t border-hairline pt-5">
          <Link href="/" className="text-sm text-link hover:underline">
            ‹ Back to home
          </Link>
        </div>
      </section>
    </div>
  );
}

function Loading() {
  return (
    <div aria-busy="true" aria-label="Loading" className="mx-auto max-w-[1120px] px-6 pt-16">
      <Skeleton className="h-4 w-32" />
      <Skeleton className="mt-3 h-10 w-72" />
      <Skeleton className="mt-10 h-56 w-full rounded-[32px] bg-muted" />
      <div className="mt-16 grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
        {[0, 1, 2, 3].map((i) => (
          <Skeleton key={i} className="h-48 rounded-3xl bg-muted" />
        ))}
      </div>
    </div>
  );
}

function Unreachable({ onRetry }: { onRetry: () => void }) {
  return (
    <div className="flex min-h-[calc(100vh-53px)] items-center justify-center px-6 py-16">
      <section className="max-w-sm text-center">
        <span className="inline-flex size-16 items-center justify-center rounded-full bg-muted text-muted-foreground">
          <Icon name="unplug" size={28} />
        </span>
        <h1 className="mt-5 mb-2 text-2xl font-bold tracking-[-0.02em]">Can’t reach Kuronuma</h1>
        <p role="alert" className="text-[15px] leading-normal text-muted-foreground">
          Kuronuma isn’t responding right now. Please try again in a moment.
        </p>
        <Button className="mt-6" onClick={onRetry}>
          Try again
        </Button>
      </section>
    </div>
  );
}

export default function Dashboard() {
  const [session, setSession] = useState<SessionState | null>(null);
  const [authError, setAuthError] = useState<string | null>(null);
  const [origin, setOrigin] = useState("");

  const load = useCallback(() => {
    setSession(null);
    fetchSession().then(setSession);
  }, []);

  useEffect(() => {
    // Cerberus reports sign-in problems as ?auth_error=code. Show a message, then clean the URL.
    const code = new URLSearchParams(window.location.search).get("auth_error");
    if (code) {
      setAuthError(authErrorMessages[code] ?? authErrorMessages.server_error);
      window.history.replaceState(null, "", window.location.pathname);
    }
    setOrigin(window.location.origin);
    load();
  }, [load]);

  async function handleSignOut() {
    await signOut();
    setSession({ status: "signed-out" });
  }

  const user = session?.status === "signed-in" ? session.user : null;

  return (
    <>
      <SiteHeader>
        <AppNav current="/dashboard" />
        {user && <Avatar user={user} />}
      </SiteHeader>

      <main>
        {session === null && <Loading />}
        {session?.status === "unreachable" && <Unreachable onRetry={load} />}
        {session?.status === "signed-out" && <SignInCard error={authError} />}

        {user && (
          <div className="mx-auto max-w-[1120px] px-6 pt-16 pb-24">
            <div className="flex flex-wrap items-end justify-between gap-6">
              <div>
                <p className="mb-1.5 text-[15px] text-muted-foreground">Welcome back</p>
                <h1 className="text-4xl leading-[1.08] font-bold tracking-[-0.03em] md:text-5xl">{user.name || user.email}</h1>
              </div>
              <div className="flex max-w-full items-center gap-3.5 rounded-full bg-muted py-2 pr-2 pl-4">
                <span className="truncate text-sm text-ink-2">{user.email}</span>
                <Button variant="outline" size="sm" onClick={handleSignOut}>
                  Sign out
                </Button>
              </div>
            </div>

            <section className="mt-10 grid grid-cols-[minmax(0,1fr)] items-center gap-8 rounded-[32px] bg-muted p-8 md:grid-cols-2 md:p-12">
              <div>
                <span className="inline-flex size-14 items-center justify-center rounded-2xl bg-background text-primary">
                  <Icon name="plug" size={26} />
                </span>
                <h2 className="mt-5 mb-2 text-[32px] font-bold tracking-[-0.025em]">Connect your assistant.</h2>
                <p className="text-[17px] leading-normal text-muted-foreground">
                  Add Kuronuma to Claude or any assistant that supports MCP connectors. It takes one URL.
                </p>
              </div>
              <div className="flex flex-col items-start gap-4">
                <div className="flex w-full items-center gap-2 rounded-full border border-hairline bg-background py-1.5 pr-1.5 pl-4.5">
                  <code className="min-w-0 flex-1 truncate font-mono text-sm">{origin ? `${origin}/mcp` : "…"}</code>
                  <Link href="/connect" className={buttonVariants({ className: "h-10" })}>
                    Connect
                  </Link>
                </div>
                <Link href="/connect" className="text-[15px] text-link hover:underline">
                  Read the setup guide ›
                </Link>
              </div>
            </section>

            <div className="mt-18 flex flex-wrap items-baseline justify-between gap-4">
              <h2 className="text-[32px] font-bold tracking-[-0.025em]">Your modules</h2>
              <span className="text-sm text-muted-foreground">
                {modules.filter((m) => m.live).length} of {modules.length} available
              </span>
            </div>
            <ul className="mt-6 grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
              {modules.map((m) => (
                <li key={m.god}>
                  <Card className={cn("flex h-full min-h-[200px] flex-col p-7", m.live ? "border border-hairline" : "bg-muted")}>
                    <div className="flex items-center justify-between">
                      <Icon name={m.icon} size={28} className={m.live ? "text-foreground" : "text-subtle"} />
                      <Badge variant={m.live ? "success" : "onMuted"}>{m.live ? "Available" : "Coming soon"}</Badge>
                    </div>
                    <h3 className={cn("mt-auto pt-10 text-2xl font-bold tracking-[-0.02em]", !m.live && "text-muted-foreground")}>{m.god}</h3>
                    <p className={cn("text-[15px] font-semibold", !m.live && "text-muted-foreground")}>{m.area}</p>
                    {m.live && <p className="mt-1 text-sm text-muted-foreground">{m.note}</p>}
                  </Card>
                </li>
              ))}
            </ul>

            <section className="mt-18">
              <h2 className="text-[32px] font-bold tracking-[-0.025em]">Try asking</h2>
              <ul className="mt-6 grid gap-4 md:grid-cols-3">
                {prompts.map((p) => (
                  <li key={p} className="rounded-[20px] bg-muted p-6 text-[17px] leading-snug">
                    {p}
                  </li>
                ))}
              </ul>
            </section>
          </div>
        )}
      </main>
    </>
  );
}
