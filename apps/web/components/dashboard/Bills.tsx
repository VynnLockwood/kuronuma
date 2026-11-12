"use client";

import { useMemo, useRef, useState } from "react";
import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { callTool } from "@/lib/web";
import { formatAmount, formatDate, formatMonth, localISODate } from "@/lib/format";
import { Confirm } from "./Confirm";
import { useDash } from "./context";
import { BlockError, BlockSkeleton } from "./states";
import { useTool } from "./useTool";

// The shapes below follow list_bills, monthly_bills_summary and the write tools in the API (plutus/bill_tools.go).
type Bill = {
  id: string;
  name: string;
  amount: string;
  currency: string;
  category: string;
  start_month: string;
  ended_from?: string;
  paid: boolean;
  paid_amount?: string;
  paid_on?: string;
  has_payments: boolean;
};
type BillList = { month: string; count: number; bills: Bill[] };
type Total = { currency: string; planned: string; paid: string; paid_count: number; unpaid_count: number };
type BillSummary = { month: string; totals: Total[] };
type Categories = { categories: { name: string }[] };
type Message = { kind: "ok" | "info" | "error"; text: string };
type Panel = { kind: "edit" | "end"; bill: Bill } | null;

const monthRe = /^\d{4}-(0[1-9]|1[0-2])$/;

/** Moves a YYYY-MM month by n months. */
function shiftMonth(ym: string, n: number): string {
  const [y, m] = ym.split("-").map(Number);
  const idx = y * 12 + (m - 1) + n;
  return `${Math.floor(idx / 12)}-${String((idx % 12) + 1).padStart(2, "0")}`;
}

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

function MessageLine({ m }: { m: Message }) {
  return (
    <p role={m.kind === "error" ? "alert" : "status"} className={m.kind === "error" ? "text-sm font-medium break-words text-destructive" : m.kind === "info" ? "text-sm font-medium break-words text-ink-2" : "text-sm font-medium break-words text-success"}>
      {m.text}
    </p>
  );
}

