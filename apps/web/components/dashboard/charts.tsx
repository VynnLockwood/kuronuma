import type { ReactNode } from "react";

// Two small charts drawn with CSS and the theme's tokens: no chart library. Each one has a table with the same
// numbers for screen readers; the drawn bars are hidden from them.

/** `value` is the full text (it goes in the table for screen readers); `display` is an optional shorter label for the drawn chart. */
export type Datum = { label: string; value: string; share: number; display?: string };

const pct = (share: number) => `${Math.max(0, Math.min(1, share)) * 100}%`;

function SrTable({ caption, rows, labelHead, valueHead }: { caption: string; rows: Datum[]; labelHead: string; valueHead: string }) {
  return (
    <table className="sr-only">
      <caption>{caption}</caption>
      <thead>
        <tr>
          <th scope="col">{labelHead}</th>
          <th scope="col">{valueHead}</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((r, i) => (
          <tr key={i}>
            <th scope="row">{r.label}</th>
            <td>{r.value}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

/** Horizontal bars, one row per item: label, a bar as wide as its share, and the value. */
export function BarList({
  caption,
  rows,
  labelHead,
  valueHead,
  empty,
}: {
  caption: string;
  rows: Datum[];
  labelHead: string;
  valueHead: string;
  empty?: ReactNode;
}) {
  if (rows.length === 0) return <>{empty ?? null}</>;
  return (
    <figure>
      <figcaption className="sr-only">{caption}</figcaption>
      <ul aria-hidden="true" className="flex flex-col gap-3">
        {rows.map((r, i) => (
          <li key={i} className="grid grid-cols-[minmax(0,1fr)_auto] items-baseline gap-x-4 gap-y-1">
            <span className="truncate text-[15px]">{r.label}</span>
            <span className="text-[15px] font-semibold tabular-nums">{r.value}</span>
            <span className="col-span-2 h-2 rounded-full bg-hairline">
              <span className="block h-2 min-w-1 rounded-full bg-primary" style={{ width: pct(r.share) }} />
            </span>
          </li>
        ))}
      </ul>
      <SrTable caption={caption} rows={rows} labelHead={labelHead} valueHead={valueHead} />
    </figure>
  );
}

/** Vertical columns, one per item, for a short series such as the last six months. */
export function Columns({
  caption,
  rows,
  labelHead,
  valueHead,
}: {
  caption: string;
  rows: Datum[];
  labelHead: string;
  valueHead: string;
}) {
  return (
    <figure>
      <figcaption className="sr-only">{caption}</figcaption>
      <div aria-hidden="true" className="grid items-end gap-2" style={{ gridTemplateColumns: `repeat(${Math.max(rows.length, 1)}, minmax(0, 1fr))` }}>
        {rows.map((r, i) => (
          <div key={i} className="flex min-w-0 flex-col items-center gap-1.5">
            <span className="h-4 max-w-full truncate text-[11px] font-semibold tabular-nums text-ink-2">{r.display ?? r.value}</span>
            <span className="flex h-28 w-full items-end rounded-xl bg-muted">
              <span className="block w-full rounded-xl bg-primary" style={{ height: pct(r.share), minHeight: r.share > 0 ? 4 : 0 }} />
            </span>
            <span className="max-w-full truncate text-[11px] text-muted-foreground">{r.label}</span>
          </div>
        ))}
      </div>
      <SrTable caption={caption} rows={rows} labelHead={labelHead} valueHead={valueHead} />
    </figure>
  );
}
