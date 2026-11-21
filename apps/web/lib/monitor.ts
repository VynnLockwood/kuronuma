// The arithmetic behind the dashboard's "Month check". Pure functions, no React. Money stays a string and is
// handled in whole minor units (see format.ts); amounts in different currencies are never combined, so the callers
// pass one currency at a time or these functions split by currency themselves. Everything here is an estimate drawn
// from the user's own records and is worded that way on screen.

import { compareAmount, diffAmount, lastMonths, percentOf, scaleAmount, sumAmount } from "./format";

const pad = (n: number) => String(n).padStart(2, "0");

/** Days in a YYYY-MM month. */
export function daysInMonth(ym: string): number {
  const [y, m] = ym.split("-").map(Number);
  return new Date(Date.UTC(y, m, 0)).getUTCDate();
}

/**
 * The same days last month: the first of the previous month up to today's day of month, clamped to that month's
 * length (31 March compares with 1 to 28 February). Dates are plain YYYY-MM-DD strings.
 */
export function samePeriod(today: string): { prevFrom: string; prevTo: string } {
  const prev = lastMonths(today, 2)[0];
  const day = Math.min(Number(today.slice(8, 10)), daysInMonth(prev));
  return { prevFrom: `${prev}-01`, prevTo: `${prev}-${pad(day)}` };
}

export type Pace = {
  spent: string;
  prevSame: string;
  prevFull: string | null;
  /** spent minus prevSame, signed. */
  delta: string;
  /** Whole percent, always positive (the direction carries the sign); null when last month's same days were zero. */
  percent: number | null;
  direction: "up" | "down" | "same";
  /** Estimate for the whole month from the pace so far; null before the 5th, when a few days say too little. */
  projected: string | null;
};

const PROJECT_FROM_DAY = 5;

export function pace(i: { today: string; spent: string; prevSame: string; prevFull: string | null }): Pace {
  const day = Number(i.today.slice(8, 10));
  const cmp = compareAmount(i.spent, i.prevSame);
  const delta = diffAmount(i.spent, i.prevSame);
  const pct = percentOf(delta, i.prevSame);
  return {
    spent: i.spent,
    prevSame: i.prevSame,
    prevFull: i.prevFull,
    delta,
    percent: pct === null ? null : Math.abs(pct),
    direction: cmp > 0 ? "up" : cmp < 0 ? "down" : "same",
    projected: day >= PROJECT_FROM_DAY ? scaleAmount(i.spent, daysInMonth(i.today.slice(0, 7)), day) : null,
  };
}

export type Group = { key: string; currency: string; total: string; count: number };
export type Tx = { id: string; amount: string; currency: string; merchant: string; category: string; occurred_on: string };

export type Mover = { category: string; now: string; before: string; delta: string; isNew: boolean };

/**
 * The categories of one currency that grew most against the same days last month, largest rise first. A category
 * with nothing last time is marked new. Categories that fell or stayed level are not listed.
 */
export function movers(now: Group[], before: Group[], currency: string, limit = 3): Mover[] {
  const prev = new Map(before.filter((g) => g.currency === currency).map((g) => [g.key, g.total]));
  return now
    .filter((g) => g.currency === currency)
    .map((g) => {
      const b = prev.get(g.key) ?? "0.00";
      return { category: g.key, now: g.total, before: b, delta: diffAmount(g.total, b), isNew: compareAmount(b, "0.00") === 0 };
    })
    .filter((m) => compareAmount(m.delta, "0.00") > 0)
    .sort((a, b) => compareAmount(b.delta, a.delta) || a.category.localeCompare(b.category))
    .slice(0, limit);
}

const bySize = (a: string, b: string) => compareAmount(a, b);

/** Median of plain amounts; an even count averages the two middle values (rounded toward zero in minor units). */
function median(amounts: string[]): string {
  const s = [...amounts].sort(bySize);
  const mid = s.length >> 1;
  return s.length % 2 ? s[mid] : scaleAmount(sumAmount([s[mid - 1], s[mid]]), 1, 2);
}

const UNUSUAL_FACTOR = 3;
const UNUSUAL_MIN_HISTORY = 3;

/**
 * Records this month that are at least three times the median of the same category's records in the previous three
 * months (and only when there are at least three such records), per currency. Largest first.
 */
export function unusual(txs: Tx[], today: string, limit = 3): Tx[] {
  const month = today.slice(0, 7);
  const window = lastMonths(today, 4).slice(0, 3);
  const key = (t: Tx) => `${t.currency}\u0000${t.category}`;
  const history = new Map<string, string[]>();
  for (const t of txs) {
    if (!window.includes(t.occurred_on.slice(0, 7))) continue;
    (history.get(key(t)) ?? history.set(key(t), []).get(key(t))!).push(t.amount);
  }
  return txs
    .filter((t) => t.occurred_on.slice(0, 7) === month && t.occurred_on <= today)
    .filter((t) => {
      const h = history.get(key(t));
      return !!h && h.length >= UNUSUAL_MIN_HISTORY && compareAmount(t.amount, scaleAmount(median(h), UNUSUAL_FACTOR, 1)) >= 0;
    })
    .sort((a, b) => bySize(b.amount, a.amount) || a.id.localeCompare(b.id))
    .slice(0, limit);
}

export type Recurring = { merchant: string; currency: string; amount: string; day: number; seenThisMonth: boolean };

/**
 * Merchants charged in each of the last two full months, in one currency, with the two amounts within 10% of the
 * larger one. Shows the latest amount and day of month, and whether the merchant has appeared this month. Largest
 * first.
 */
export function recurring(txs: Tx[], today: string, limit = 5): Recurring[] {
  const [older, newer, current] = lastMonths(today, 3);
  const latest = new Map<string, Map<string, Tx>>(); // key -> month -> the month's latest record
  for (const t of txs) {
    const month = t.occurred_on.slice(0, 7);
    if (month !== older && month !== newer && month !== current) continue;
    const k = `${t.currency}\u0000${t.merchant.trim().toLowerCase()}`;
    const months = latest.get(k) ?? latest.set(k, new Map()).get(k)!;
    const seen = months.get(month);
    if (!seen || t.occurred_on > seen.occurred_on || (t.occurred_on === seen.occurred_on && t.id > seen.id)) months.set(month, t);
  }
  const out: Recurring[] = [];
  for (const months of latest.values()) {
    const a = months.get(newer);
    const b = months.get(older);
    if (!a || !b) continue;
    const [big, small] = compareAmount(a.amount, b.amount) >= 0 ? [a.amount, b.amount] : [b.amount, a.amount];
    if (compareAmount(scaleAmount(diffAmount(big, small), 10, 1), big) > 0) continue;
    out.push({ merchant: a.merchant.trim(), currency: a.currency, amount: a.amount, day: Number(a.occurred_on.slice(8, 10)), seenThisMonth: months.has(current) });
  }
  return out.sort((x, y) => bySize(y.amount, x.amount) || x.merchant.localeCompare(y.merchant)).slice(0, limit);
}
