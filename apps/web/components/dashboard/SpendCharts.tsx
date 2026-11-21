"use client";

import { useEffect, useState } from "react";
import { Area, AreaChart, Bar, BarChart, CartesianGrid, Cell, Pie, PieChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { SrTable, type Datum } from "./charts";

// Interactive charts for spending, drawn by Recharts with the theme's --chart-* tokens (no colours in this file).
// Amounts stay strings: `value` is the exact text for tooltips and the screen-reader table; `amount` is a number
// used only for the size of a drawn shape. Drawn charts are hidden from screen readers, the table carries the data.

export type Point = { key?: string; label: string; value: string; amount: number; raw: string; share: number };

const series = (i: number) => `var(--chart-${(i % 6) + 1})`;

function useStillMotion() {
  const [still, setStill] = useState(false);
  useEffect(() => {
    const q = window.matchMedia("(prefers-reduced-motion: reduce)");
    setStill(q.matches);
    const on = (e: MediaQueryListEvent) => setStill(e.matches);
    q.addEventListener("change", on);
    return () => q.removeEventListener("change", on);
  }, []);
  return still;
}

type TipProps = { active?: boolean; payload?: { payload: Point }[] };
function Tip({ active, payload }: TipProps) {
  const p = active ? payload?.[0]?.payload : undefined;
  if (!p) return null;
  return (
    <div className="rounded-2xl border border-hairline bg-popover px-3.5 py-2.5 text-sm shadow-[0_8px_30px_rgba(0,0,0,0.08)]">
      <p className="font-semibold break-words">{p.label}</p>
      <p className="tabular-nums text-ink-2">{p.value}</p>
    </div>
  );
}

const asDatum = (rows: Point[]): Datum[] => rows.map((r) => ({ label: r.label, value: r.value, share: r.share }));

/** Donut of shares, with the total in the middle and a legend that doubles as the key. */
export function Donut({ caption, rows, total, labelHead, valueHead }: { caption: string; rows: Point[]; total: string; labelHead: string; valueHead: string }) {
  const still = useStillMotion();
  return (
    <figure className="grid grid-cols-[minmax(0,1fr)] items-center gap-6 sm:grid-cols-[220px_minmax(0,1fr)]">
      <figcaption className="sr-only">{caption}</figcaption>
      <div aria-hidden="true" className="relative mx-auto size-[220px]">
        <ResponsiveContainer width="100%" height="100%" initialDimension={{ width: 220, height: 220 }}>
          <PieChart>
            <Pie data={rows} dataKey="amount" nameKey="label" innerRadius={72} outerRadius={104} paddingAngle={rows.length > 1 ? 2 : 0} stroke="none" isAnimationActive={!still}>
              {rows.map((_, i) => (
                <Cell key={i} fill={series(i)} />
              ))}
            </Pie>
            <Tooltip content={<Tip />} />
          </PieChart>
        </ResponsiveContainer>
        <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center px-12 text-center">
          <span className="max-w-full text-[17px] leading-tight font-bold tracking-[-0.02em] break-words tabular-nums">{total}</span>
        </div>
      </div>
      <ul aria-hidden="true" className="flex flex-col gap-2.5">
        {rows.map((r, i) => (
          <li key={i} className="grid grid-cols-[auto_minmax(0,1fr)_auto] items-baseline gap-x-3 text-[15px]">
            <span className="size-3 translate-y-[1px] rounded-full" style={{ background: series(i) }} />
            <span className="truncate">{r.label}</span>
            <span className="tabular-nums text-muted-foreground">
              {Math.round(r.share * 100)}% · <span className="font-semibold text-foreground">{r.value}</span>
            </span>
          </li>
        ))}
      </ul>
      <SrTable caption={caption} rows={asDatum(rows)} labelHead={labelHead} valueHead={valueHead} />
    </figure>
  );
}

/** Area over time, one point per period. `compact` gives the short axis labels. */
export function Trend({ caption, rows, compact, labelHead, valueHead }: { caption: string; rows: Point[]; compact: (amount: number) => string; labelHead: string; valueHead: string }) {
  const still = useStillMotion();
  return (
    <figure>
      <figcaption className="sr-only">{caption}</figcaption>
      <div aria-hidden="true" className="h-60 w-full">
        <ResponsiveContainer width="100%" height="100%" initialDimension={{ width: 560, height: 240 }}>
          <AreaChart data={rows} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
            <CartesianGrid vertical={false} stroke="var(--hairline)" />
            <XAxis dataKey="label" tickLine={false} axisLine={false} tick={{ fill: "var(--muted-foreground)", fontSize: 12 }} interval="preserveStartEnd" minTickGap={16} />
            <YAxis width={44} tickLine={false} axisLine={false} tick={{ fill: "var(--muted-foreground)", fontSize: 12 }} tickFormatter={compact} />
            <Tooltip content={<Tip />} cursor={{ stroke: "var(--border)" }} />
            <Area type="monotone" dataKey="amount" stroke="var(--chart-1)" strokeWidth={2.5} fill="var(--chart-1)" fillOpacity={0.12} dot={{ r: 3, fill: "var(--chart-1)", strokeWidth: 0 }} activeDot={{ r: 5 }} isAnimationActive={!still} />
          </AreaChart>
        </ResponsiveContainer>
      </div>
      <SrTable caption={caption} rows={asDatum(rows)} labelHead={labelHead} valueHead={valueHead} />
    </figure>
  );
}

/** Horizontal ranked bars, for merchants and other long label lists. */
export function Ranked({ caption, rows, labelHead, valueHead }: { caption: string; rows: Point[]; labelHead: string; valueHead: string }) {
  const still = useStillMotion();
  const height = Math.max(120, rows.length * 40 + 8);
  return (
    <figure>
      <figcaption className="sr-only">{caption}</figcaption>
      <div aria-hidden="true" style={{ height }} className="w-full">
        <ResponsiveContainer width="100%" height="100%" initialDimension={{ width: 560, height }}>
          <BarChart data={rows} layout="vertical" margin={{ top: 0, right: 8, bottom: 0, left: 0 }} barCategoryGap={8}>
            <XAxis type="number" hide />
            <YAxis type="category" dataKey="label" width={120} tickLine={false} axisLine={false} tick={{ fill: "var(--ink-2)", fontSize: 13 }} />
            <Tooltip content={<Tip />} cursor={{ fill: "var(--muted)" }} />
            <Bar dataKey="amount" radius={[0, 10, 10, 0]} isAnimationActive={!still}>
              {rows.map((_, i) => (
                <Cell key={i} fill={series(i)} />
              ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </div>
      <SrTable caption={caption} rows={asDatum(rows)} labelHead={labelHead} valueHead={valueHead} />
    </figure>
  );
}

/** A figure for one KPI tile. */
export function Kpi({ label, value, hint }: { label: string; value: string; hint?: string }) {
  return (
    <div className="min-w-0 rounded-3xl bg-muted p-5">
      <p className="text-sm text-muted-foreground">{label}</p>
      <p className="mt-1 text-[22px] leading-tight font-bold tracking-[-0.02em] break-words tabular-nums">{value}</p>
      {hint && <p className="mt-1 truncate text-sm text-ink-2">{hint}</p>}
    </div>
  );
}
