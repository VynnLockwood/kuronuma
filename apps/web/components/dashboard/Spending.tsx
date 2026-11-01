"use client";

import Link from "next/link";
import { useEffect, useMemo, useRef, useState } from "react";
import { Spot } from "@/components/motifs/Spot";
import { Alert } from "@/components/ui/alert";
import { Button, buttonVariants } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { callTool, type Lang } from "@/lib/web";
import { formatAmount, formatDate, formatMonth, lastMonths, localISODate, share } from "@/lib/format";
import { BarList } from "./charts";
import { Confirm } from "./Confirm";
import { useDash } from "./context";
import { BlockError, BlockSkeleton } from "./states";
import { useTool } from "./useTool";

type Tx = { id: string; amount: string; currency: string; merchant: string; category: string; occurred_on: string; note?: string };
type TxList = { count: number; transactions: Tx[] };
type Group = { count: number; currency: string; key: string; total: string };
type Summary = { groups: Group[]; totals_by_currency: { count: number; currency: string; total: string }[]; period?: { from: string; to: string } };
type Categories = { categories: { name: string }[] };
type GroupBy = "category" | "month" | "merchant";

const LIMIT = 200;
const groupLabel = (g: GroupBy, key: string, lang: Lang) => (g === "month" ? formatMonth(key, lang) : key);

function Field({ id, label, hint, children }: { id: string; label: string; hint?: string; children: React.ReactNode }) {
  return (
    <div className="flex min-w-0 flex-col gap-1.5">
      <label htmlFor={id} className="text-sm font-semibold text-ink-2">
        {label}
      </label>
      {children}
      {hint && <p className="text-xs text-muted-foreground">{hint}</p>}
    </div>
  );
}

const selectClass =
  "h-12 w-full min-w-0 rounded-2xl border border-input bg-background px-4 text-[15px] text-foreground focus-visible:border-ring";

