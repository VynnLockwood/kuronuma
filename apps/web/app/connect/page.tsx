"use client";

import { useEffect, useState } from "react";
import { AppNav } from "@/components/AppNav";
import { Icon, type IconName } from "@/components/icons";
import { SiteHeader } from "@/components/SiteHeader";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

const claudeSteps = [
  "Open Settings, then Connectors, then add a custom connector.",
  "Paste the URL above and save.",
  "Choose Connect. Sign in with Google, review what Nemo asks to access, then approve.",
];

const tryIt: { icon: IconName; say: string; body: string }[] = [
  { icon: "camera", say: "“Save this to Nemo.”", body: "Attach a bank slip photo. Confirm the amount, payee and date it shows you." },
  { icon: "receipt", say: "“How much did I spend on food this month?”", body: "Totals by category, month or merchant." },
  {
    icon: "check",
    say: "“Add a to-do for Friday: send the report. Tag it work.”",
    body: "Later ask: “What is on my to-do list for work?”",
  },
];

export default function Connect() {
  // The MCP endpoint lives on the same origin as this page (nginx routes /mcp to the API).
  const [origin, setOrigin] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    setOrigin(window.location.origin);
  }, []);

  const mcpUrl = origin ? `${origin}/mcp` : "";

  async function copy() {
    try {
      await navigator.clipboard.writeText(mcpUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard can be unavailable (insecure context); the URL is selectable on the page.
    }
  }

  return (
    <>
      <SiteHeader>
        <AppNav current="/connect" />
      </SiteHeader>

      <main>
        <section className="px-6 pt-20 pb-14 text-center md:pt-22">
          <p className="mb-3 text-[17px] font-semibold text-link">Setup guide</p>
          <h1 className="text-[40px] leading-[1.05] font-bold tracking-[-0.035em] md:text-[64px]">Connect your assistant.</h1>
          <p className="mx-auto mt-5 max-w-2xl text-lg leading-normal text-muted-foreground md:text-[21px]">
            One URL. Your assistant signs you in with Google and asks for your approval.
          </p>
        </section>

        <section className="px-6">
          <div className="mx-auto max-w-[760px] rounded-[28px] bg-muted p-6 sm:p-9">
            <label htmlFor="mcp-url" className="mb-3 block text-sm font-semibold">
              Your Nemo MCP URL
            </label>
            <div className="flex flex-wrap gap-2.5">
              <Input
                id="mcp-url"
                data-testid="mcp-url"
                readOnly
                value={mcpUrl || "Loading…"}
                onFocus={(e) => e.currentTarget.select()}
                className="flex-[1_1_300px] font-mono"
              />
              <Button size="lg" onClick={copy} disabled={!mcpUrl} aria-live="polite">
                <Icon name={copied ? "check" : "copy"} size={16} /> {copied ? "Copied" : "Copy"}
              </Button>
            </div>
            <p className="mt-3.5 text-sm leading-normal text-muted-foreground">
              Remote assistants need a public HTTPS address. If you run Nemo on your own computer, only apps on that computer can reach it.
            </p>
          </div>
          <p className="mx-auto mt-4 max-w-[760px] text-center text-sm text-muted-foreground">
            Connected before? Remove Nemo from your assistant’s connectors and add it again to see new tools such as tasks and categories.
          </p>
        </section>

        <section className="px-6 pt-18">
          <Tabs defaultValue="claude" className="mx-auto max-w-[760px]">
            <TabsList aria-label="Assistant">
              <TabsTrigger value="claude">Claude</TabsTrigger>
              <TabsTrigger value="code">Claude Code</TabsTrigger>
              <TabsTrigger value="other">Other clients</TabsTrigger>
            </TabsList>
            <div className="min-h-[220px] rounded-[28px] border border-hairline p-6 sm:p-10">
              <TabsContent value="claude">
                <h2 className="mb-5 text-2xl font-bold tracking-[-0.02em]">Claude</h2>
                <ol className="flex flex-col gap-4">
                  {claudeSteps.map((s, i) => (
                    <li key={s} className="flex gap-4 text-[17px] leading-snug">
                      <span className="inline-flex size-7 shrink-0 items-center justify-center rounded-full bg-muted text-[13px] font-semibold">{i + 1}</span>
                      {s}
                    </li>
                  ))}
                </ol>
                <p className="mt-6 text-sm text-muted-foreground">Menu names can differ between plans and app versions.</p>
              </TabsContent>
              <TabsContent value="code">
                <h2 className="mb-5 text-2xl font-bold tracking-[-0.02em]">Claude Code</h2>
                <pre className="rounded-2xl bg-foreground p-5 font-mono text-sm leading-normal break-all whitespace-pre-wrap text-background">
                  {`claude mcp add --transport http nemosyne ${mcpUrl || "<your-url>/mcp"}`}
                </pre>
                <p className="mt-5 text-[17px] leading-snug">
                  Then run <code className="rounded-md bg-muted px-1.5 py-0.5 font-mono text-[15px]">/mcp</code> inside Claude Code and choose nemosyne to sign in.
                </p>
              </TabsContent>
              <TabsContent value="other">
                <h2 className="mb-5 text-2xl font-bold tracking-[-0.02em]">Other MCP clients</h2>
                <p className="text-[17px] leading-normal">
                  Add a remote (Streamable HTTP) server with the URL above. Clients that support OAuth discovery register and sign in automatically.
                </p>
              </TabsContent>
            </div>
          </Tabs>
        </section>

        <section className="px-6 pt-24 pb-30">
          <div className="mx-auto max-w-[1120px]">
            <h2 className="text-center text-4xl leading-[1.1] font-bold tracking-[-0.03em]">Then just ask.</h2>
            <ul className="mt-10 grid gap-5 md:grid-cols-3">
              {tryIt.map((t) => (
                <li key={t.say} className="rounded-3xl bg-muted p-8">
                  <Icon name={t.icon} size={28} />
                  <p className="mt-5 mb-1.5 text-[19px] leading-snug font-semibold">{t.say}</p>
                  <p className="text-[15px] text-muted-foreground">{t.body}</p>
                </li>
              ))}
            </ul>
          </div>
        </section>
      </main>
    </>
  );
}
