"use client";

import { Card } from "@/components/ui/card";
import { compareAmount, formatAmount, formatDate, share, sumAmount } from "@/lib/format";
import { movers, pace, type Group, type Recurring, type Tx } from "@/lib/monitor";
import { useDash } from "./context";
import { BlockSkeleton } from "./states";

// One card per currency: how the month is going against the same days last month, where it is heading, what changed,
// what looks unusual and what looks recurring. Everything is an estimate from the user's own records and is worded
// that way; amounts of different currencies are never mixed, so each card is one currency. Direction is always in
// words, never colour alone.

type Props = {
  currency: string;
  today: string;
  /** This month's and the same days last month's groups by category (any currency; this card filters its own). */
  month: Group[];
  prev: Group[];
  monthTotal: string;
  prevTotal: string;
  /** Last month's full total from the 6-month series, or null when that is not known. */
  prevFull: string | null;
  /** Already limited to this currency and to a few rows. */
  unusual: Tx[];
  recurring: Recurring[];
  recentLoading: boolean;
};

function Block({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="mt-6 border-t border-hairline pt-5">
      <h4 className="text-[15px] font-semibold text-ink-2">{title}</h4>
      {children}
    </div>
  );
}

export function MonthCheck({ currency, today, month, prev, monthTotal, prevTotal, prevFull, unusual, recurring, recentLoading }: Props) {
  const { lang, copy } = useDash();
  const c = copy.overview.check;
  const money = (a: string) => formatAmount(a, currency);
  const zero = "0.00";
  const p = pace({ today, spent: monthTotal, prevSame: prevTotal, prevFull });
  const rises = movers(month, prev, currency);
  const hadLast = compareAmount(prevTotal, zero) > 0;
  const comparison = !hadLast
    ? c.noLast
    : p.direction === "same" || p.percent === 0
      ? c.same
      : p.percent === null
        ? c.noLast
        : p.direction === "up"
          ? c.up(p.percent)
          : c.down(p.percent);
  const biggest = compareAmount(monthTotal, prevTotal) >= 0 ? monthTotal : prevTotal;
  const bars = [
    { label: c.thisMonthBar, amount: monthTotal, fill: "var(--chart-1)" },
    { label: c.lastMonthBar, amount: prevTotal, fill: "var(--chart-6)" },
  ];

  return (
    <div>
      <h3 className="mb-3 text-[15px] font-semibold text-ink-2">{currency}</h3>
      <Card className="rounded-[28px] border border-hairline bg-card p-5 md:p-7">
        <p className="text-[28px] leading-tight font-bold tracking-[-0.02em] break-words tabular-nums">{c.soFar(money(monthTotal))}</p>
        <p className="mt-1 text-[15px] text-ink-2">{comparison}</p>

        <div className="mt-5 flex flex-col gap-3.5">
          {bars.map((b) => (
            <div key={b.label}>
              <div className="flex items-baseline justify-between gap-3 text-sm">
                <span className="min-w-0 text-ink-2">{b.label}</span>
                <span className="font-semibold tabular-nums">{money(b.amount)}</span>
              </div>
              <div aria-hidden="true" className="mt-1.5 h-2.5 w-full rounded-full bg-muted">
                <div className="h-full rounded-full" style={{ width: `${Math.round(share(b.amount, biggest) * 100)}%`, background: b.fill }} />
              </div>
            </div>
          ))}
        </div>

        {compareAmount(monthTotal, zero) > 0 && (
          <div className="mt-5">
            {p.projected ? (
              <>
                <p className="text-[17px] font-semibold break-words tabular-nums">{c.projected(money(p.projected))}</p>
                <p className="mt-0.5 text-sm text-muted-foreground">{c.projectedNote}</p>
              </>
            ) : (
              <p className="text-sm text-muted-foreground">{c.tooEarly}</p>
            )}
          </div>
        )}
        {prevFull !== null && compareAmount(prevFull, zero) > 0 && <p className="mt-2 text-sm text-ink-2 tabular-nums">{c.lastMonthTotal(money(prevFull))}</p>}

        {rises.length > 0 && (
          <Block title={c.changedTitle}>
            <ul className="mt-2 divide-y divide-hairline">
              {rises.map((m) => (
                <li key={m.category} className="py-2.5 text-[15px] break-words">
                  {c.changedLine(m.category, money(m.delta))}
                  {m.isNew && <span className="mt-0.5 block text-sm text-muted-foreground">{c.isNew}</span>}
                </li>
              ))}
            </ul>
          </Block>
        )}

        {recentLoading && <BlockSkeleton className="mt-6 h-24" />}

        {unusual.length > 0 && (
          <Block title={c.unusualTitle}>
            <ul className="mt-2 divide-y divide-hairline">
              {unusual.map((t) => (
                <li key={t.id} className="grid grid-cols-[minmax(0,1fr)_auto] items-start gap-x-4 py-2.5">
                  <div className="min-w-0">
                    <p className="font-semibold break-words">{t.merchant}</p>
                    <p className="text-sm break-words text-muted-foreground">
                      {formatDate(t.occurred_on, lang)} · {c.unusualHint(t.category)}
                    </p>
                  </div>
                  <p className="font-semibold tabular-nums">{money(t.amount)}</p>
                </li>
              ))}
            </ul>
          </Block>
        )}

        {recurring.length > 0 && (
          <Block title={c.recurringTitle}>
            <ul className="mt-2 divide-y divide-hairline">
              {recurring.map((r) => (
                <li key={r.merchant} className="grid grid-cols-[minmax(0,1fr)_auto] items-start gap-x-4 py-2.5">
                  <div className="min-w-0">
                    <p className="font-semibold break-words">{r.merchant}</p>
                    <p className="text-sm break-words text-muted-foreground">
                      {c.usuallyDay(r.day)}
                      {!r.seenThisMonth && ` · ${c.notYet}`}
                    </p>
                  </div>
                  <p className="font-semibold tabular-nums">{money(r.amount)}</p>
                </li>
              ))}
            </ul>
            <p className="mt-3 text-sm text-ink-2 tabular-nums">{c.recurringTotal(money(sumAmount(recurring.map((r) => r.amount))))}</p>
          </Block>
        )}
      </Card>
    </div>
  );
}