export function Bills() {
  const { lang, copy, onSignedOut } = useDash();
  const c = copy.bills;
  const thisMonth = useMemo(() => localISODate(new Date()).slice(0, 7), []);

  // The month. The draft lets a browser without a month picker take typing until it is a whole YYYY-MM.
  const [month, setMonth] = useState(thisMonth);
  const [draft, setDraft] = useState(thisMonth);
  const chooseMonth = (ym: string) => {
    setMonth(ym);
    setDraft(ym);
    setMessage(null);
    setPanel(null);
  };

  const [list, reloadList] = useTool<BillList>("list_bills", { month });
  const [summary, reloadSummary] = useTool<BillSummary>("monthly_bills_summary", { month });
  const [cats, reloadCats] = useTool<Categories>("list_categories", {});
  const names = cats.status === "ok" ? cats.data.categories.map((x) => x.name) : [];
  const refresh = () => {
    reloadList();
    reloadSummary();
    reloadCats();
  };

  // One place for the outcome of every action.
  const [message, setMessage] = useState<Message | null>(null);
  const [panel, setPanel] = useState<Panel>(null);

  // Runs a write tool. A server refusal is shown as the server wrote it.
  async function run<T extends { status?: string; message?: string }>(tool: string, args: Record<string, unknown>): Promise<T | null> {
    const r = await callTool<T>(tool, args, lang);
    if (!r.ok) {
      if (r.kind === "signed-out") onSignedOut();
      else setMessage({ kind: "error", text: r.message });
      return null;
    }
    return r.data;
  }

  // Mark paid: the planned amount, no date; the server applies its own rules for the date.
  const [busyId, setBusyId] = useState<string | null>(null);
  async function markPaid(b: Bill) {
    setBusyId(b.id);
    setMessage(null);
    const d = await run<{ status?: string; message?: string }>("confirm_bill_payment", { bill_id: b.id, month, amount: b.amount });
    setBusyId(null);
    if (!d) return;
    setMessage(d.status === "already_paid" ? { kind: "info", text: d.message ?? c.statusPaid } : { kind: "ok", text: c.markedPaid });
    refresh();
  }

  // Delete, after a confirmation. Only bills with no payments ever offer it.
  const [toDelete, setToDelete] = useState<Bill | null>(null);
  const [deleting, setDeleting] = useState(false);
  async function confirmDelete() {
    if (!toDelete) return;
    setDeleting(true);
    setMessage(null);
    const d = await run("delete_bill", { bill_id: toDelete.id });
    setDeleting(false);
    setToDelete(null);
    if (d) refresh();
  }

  // The add form.
  const [name, setName] = useState("");
  const [amount, setAmount] = useState("");
  const [currency, setCurrency] = useState("THB");
  const [cat, setCat] = useState("");
  const [start, setStart] = useState("");
  const [saving, setSaving] = useState(false);
  const [formMessage, setFormMessage] = useState<Message | null>(null);
  async function add() {
    setSaving(true);
    setFormMessage(null);
    const args: Record<string, unknown> = {
      name: name.trim(),
      amount: amount.trim(),
      currency: currency.trim().toUpperCase(),
      ...(cat.trim() && { category: cat.trim() }),
      ...(start.trim() && { start_month: start.trim() }),
    };
    const r = await callTool<{ category_created?: boolean; bill?: Bill }>("add_bill", args, lang);
    setSaving(false);
    if (!r.ok) {
      if (r.kind === "signed-out") onSignedOut();
      else setFormMessage({ kind: "error", text: r.message }); // the typing stays
      return;
    }
    setFormMessage({ kind: "ok", text: r.data.category_created ? c.form.savedNewCategory(r.data.bill?.category ?? cat.trim()) : c.form.saved });
    setName("");
    setAmount("");
    setCat("");
    setStart("");
    refresh();
  }

  // The edit and end panels.
  const [eName, setEName] = useState("");
  const [eAmount, setEAmount] = useState("");
  const [eCat, setECat] = useState("");
  const [eStart, setEStart] = useState("");
  const [eEnd, setEEnd] = useState(month);
  const [panelBusy, setPanelBusy] = useState(false);
  const [panelMessage, setPanelMessage] = useState<Message | null>(null);
  const panelHeading = useRef<HTMLHeadingElement>(null);
  function open(kind: "edit" | "end", b: Bill) {
    setPanel({ kind, bill: b });
    setPanelMessage(null);
    setEName(b.name);
    setEAmount(b.amount);
    setECat(b.category);
    setEStart(b.start_month);
    setEEnd(month);
    setTimeout(() => panelHeading.current?.focus(), 0);
  }
  async function saveEdit() {
    if (!panel) return;
    const b = panel.bill;
    // Only what changed is sent.
    const args: Record<string, unknown> = {
      bill_id: b.id,
      ...(eName.trim() !== b.name && { name: eName.trim() }),
      ...(eAmount.trim() !== b.amount && { amount: eAmount.trim() }),
      ...(eCat.trim() !== b.category && { category: eCat.trim() }),
      ...(eStart.trim() !== b.start_month && { start_month: eStart.trim() }),
    };
    if (Object.keys(args).length === 1) {
      setPanelMessage({ kind: "info", text: c.form.nothingToChange });
      return;
    }
    setPanelBusy(true);
    setPanelMessage(null);
    const r = await callTool("update_bill", args, lang);
    setPanelBusy(false);
    if (!r.ok) {
      if (r.kind === "signed-out") onSignedOut();
      else setPanelMessage({ kind: "error", text: r.message });
      return;
    }
    setPanel(null);
    setMessage({ kind: "ok", text: c.form.saved });
    refresh();
  }
  async function confirmEnd() {
    if (!panel) return;
    setPanelBusy(true);
    setPanelMessage(null);
    const r = await callTool("end_bill", { bill_id: panel.bill.id, ended_from: eEnd.trim() }, lang);
    setPanelBusy(false);
    if (!r.ok) {
      if (r.kind === "signed-out") onSignedOut();
      else setPanelMessage({ kind: "error", text: r.message });
      return;
    }
    setPanel(null);
    setMessage({ kind: "ok", text: c.endForm.ended });
    refresh();
  }

  return (
    <div className="flex flex-col gap-12">
      {/* Month */}
      <section aria-labelledby="bl-totals">
        <div className="mb-4 flex flex-wrap items-end justify-between gap-4">
          <h2 id="bl-totals" className="text-[22px] font-bold tracking-[-0.02em]">
            {c.totals}
          </h2>
          <div className="flex items-end gap-2">
            <Button variant="outline" size="sm" className="h-10" aria-label={c.prevMonth} onClick={() => chooseMonth(shiftMonth(month, -1))}>
              ‹
            </Button>
            <Field id="bl-month" label={c.month}>
              <Input
                id="bl-month"
                type="month"
                value={draft}
                onChange={(e) => {
                  setDraft(e.target.value);
                  if (monthRe.test(e.target.value)) {
                    setMonth(e.target.value);
                    setMessage(null);
                    setPanel(null);
                  }
                }}
              />
            </Field>
            <Button variant="outline" size="sm" className="h-10" aria-label={c.nextMonth} onClick={() => chooseMonth(shiftMonth(month, 1))}>
              ›
            </Button>
          </div>
        </div>
        {summary.status === "loading" && <BlockSkeleton />}
        {summary.status === "error" && <BlockError message={summary.message} onRetry={reloadSummary} />}
        {summary.status === "ok" &&
          (summary.data.totals.length === 0 ? (
            <p className="text-[15px] text-muted-foreground">{c.noTotals}</p>
          ) : (
            <div className="grid grid-cols-[minmax(0,1fr)] gap-4 md:grid-cols-2">
              {summary.data.totals.map((t) => (
                <Card key={t.currency} className="rounded-[28px] bg-muted p-6">
                  <h3 className="text-[15px] font-semibold text-ink-2">{t.currency}</h3>
                  <dl className="mt-3 grid grid-cols-2 gap-4">
                    <div>
                      <dt className="text-sm text-muted-foreground">{c.planned}</dt>
                      <dd className="text-xl font-bold tabular-nums">{formatAmount(t.planned, t.currency)}</dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{c.paid}</dt>
                      <dd className="text-xl font-bold tabular-nums">{formatAmount(t.paid, t.currency)}</dd>
                    </div>
                  </dl>
                  <p className="mt-3 text-sm text-muted-foreground">{c.counts(t.paid_count, t.unpaid_count)}</p>
                </Card>
              ))}
            </div>
          ))}
      </section>

      {/* The bills of the month */}
      <section aria-labelledby="bl-list">
        <h2 id="bl-list" className="mb-4 text-[22px] font-bold tracking-[-0.02em]">
          {c.listTitle}
        </h2>
        {message && (
          <div className="mb-4">
            {message.kind === "error" ? (
              <Alert variant="destructive">
                <div className="min-w-0">
                  <p className="font-semibold">{c.actionFailed}</p>
                  <p role="alert" className="mt-0.5 break-words opacity-90">
                    {message.text}
                  </p>
                </div>
              </Alert>
            ) : (
              <MessageLine m={message} />
            )}
          </div>
        )}
        {list.status === "loading" && <BlockSkeleton className="h-48" />}
        {list.status === "error" && <BlockError message={list.message} onRetry={reloadList} />}
        {list.status === "ok" &&
          (list.data.bills.length === 0 ? (
            <div className="rounded-[28px] bg-muted p-8 text-center">
              <p className="text-[15px] text-muted-foreground">{c.empty}</p>
            </div>
          ) : (
            <ul className="divide-y divide-hairline rounded-[28px] border border-hairline px-5 md:px-6">
              {list.data.bills.map((b) => {
                const shown = formatAmount(b.amount, b.currency);
                return (
                  <li key={b.id} className="py-4">
                    <div className="grid grid-cols-[minmax(0,1fr)_auto] items-start gap-x-4 gap-y-2">
                      <div className="min-w-0">
                        <p className="font-semibold break-words">{b.name}</p>
                        <p className="text-sm text-muted-foreground">{b.category}</p>
                        {b.paid && b.paid_amount && b.paid_on && <p className="mt-1 text-sm text-ink-2">{c.paidOn(formatAmount(b.paid_amount, b.currency), formatDate(b.paid_on, lang))}</p>}
                        {b.ended_from && <p className="mt-1 text-sm text-muted-foreground">{c.endsFrom(formatMonth(b.ended_from, lang))}</p>}
                      </div>
                      <div className="flex flex-col items-end gap-2">
                        <p className="font-semibold tabular-nums">{shown}</p>
                        <Badge variant={b.paid ? "success" : "muted"}>{b.paid ? c.statusPaid : c.statusUnpaid}</Badge>
                      </div>
                    </div>
                    <div className="mt-3 flex flex-wrap gap-2">
                      {!b.paid && (
                        <Button size="sm" className="h-10" aria-label={c.markPaidLabel(b.name, shown)} disabled={busyId !== null} onClick={() => void markPaid(b)}>
                          {busyId === b.id ? c.marking : c.markPaid}
                        </Button>
                      )}
                      <Button variant="outline" size="sm" className="h-10" aria-label={c.editLabel(b.name)} onClick={() => open("edit", b)}>
                        {c.edit}
                      </Button>
                      <Button variant="outline" size="sm" className="h-10" aria-label={c.endLabel(b.name)} onClick={() => open("end", b)}>
                        {c.end}
                      </Button>
                      {!b.has_payments && (
                        <Button variant="outline" size="sm" className="h-10" aria-label={c.deleteLabel(b.name)} onClick={() => setToDelete(b)}>
                          {copy.common.delete}
                        </Button>
                      )}
                    </div>
                    {panel?.bill.id === b.id && panel.kind === "edit" && (
                      <Card className="mt-4 rounded-[28px] bg-muted p-6">
                        <h3 ref={panelHeading} tabIndex={-1} className="mb-4 text-lg font-bold break-words outline-none">
                          {c.form.editTitle(b.name)}
                        </h3>
                        <form
                          onSubmit={(e) => {
                            e.preventDefault();
                            if (!panelBusy) void saveEdit();
                          }}
                          className="grid grid-cols-[minmax(0,1fr)] gap-4 md:grid-cols-2"
                        >
                          <Field id="bl-e-name" label={c.form.name}>
                            <Input id="bl-e-name" autoComplete="off" maxLength={100} value={eName} onChange={(e) => setEName(e.target.value)} />
                          </Field>
                          <Field id="bl-e-amount" label={c.form.amount}>
                            <Input id="bl-e-amount" inputMode="decimal" autoComplete="off" value={eAmount} onChange={(e) => setEAmount(e.target.value)} />
                          </Field>
                          <Field id="bl-e-currency" label={c.form.currency} hint={c.form.currencyLocked}>
                            <Input id="bl-e-currency" value={b.currency} disabled readOnly />
                          </Field>
                          <Field id="bl-e-cat" label={c.form.category} hint={c.form.categoryHint}>
                            <Input id="bl-e-cat" list="bl-cat-list" autoComplete="off" maxLength={40} value={eCat} onChange={(e) => setECat(e.target.value)} />
                          </Field>
                          <Field id="bl-e-start" label={c.form.startMonth}>
                            <Input id="bl-e-start" type="month" value={eStart} onChange={(e) => setEStart(e.target.value)} />
                          </Field>
                          <div className="flex flex-wrap items-center gap-3 md:col-span-2">
                            <Button type="submit" disabled={panelBusy}>
                              {panelBusy ? c.form.saving : c.form.save}
                            </Button>
                            <Button type="button" variant="outline" disabled={panelBusy} onClick={() => setPanel(null)}>
                              {c.form.cancel}
                            </Button>
                            {panelMessage && <MessageLine m={panelMessage} />}
                          </div>
                        </form>
                      </Card>
                    )}
                    {panel?.bill.id === b.id && panel.kind === "end" && (
                      <Card className="mt-4 rounded-[28px] bg-muted p-6">
                        <h3 ref={panelHeading} tabIndex={-1} className="mb-4 text-lg font-bold break-words outline-none">
                          {c.endForm.title(b.name)}
                        </h3>
                        <form
                          onSubmit={(e) => {
                            e.preventDefault();
                            if (!panelBusy) void confirmEnd();
                          }}
                          className="flex flex-col gap-4"
                        >
                          <Field id="bl-end-from" label={c.endForm.from} hint={c.endForm.hint}>
                            <Input id="bl-end-from" type="month" value={eEnd} onChange={(e) => setEEnd(e.target.value)} />
                          </Field>
                          <div className="flex flex-wrap items-center gap-3">
                            <Button type="submit" variant="dark" disabled={panelBusy}>
                              {panelBusy ? c.endForm.ending : c.endForm.confirm}
                            </Button>
                            <Button type="button" variant="outline" disabled={panelBusy} onClick={() => setPanel(null)}>
                              {c.form.cancel}
                            </Button>
                            {panelMessage && <MessageLine m={panelMessage} />}
                          </div>
                        </form>
                      </Card>
                    )}
                  </li>
                );
              })}
            </ul>
          ))}
        <datalist id="bl-cat-list">
          {names.map((n) => (
            <option key={n} value={n} />
          ))}
        </datalist>
      </section>

      {/* Add a bill */}
      <section aria-labelledby="bl-add">
        <h2 id="bl-add" className="mb-4 text-[22px] font-bold tracking-[-0.02em]">
          {c.form.addTitle}
        </h2>
        <Card className="rounded-[28px] bg-muted p-6 md:p-8">
          <form
            onSubmit={(e) => {
              e.preventDefault();
              if (!saving) void add();
            }}
            className="grid grid-cols-[minmax(0,1fr)] gap-4 md:grid-cols-6"
          >
            <div className="md:col-span-3">
              <Field id="bl-name" label={c.form.name}>
                <Input id="bl-name" autoComplete="off" maxLength={100} value={name} onChange={(e) => setName(e.target.value)} />
              </Field>
            </div>
            <div className="md:col-span-2">
              <Field id="bl-amount" label={c.form.amount}>
                <Input id="bl-amount" inputMode="decimal" autoComplete="off" value={amount} onChange={(e) => setAmount(e.target.value)} />
              </Field>
            </div>
            <div className="md:col-span-1">
              <Field id="bl-currency" label={c.form.currency}>
                <Input id="bl-currency" autoComplete="off" maxLength={3} className="uppercase" value={currency} onChange={(e) => setCurrency(e.target.value)} />
              </Field>
            </div>
            <div className="md:col-span-3">
              <Field id="bl-cat" label={c.form.category} hint={c.form.categoryHint}>
                <Input id="bl-cat" list="bl-cat-list" autoComplete="off" maxLength={40} value={cat} onChange={(e) => setCat(e.target.value)} />
              </Field>
            </div>
            <div className="md:col-span-3">
              <Field id="bl-start" label={c.form.startMonth} hint={c.form.startMonthHint}>
                <Input id="bl-start" type="month" value={start} onChange={(e) => setStart(e.target.value)} />
              </Field>
            </div>
            <div className="flex flex-wrap items-center gap-4 md:col-span-6">
              <Button type="submit" size="lg" disabled={saving}>
                {saving ? c.form.saving : c.form.add}
              </Button>
              {formMessage && <MessageLine m={formMessage} />}
            </div>
          </form>
        </Card>
      </section>

      <Confirm
        open={toDelete !== null}
        title={c.deleteTitle}
        body={toDelete ? c.deleteBody(toDelete.name) : ""}
        confirmLabel={copy.common.delete}
        cancelLabel={copy.common.cancel}
        busy={deleting}
        onConfirm={() => void confirmDelete()}
        onCancel={() => setToDelete(null)}
      />
    </div>
  );
}