export function Spending() {
  const { lang, copy, onSignedOut } = useDash();
  const c = copy.spending;
  const today = useMemo(() => localISODate(new Date()), []);

  // Filters and the summary grouping.
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [category, setCategory] = useState("");
  const [group, setGroup] = useState<GroupBy>("category");

  const listArgs = { ...(from && { from }), ...(to && { to }), ...(category && { category }), limit: LIMIT };
  const [list, reloadList] = useTool<TxList>("list_transactions", listArgs);
  // Without dates the tool covers this month. Grouped by month that would be one bar, so it asks for 12 months.
  const wide = group === "month" && !from && !to;
  const [summary, reloadSummary] = useTool<Summary>("get_spending_summary", {
    group_by: group,
    ...(wide ? { from: `${lastMonths(today, 12)[0]}-01`, to: today } : { ...(from && { from }), ...(to && { to }) }),
  });
  const [cats, reloadCats] = useTool<Categories>("list_categories", {});
  const names = cats.status === "ok" ? cats.data.categories.map((x) => x.name) : [];
  const refresh = () => {
    reloadList();
    reloadSummary();
    reloadCats();
  };

  // The add form.
  const [amount, setAmount] = useState("");
  const [currency, setCurrency] = useState("THB");
  const [merchant, setMerchant] = useState("");
  const [date, setDate] = useState(today);
  const [cat, setCat] = useState("");
  const [note, setNote] = useState("");
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ kind: "ok" | "info" | "error"; text: string } | null>(null);
  const [dup, setDup] = useState<Record<string, unknown> | null>(null);

  async function save(allowDuplicate: boolean) {
    const args: Record<string, unknown> = {
      amount: amount.trim(),
      currency: currency.trim().toUpperCase(),
      merchant: merchant.trim(),
      occurred_on: date,
      ...(cat.trim() && { category: cat.trim() }),
      ...(note.trim() && { note: note.trim() }),
      ...(allowDuplicate && { allow_duplicate: true }),
    };
    setSaving(true);
    setMessage(null);
    const r = await callTool<{ status?: string; category_created?: boolean; transaction?: Tx }>("add_transaction", args, lang);
    setSaving(false);
    setDup(null);
    if (!r.ok) {
      if (r.kind === "signed-out") onSignedOut();
      else setMessage({ kind: "error", text: r.message }); // the typing stays
      return;
    }
    const status = r.data.status;
    if (status === "possible_duplicate") {
      setDup(args);
      return;
    }
    if (status === "already_saved") {
      setMessage({ kind: "info", text: c.add.alreadySaved });
    } else {
      setMessage({ kind: "ok", text: r.data.category_created ? c.add.savedNewCategory(r.data.transaction?.category ?? String(args.category ?? "")) : c.add.saved });
    }
    // Everything but the currency and the date is cleared: a category left behind could silently file the next,
    // unrelated record in the wrong place.
    setAmount("");
    setMerchant("");
    setCat("");
    setNote("");
    refresh();
  }

  // Delete, after a confirmation.
  const [toDelete, setToDelete] = useState<Tx | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const recordsHeading = useRef<HTMLHeadingElement>(null);
  async function confirmDelete() {
    if (!toDelete) return;
    setDeleting(true);
    const r = await callTool("delete_transaction", { id: toDelete.id }, lang);
    setDeleting(false);
    setToDelete(null);
    if (!r.ok) {
      if (r.kind === "signed-out") onSignedOut();
      else setDeleteError(r.message);
      return;
    }
    setDeleteError(null);
    focusAfterDelete.current = true; // the row is about to go; see the effect below
    refresh();
  }
  // Focus moves only after the dialog has closed (it returns focus to the Delete button, which is then removed),
  // so a keyboard user lands on the Records heading instead of the page body.
  const focusAfterDelete = useRef(false);
  useEffect(() => {
    if (toDelete === null && focusAfterDelete.current) {
      focusAfterDelete.current = false;
      recordsHeading.current?.focus();
    }
  }, [toDelete]);

  const filtered = Boolean(from || to || category);

  return (
    <div className="flex flex-col gap-12">
      {/* Totals */}
      <section aria-labelledby="sp-totals">
        <div className="mb-4 flex flex-wrap items-end justify-between gap-4">
          <h2 id="sp-totals" className="text-[22px] font-bold tracking-[-0.02em]">
            {c.totals}
          </h2>
          <fieldset className="min-w-0">
            <legend className="sr-only">{c.groupBy}</legend>
            <div className="flex rounded-full bg-muted p-1 text-sm">
              {(["category", "month", "merchant"] as const).map((g) => (
                <label key={g} className="cursor-pointer">
                  <input type="radio" name="sp-group" value={g} checked={group === g} onChange={() => setGroup(g)} className="peer sr-only" />
                  <span className="flex h-10 items-center rounded-full px-4 font-medium text-muted-foreground peer-checked:bg-background peer-checked:text-foreground peer-checked:shadow-[0_1px_3px_rgba(0,0,0,0.1)] peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2 peer-focus-visible:outline-ring">
                    {c.groups[g]}
                  </span>
                </label>
              ))}
            </div>
          </fieldset>
        </div>
        {summary.status === "loading" && <BlockSkeleton />}
        {summary.status === "error" && <BlockError message={summary.message} onRetry={reloadSummary} />}
        {summary.status === "ok" && (
          <p className="mb-6 text-sm text-muted-foreground">
            {summary.data.period && <span className="font-medium text-ink-2">{c.period(formatDate(summary.data.period.from, lang), formatDate(summary.data.period.to, lang))}. </span>}
            {c.totalsNote}
          </p>
        )}
        {summary.status === "ok" &&
          (summary.data.groups.length === 0 ? (
            <p className="text-[15px] text-muted-foreground">{c.noTotals}</p>
          ) : (
            <div className="flex flex-col gap-8">
              {summary.data.totals_by_currency.map((tot) => {
                const rows = summary.data.groups.filter((g) => g.currency === tot.currency);
                return (
                  <div key={tot.currency} className="flex flex-col gap-4">
                    <h3 className="text-[15px] font-semibold text-ink-2">{formatAmount(tot.total, tot.currency)}</h3>
                    <BarList
                      table={false}
                      caption={`${c.totals} (${tot.currency})`}
                      labelHead={c.groups[group]}
                      valueHead={c.totalHead}
                      rows={rows.map((g) => ({ label: groupLabel(group, g.key, lang), value: formatAmount(g.total, tot.currency), share: share(g.total, tot.total) }))}
                    />
                    <div className="overflow-x-auto rounded-2xl border border-hairline">
                      <table className="w-full min-w-[320px] text-left text-sm">
                        <caption className="sr-only">{`${c.totals} (${tot.currency})`}</caption>
                        <thead className="bg-muted text-muted-foreground">
                          <tr>
                            <th scope="col" className="px-4 py-2.5 font-semibold">
                              {c.groups[group]}
                            </th>
                            <th scope="col" className="px-4 py-2.5 text-right font-semibold">
                              {c.recordsHead}
                            </th>
                            <th scope="col" className="px-4 py-2.5 text-right font-semibold">
                              {c.totalHead}
                            </th>
                          </tr>
                        </thead>
                        <tbody>
                          {rows.map((g) => (
                            <tr key={g.key} className="border-t border-hairline">
                              <th scope="row" className="px-4 py-2.5 font-medium break-words">
                                {groupLabel(group, g.key, lang)}
                              </th>
                              <td className="px-4 py-2.5 text-right tabular-nums">{g.count}</td>
                              <td className="px-4 py-2.5 text-right tabular-nums">{formatAmount(g.total, tot.currency)}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                );
              })}
            </div>
          ))}
      </section>

      {/* Add a record */}
      <section aria-labelledby="sp-add">
        <h2 id="sp-add" className="mb-4 text-[22px] font-bold tracking-[-0.02em]">
          {c.add.title}
        </h2>
        <Card className="rounded-[28px] bg-muted p-6 md:p-8">
          <form
            onSubmit={(e) => {
              e.preventDefault();
              if (!saving) void save(false);
            }}
            className="grid grid-cols-[minmax(0,1fr)] gap-4 md:grid-cols-6"
          >
            <div className="md:col-span-2">
              <Field id="sp-amount" label={c.add.amount}>
                <Input id="sp-amount" inputMode="decimal" autoComplete="off" value={amount} onChange={(e) => setAmount(e.target.value)} />
              </Field>
            </div>
            <div className="md:col-span-1">
              <Field id="sp-currency" label={c.add.currency}>
                <Input id="sp-currency" autoComplete="off" maxLength={3} className="uppercase" value={currency} onChange={(e) => setCurrency(e.target.value)} />
              </Field>
            </div>
            <div className="md:col-span-3">
              <Field id="sp-merchant" label={c.add.merchant}>
                <Input id="sp-merchant" autoComplete="off" maxLength={100} value={merchant} onChange={(e) => setMerchant(e.target.value)} />
              </Field>
            </div>
            <div className="md:col-span-2">
              <Field id="sp-date" label={c.add.date}>
                <Input id="sp-date" type="date" value={date} onChange={(e) => setDate(e.target.value)} />
              </Field>
            </div>
            <div className="md:col-span-2">
              <Field id="sp-cat" label={c.add.category} hint={c.add.categoryHint}>
                <Input id="sp-cat" list="sp-cat-list" autoComplete="off" maxLength={40} value={cat} onChange={(e) => setCat(e.target.value)} />
                <datalist id="sp-cat-list">
                  {names.map((n) => (
                    <option key={n} value={n} />
                  ))}
                </datalist>
              </Field>
            </div>
            <div className="md:col-span-2">
              <Field id="sp-note" label={c.add.note}>
                <Input id="sp-note" autoComplete="off" maxLength={300} value={note} onChange={(e) => setNote(e.target.value)} />
              </Field>
            </div>
            <div className="flex flex-wrap items-center gap-4 md:col-span-6">
              <Button type="submit" size="lg" disabled={saving}>
                {saving ? c.add.saving : c.add.submit}
              </Button>
              {message && (
                <p role={message.kind === "error" ? "alert" : "status"} className={message.kind === "error" ? "text-sm font-medium text-destructive" : "text-sm font-medium text-success"}>
                  {message.text}
                </p>
              )}
            </div>
          </form>
        </Card>
      </section>

      {/* Records */}
      <section aria-labelledby="sp-records">
        <h2 id="sp-records" ref={recordsHeading} tabIndex={-1} className="mb-4 text-[22px] font-bold tracking-[-0.02em] outline-none">
          {c.records}
        </h2>
        <div className="mb-6 grid grid-cols-[minmax(0,1fr)] gap-4 sm:grid-cols-3">
          <Field id="sp-from" label={c.from}>
            <Input id="sp-from" type="date" value={from} onChange={(e) => setFrom(e.target.value)} />
          </Field>
          <Field id="sp-to" label={c.to}>
            <Input id="sp-to" type="date" value={to} onChange={(e) => setTo(e.target.value)} />
          </Field>
          <Field id="sp-filter-cat" label={c.category}>
            <select id="sp-filter-cat" className={selectClass} value={category} onChange={(e) => setCategory(e.target.value)}>
              <option value="">{c.allCategories}</option>
              {names.map((n) => (
                <option key={n} value={n}>
                  {n}
                </option>
              ))}
            </select>
          </Field>
        </div>
        {filtered && (
          <Button
            variant="ghost"
            size="sm"
            className="mb-4 -mt-2"
            onClick={() => {
              setFrom("");
              setTo("");
              setCategory("");
            }}
          >
            {c.clearFilters}
          </Button>
        )}
        {deleteError && (
          <Alert variant="destructive" className="mb-4">
            <div>
              <p className="font-semibold">{c.deleteFailed}</p>
              <p className="mt-0.5 break-words opacity-90">{deleteError}</p>
            </div>
          </Alert>
        )}
        {list.status === "loading" && <BlockSkeleton className="h-48" />}
        {list.status === "error" && <BlockError message={list.message} onRetry={reloadList} />}
        {list.status === "ok" &&
          (list.data.transactions.length === 0 ? (
            <div className="rounded-[28px] bg-muted p-8 text-center">
              <Spot motif={filtered ? "mushroom" : "cupcake"} size={filtered ? 64 : 96} className="mb-4" />
              <p className="text-[15px] text-muted-foreground">{filtered ? c.emptyFiltered : c.emptyNone}</p>
              {!filtered && (
                <Link href="/connect" className={buttonVariants({ className: "mt-4" })}>
                  {copy.common.connectLink}
                </Link>
              )}
            </div>
          ) : (
            <>
              {list.data.count >= LIMIT && (
                <p data-testid="full-note" className="mb-4 rounded-2xl bg-muted px-4 py-3 text-sm text-ink-2">
                  {c.fullNote(LIMIT)}
                </p>
              )}
              <ul className="divide-y divide-hairline rounded-[28px] border border-hairline px-5 md:px-6">
                {list.data.transactions.map((t) => (
                  <li key={t.id} className="grid grid-cols-[minmax(0,1fr)_auto] items-start gap-x-4 gap-y-1 py-4">
                    <div className="min-w-0">
                      <p className="font-semibold break-words">{t.merchant}</p>
                      <p className="text-sm text-muted-foreground">
                        {formatDate(t.occurred_on, lang)} · {t.category}
                      </p>
                      {t.note && <p className="mt-1 text-sm break-words text-ink-2">{t.note}</p>}
                    </div>
                    <div className="flex flex-col items-end gap-2">
                      <p className="font-semibold tabular-nums">{formatAmount(t.amount, t.currency)}</p>
                      <Button variant="outline" size="sm" className="h-10" aria-label={c.deleteLabel(t.merchant, formatAmount(t.amount, t.currency))} onClick={() => setToDelete(t)}>
                        {copy.common.delete}
                      </Button>
                    </div>
                  </li>
                ))}
              </ul>
            </>
          ))}
      </section>

      <Confirm
        open={toDelete !== null}
        title={c.deleteTitle}
        body={toDelete ? c.deleteBody(toDelete.merchant, formatAmount(toDelete.amount, toDelete.currency)) : ""}
        confirmLabel={copy.common.delete}
        cancelLabel={copy.common.cancel}
        busy={deleting}
        onConfirm={() => void confirmDelete()}
        onCancel={() => setToDelete(null)}
      />
      <Confirm
        open={dup !== null}
        title={c.add.dupTitle}
        body={c.add.dupBody}
        confirmLabel={c.add.saveAnyway}
        cancelLabel={c.add.dontSave}
        busy={saving}
        onConfirm={() => void save(true)}
        onCancel={() => setDup(null)}
      />
    </div>
  );
}
