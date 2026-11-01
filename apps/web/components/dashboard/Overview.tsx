"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { Icon } from "@/components/icons";
import { buttonVariants } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { compactAmount, compareAmount, dueBucket, formatAmount, lastMonths, localISODate, monthName, monthStartISO, share } from "@/lib/format";
import { BarList, Columns } from "./charts";
import { useDash } from "./context";
import { BlockError, BlockSkeleton } from "./states";
import { useTool } from "./useTool";

type Group = { count: number; currency: string; key: string; total: string };
type Summary = { groups: Group[]; totals_by_currency: { count: number; currency: string; total: string }[] };
type Memory = { id: string; kind: string; title: string; status: string; due_on?: string; created_at: string };
type Memories = { memories: Memory[] };
type Categories = { categories: { name: string }[] };

const plain = (amount: string) => formatAmount(amount, "").trim();

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section>
      <h2 className="mb-4 text-[22px] font-bold tracking-[-0.02em]">{title}</h2>
      {children}
    </section>
  );
}

export function Overview() {
  const { lang, copy } = useDash();
  const o = copy.overview;
  const today = useMemo(() => localISODate(new Date()), []);
  const months = useMemo(() => lastMonths(today, 6), [today]);
  const [origin, setOrigin] = useState("");
  useEffect(() => setOrigin(window.location.origin), []);

  const [month, reloadMonth] = useTool<Summary>("get_spending_summary", { group_by: "category", from: monthStartISO(today), to: today });
  const [series, reloadSeries] = useTool<Summary>("get_spending_summary", { group_by: "month", from: `${months[0]}-01`, to: today });
  const [open, reloadOpen] = useTool<Memories>("list_memories", { status: "open", limit: 200 });
  const [notes, reloadNotes] = useTool<Memories>("list_memories", { kind: "note", limit: 200 });
  const [cats, reloadCats] = useTool<Categories>("list_categories", {});

  const allLoaded = [month, series, open, notes, cats].every((s) => s.status === "ok");
  const allEmpty =
    allLoaded &&
    month.status === "ok" && month.data.totals_by_currency.length === 0 &&
    series.status === "ok" && series.data.groups.length === 0 &&
    open.status === "ok" && open.data.memories.length === 0 &&
    notes.status === "ok" && notes.data.memories.length === 0;

  const connectCard = (
    <Card className="grid grid-cols-[minmax(0,1fr)] items-center gap-5 rounded-[28px] bg-muted p-7 md:grid-cols-2">
      <div>
        <span className="inline-flex size-12 items-center justify-center rounded-2xl bg-background text-primary">
          <Icon name="plug" size={24} />
        </span>
        <h2 className="mt-4 mb-1.5 text-2xl font-bold tracking-[-0.02em]">{o.connect.title}</h2>
        <p className="text-[15px] leading-normal text-muted-foreground">{o.connect.body}</p>
      </div>
      <div className="flex flex-col items-start gap-3">
        <div className="flex w-full min-w-0 items-center gap-2 rounded-full border border-hairline bg-background py-1.5 pr-1.5 pl-4">
          <code className="min-w-0 flex-1 truncate font-mono text-sm">{origin ? `${origin}/mcp` : "…"}</code>
          <Link href="/connect" className={buttonVariants({ className: "h-10" })}>
            {o.connect.button}
          </Link>
        </div>
        <Link href="/connect" className="text-[15px] text-link hover:underline">
          {o.connect.guide}
        </Link>
      </div>
    </Card>
  );

  if (allEmpty) {
    return (
      <div className="flex flex-col gap-10">
        <Card className="rounded-[28px] bg-muted p-8 text-center">
          <h2 className="text-2xl font-bold tracking-[-0.02em]">{o.emptyTitle}</h2>
          <p className="mx-auto mt-2 max-w-md text-[15px] leading-normal text-muted-foreground">{o.emptyBody}</p>
          <Link href="/connect" className={buttonVariants({ className: "mt-5" })}>
            {copy.common.connectLink}
          </Link>
        </Card>
        <Section title={o.tryAsking}>
          <ul className="grid gap-4 md:grid-cols-3">
            {o.prompts.map((p) => (
              <li key={p} className="rounded-[20px] bg-muted p-6 text-[17px] leading-snug">
                {p}
              </li>
            ))}
          </ul>
        </Section>
      </div>
    );
  }

  // Spending, one currency at a time: amounts in different currencies are never added together.
  const currencies = month.status === "ok" ? month.data.totals_by_currency : [];

  return (
    <div className="flex flex-col gap-12">
      <Section title={o.thisMonth}>
        {month.status === "loading" && <BlockSkeleton />}
        {month.status === "error" && <BlockError message={month.message} onRetry={reloadMonth} />}
        {month.status === "ok" &&
          (currencies.length === 0 ? (
            <p className="text-[15px] text-muted-foreground">{o.noSpendingMonth}</p>
          ) : (
            <div className="flex flex-col gap-6">
              <ul className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                {currencies.map((c) => (
                  <li key={c.currency}>
                    <Card className="rounded-3xl bg-muted p-6">
                      <p className="text-[28px] font-bold tracking-[-0.02em] break-words tabular-nums">{formatAmount(c.total, c.currency)}</p>
                      <p className="mt-1 text-sm text-muted-foreground">{o.records(c.count)}</p>
                    </Card>
                  </li>
                ))}
              </ul>
              {currencies.map((c) => {
                const rows = month.data.groups
                  .filter((g) => g.currency === c.currency)
                  .map((g) => ({ label: g.key, value: formatAmount(g.total, c.currency), share: share(g.total, c.total) }));
                return (
                  <div key={c.currency}>
                    <h3 className="mb-3 text-[15px] font-semibold text-ink-2">
                      {o.byCategory} · {c.currency}
                    </h3>
                    <BarList caption={`${o.byCategory} (${c.currency})`} rows={rows} labelHead={o.categoryHead} valueHead={o.amountHead} />
                  </div>
                );
              })}
            </div>
          ))}
      </Section>

      <Section title={o.lastMonths}>
        {series.status === "loading" && <BlockSkeleton />}
        {series.status === "error" && <BlockError message={series.message} onRetry={reloadSeries} />}
        {series.status === "ok" &&
          (series.data.groups.length === 0 ? (
            <p className="text-[15px] text-muted-foreground">{o.noSpendingMonth}</p>
          ) : (
            <div className="flex flex-col gap-8">
              {Array.from(new Set(series.data.groups.map((g) => g.currency))).map((cur) => {
                const byMonth = new Map(series.data.groups.filter((g) => g.currency === cur).map((g) => [g.key, g.total]));
                const totals = months.map((m) => byMonth.get(m) ?? "0.00");
                const max = totals.reduce((a, b) => (compareAmount(b, a) > 0 ? b : a), "0.00");
                const rows = months.map((m, i) => ({
                  label: monthName(m, lang),
                  value: `${plain(totals[i])} ${cur}`,
                  display: share(totals[i], max) > 0 ? compactAmount(totals[i]) : "",
                  share: share(totals[i], max),
                }));
                return (
                  <div key={cur}>
                    <h3 className="mb-3 text-[15px] font-semibold text-ink-2">{cur}</h3>
                    <Columns caption={`${o.lastMonths} (${cur})`} rows={rows} labelHead={o.monthHead} valueHead={o.amountHead} />
                  </div>
                );
              })}
            </div>
          ))}
      </Section>

      <Section title={o.tasks}>
        {open.status === "loading" && <BlockSkeleton className="h-24" />}
        {open.status === "error" && <BlockError message={open.message} onRetry={reloadOpen} />}
        {open.status === "ok" &&
          (() => {
            const items = open.data.memories.filter((m) => m.kind !== "note");
            const bucket = (b: string) => items.filter((m) => dueBucket(m.due_on, today) === b).length;
            if (items.length === 0) return <p className="text-[15px] text-muted-foreground">{o.noTasks}</p>;
            const stats = [
              { id: "open", label: o.open, n: items.length },
              { id: "overdue", label: o.overdue, n: bucket("overdue") },
              { id: "week", label: o.dueWeek, n: bucket("week") },
            ];
            return (
              <ul className="grid gap-4 sm:grid-cols-3">
                {stats.map((s) => (
                  <li key={s.id}>
                    <Card className="rounded-3xl bg-muted p-6" data-stat={s.id}>
                      <p className="text-[34px] font-bold tracking-[-0.02em] tabular-nums">{s.n}</p>
                      <p className="mt-1 text-sm text-muted-foreground">{s.label}</p>
                    </Card>
                  </li>
                ))}
              </ul>
            );
          })()}
      </Section>

      <Section title={o.notes}>
        {notes.status === "loading" && <BlockSkeleton className="h-24" />}
        {notes.status === "error" && <BlockError message={notes.message} onRetry={reloadNotes} />}
        {notes.status === "ok" &&
          (() => {
            const newest = notes.data.memories
              .filter((m) => m.kind === "note" && m.status !== "archived")
              .sort((a, b) => (a.created_at < b.created_at ? 1 : a.created_at > b.created_at ? -1 : 0))
              .slice(0, 5);
            if (newest.length === 0) return <p className="text-[15px] text-muted-foreground">{o.noNotes}</p>;
            return (
              <ul className="flex flex-col gap-3">
                {newest.map((m) => (
                  <li key={m.id} className="rounded-2xl bg-muted px-5 py-3.5 text-[15px] break-words">
                    {m.title}
                  </li>
                ))}
              </ul>
            );
          })()}
      </Section>

      <Section title={o.categories}>
        {cats.status === "loading" && <BlockSkeleton className="h-12" />}
        {cats.status === "error" && <BlockError message={cats.message} onRetry={reloadCats} />}
        {cats.status === "ok" && (
          <p className="text-[15px] text-muted-foreground">{o.categoriesCount(cats.data.categories.filter((c) => c.name !== "uncategorized").length)}</p>
        )}
      </Section>

      {connectCard}
    </div>
  );
}
