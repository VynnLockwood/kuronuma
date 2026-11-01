// Formatting for the dashboard. Money stays a string from the server to the screen: it is grouped and compared as
// text and whole minor units (BigInt), never through floating point. Dates are shown with the Gregorian year in
// both languages and no Intl, so server and browser always agree.

import type { Lang } from "./web";

const monthsEn = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
const monthsTh = ["ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.", "ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค."];
const months = (lang: Lang) => (lang === "th" ? monthsTh : monthsEn);

const amountRe = /^(-?)(\d+)(?:\.(\d+))?$/;

/** "1234.5" and "THB" become "1,234.5 THB". An unexpected value is shown as it came. */
export function formatAmount(amount: string, currency: string): string {
  const m = amountRe.exec(amount);
  if (!m) return `${amount} ${currency}`;
  const grouped = m[2].replace(/\B(?=(\d{3})+(?!\d))/g, ",");
  return `${m[1]}${grouped}${m[3] !== undefined ? "." + m[3] : ""} ${currency}`;
}

/** Whole minor units (hundredths) of a plain amount string, or null. */
function minor(amount: string): bigint | null {
  const m = /^(-?)(\d+)(?:\.(\d{1,2}))?$/.exec(amount);
  if (!m) return null;
  const v = BigInt(m[2]) * 100n + BigInt((m[3] ?? "").padEnd(2, "0") || "0");
  return m[1] ? -v : v;
}

/** part / whole as a number from 0 to 1, for the width of a bar. Bad input or a zero whole gives 0. */
export function share(part: string, whole: string): number {
  const a = minor(part);
  const b = minor(whole);
  if (a === null || b === null || b <= 0n || a <= 0n) return 0;
  const ratio = Number((a * 10000n) / b) / 10000;
  return Math.min(1, ratio);
}

/** Groups rows by their currency. Totals in different currencies are never added together. */
export function groupByCurrency<T extends { currency: string }>(rows: T[]): Record<string, T[]> {
  const out: Record<string, T[]> = {};
  for (const r of rows) (out[r.currency] ??= []).push(r);
  return out;
}

/** "2026-10-28" becomes "28 Oct 2026" (Thai: "28 ต.ค. 2026"). */
export function formatDate(iso: string, lang: Lang): string {
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(iso);
  if (!m || Number(m[2]) < 1 || Number(m[2]) > 12) return iso;
  return `${Number(m[3])} ${months(lang)[Number(m[2]) - 1]} ${m[1]}`;
}

/** "2026-10" becomes "Oct 2026". */
export function formatMonth(ym: string, lang: Lang): string {
  const m = /^(\d{4})-(\d{2})$/.exec(ym);
  if (!m || Number(m[2]) < 1 || Number(m[2]) > 12) return ym;
  return `${months(lang)[Number(m[2]) - 1]} ${m[1]}`;
}

const pad = (n: number) => String(n).padStart(2, "0");

/** The date on the visitor's own calendar (not the UTC date), as YYYY-MM-DD. */
export function localISODate(d: Date): string {
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

/** Adds days to a YYYY-MM-DD date. Dates are compared as dates, never as instants. */
export function addDaysISO(iso: string, days: number): string {
  const [y, m, d] = iso.split("-").map(Number);
  const t = new Date(Date.UTC(y, m - 1, d + days));
  return `${t.getUTCFullYear()}-${pad(t.getUTCMonth() + 1)}-${pad(t.getUTCDate())}`;
}

export const monthStartISO = (iso: string) => `${iso.slice(0, 7)}-01`;

/** The last n months ending with the month of iso, oldest first, as YYYY-MM. */
export function lastMonths(iso: string, n: number): string[] {
  const [y, m] = iso.split("-").map(Number);
  const out: string[] = [];
  for (let i = n - 1; i >= 0; i--) {
    const t = new Date(Date.UTC(y, m - 1 - i, 1));
    out.push(`${t.getUTCFullYear()}-${pad(t.getUTCMonth() + 1)}`);
  }
  return out;
}

/**
 * Where a due date falls relative to today: "overdue" (before today), "week" (today to six days ahead), "later",
 * or "none" (no usable date). Dates are compared as dates, so midnight and time zones cannot shift the result.
 * The Overview and the Tasks tab both use this one rule.
 */
export function dueBucket(due: string | undefined, today: string): "overdue" | "week" | "later" | "none" {
  if (!due || !/^\d{4}-\d{2}-\d{2}$/.test(due)) return "none";
  if (due < today) return "overdue";
  return due <= addDaysISO(today, 6) ? "week" : "later";
}

/** Compares two plain amount strings in whole minor units, without floating point. Unusable input counts as zero. */
export function compareAmount(a: string, b: string): number {
  const d = (minor(a) ?? 0n) - (minor(b) ?? 0n);
  return d > 0n ? 1 : d < 0n ? -1 : 0;
}

/** "2026-05" becomes "May" (Thai: "พ.ค."), for the label under a column. */
export function monthName(ym: string, lang: Lang): string {
  const m = /^(\d{4})-(\d{2})$/.exec(ym);
  if (!m || Number(m[2]) < 1 || Number(m[2]) > 12) return ym;
  return months(lang)[Number(m[2]) - 1];
}
