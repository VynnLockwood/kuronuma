// The arithmetic behind the dashboard's "Month check". Pure functions, no React. Money stays a string and is
// handled in whole minor units (see format.ts); amounts in different currencies are never combined, so the callers
// pass one currency at a time or these functions split by currency themselves. Everything here is an estimate drawn
// from the user's own records and is worded that way on screen.

import { compareAmount, diffAmount, lastMonths, percentOf, scaleAmount } from "./format";

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
