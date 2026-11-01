"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button, buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { dueBucket, formatDate, localISODate } from "@/lib/format";
import { callTool } from "@/lib/web";
import { useDash } from "./context";
import { BlockError, BlockSkeleton } from "./states";
import { useTool } from "./useTool";

export type Memory = {
  id: string;
  kind: "task" | "todo" | "note";
  title: string;
  body?: string;
  status: "open" | "done" | "archived";
  priority: "low" | "normal" | "high";
  category: string;
  tags: string[];
  due_on?: string;
};
type MemoryList = { count: number; memories: Memory[] };
type Categories = { categories: { name: string }[] };
type Filters = { kind: string; status: string; category: string; tag: string; q: string };

const LIMIT = 200;
const empty: Filters = { kind: "", status: "", category: "", tag: "", q: "" };
const statusOrder = ["open", "done", "archived"] as const;

// The filters live in the query string (next to ?tab=tasks), so a reload or a shared link keeps them.
const readFilters = (): Filters => {
  if (typeof window === "undefined") return empty;
  const p = new URLSearchParams(window.location.search);
  return { kind: p.get("kind") ?? "", status: p.get("status") ?? "", category: p.get("category") ?? "", tag: p.get("tag") ?? "", q: p.get("q") ?? "" };
};
const writeFilters = (f: Filters) => {
  const p = new URLSearchParams({ tab: "tasks" });
  for (const [k, v] of Object.entries(f)) if (v) p.set(k, v);
  window.history.replaceState(null, "", `${window.location.pathname}?${p.toString()}`);
};

function useDebounced<T>(value: T, ms: number): T {
  const [v, setV] = useState(value);
  useEffect(() => {
    const id = setTimeout(() => setV(value), ms);
    return () => clearTimeout(id);
  }, [value, ms]);
  return v;
}

const selectClass = "h-12 w-full min-w-0 rounded-2xl border border-input bg-background px-4 text-[15px] text-foreground focus-visible:border-ring";

function Field({ id, label, children }: { id: string; label: string; children: React.ReactNode }) {
  return (
    <div className="flex min-w-0 flex-col gap-1.5">
      <label htmlFor={id} className="text-sm font-semibold text-ink-2">
        {label}
      </label>
      {children}
    </div>
  );
}

export function Memories() {
  const { lang, copy, onSignedOut } = useDash();
  const c = copy.tasks;
  const today = useMemo(() => localISODate(new Date()), []);

  const [filters, setFilters] = useState<Filters>(readFilters);
  // Typing in the tag and search boxes waits a moment before it asks the server.
  const [tagText, setTagText] = useState(filters.tag);
  const [qText, setQText] = useState(filters.q);
  const tag = useDebounced(tagText, 350);
  const q = useDebounced(qText, 350);
  useEffect(() => setFilters((f) => (f.tag === tag && f.q === q ? f : { ...f, tag, q })), [tag, q]);
  useEffect(() => writeFilters(filters), [filters]);

  const args = {
    ...(filters.kind && { kind: filters.kind }),
    ...(filters.status && { status: filters.status }),
    ...(filters.category && { category: filters.category }),
    ...(filters.tag && { tag: filters.tag }),
    ...(filters.q && { query: filters.q }),
    limit: LIMIT,
  };
  const [list, reloadList] = useTool<MemoryList>("list_memories", args);
  const [cats] = useTool<Categories>("list_categories", {});
  const names = cats.status === "ok" ? cats.data.categories.map((x) => x.name) : [];
  const filtered = Boolean(filters.kind || filters.status || filters.category || filters.tag || filters.q);
  const clear = () => {
    setFilters(empty);
    setTagText("");
    setQText("");
  };

  const [busy, setBusy] = useState<string | null>(null);
  const [notice, setNotice] = useState<{ kind: "info" | "error"; text: string } | null>(null);

  async function setStatus(m: Memory, status: Memory["status"]) {
    setBusy(m.id);
    setNotice(null);
    const r = await callTool<{ status?: string }>("update_memory", { id: m.id, status }, lang);
    setBusy(null);
    if (!r.ok) {
      if (r.kind === "signed-out") onSignedOut();
      else setNotice({ kind: "error", text: r.message });
      return;
    }
    if (r.data.status === "not_found") setNotice({ kind: "info", text: c.gone });
    reloadList();
  }

  return (
    <section aria-labelledby="mem-title" className="flex flex-col gap-8">
      <h2 id="mem-title" className="text-[22px] font-bold tracking-[-0.02em]">
        {c.title}
      </h2>

      <div className="grid grid-cols-[minmax(0,1fr)] gap-4 sm:grid-cols-2 lg:grid-cols-5">
        <Field id="mem-kind" label={c.filters.kind}>
          <select id="mem-kind" className={selectClass} value={filters.kind} onChange={(e) => setFilters({ ...filters, kind: e.target.value })}>
            <option value="">{c.filters.allKinds}</option>
            {(["task", "todo", "note"] as const).map((k) => (
              <option key={k} value={k}>
                {c.kinds[k]}
              </option>
            ))}
          </select>
        </Field>
        <Field id="mem-status" label={c.filters.status}>
          <select id="mem-status" className={selectClass} value={filters.status} onChange={(e) => setFilters({ ...filters, status: e.target.value })}>
            <option value="">{c.filters.allStatuses}</option>
            {statusOrder.map((s) => (
              <option key={s} value={s}>
                {c.statuses[s]}
              </option>
            ))}
          </select>
        </Field>
        <Field id="mem-cat" label={c.filters.category}>
          <select id="mem-cat" className={selectClass} value={filters.category} onChange={(e) => setFilters({ ...filters, category: e.target.value })}>
            <option value="">{c.filters.allCategories}</option>
            {names.map((n) => (
              <option key={n} value={n}>
                {n}
              </option>
            ))}
          </select>
        </Field>
        <Field id="mem-tag" label={c.filters.tag}>
          <Input id="mem-tag" autoComplete="off" value={tagText} onChange={(e) => setTagText(e.target.value)} />
        </Field>
        <Field id="mem-q" label={c.filters.search}>
          <Input id="mem-q" type="search" autoComplete="off" value={qText} onChange={(e) => setQText(e.target.value)} />
        </Field>
      </div>
      {filtered && (
        <div className="-mt-4">
          <Button variant="ghost" size="sm" className="h-10" onClick={clear}>
            {c.filters.clear}
          </Button>
        </div>
      )}

      {notice && (
        <Alert variant={notice.kind === "error" ? "destructive" : "default"}>
          <div>
            {notice.kind === "error" && <p className="font-semibold">{c.updateFailed}</p>}
            <p className="break-words">{notice.text}</p>
          </div>
        </Alert>
      )}

      {list.status === "loading" && <BlockSkeleton className="h-56" />}
      {list.status === "error" && <BlockError message={list.message} onRetry={reloadList} />}
      {list.status === "ok" &&
        (list.data.memories.length === 0 ? (
          <div className="rounded-[28px] bg-muted p-8 text-center">
            <p className="text-[15px] text-muted-foreground">{filtered ? c.emptyFiltered : c.emptyNone}</p>
            {!filtered && (
              <Link href="/connect" className={buttonVariants({ className: "mt-4" })}>
                {copy.common.connectLink}
              </Link>
            )}
          </div>
        ) : (
          <div className="flex flex-col gap-10">
            {list.data.count >= LIMIT && <p className="rounded-2xl bg-muted px-4 py-3 text-sm text-ink-2">{c.fullNote(LIMIT)}</p>}
            {statusOrder.map((s) => {
              const items = list.data.memories.filter((m) => m.status === s);
              if (items.length === 0) return null;
              return (
                <div key={s}>
                  <h3 className="mb-3 text-[17px] font-bold tracking-[-0.01em]">
                    {c.statuses[s]} · {items.length}
                  </h3>
                  <ul className="divide-y divide-hairline rounded-[28px] border border-hairline px-5 md:px-6">
                    {items.map((m) => {
                      const overdue = m.status === "open" && dueBucket(m.due_on, today) === "overdue";
                      return (
                        <li key={m.id} className="flex flex-col gap-3 py-4">
                          <div className="min-w-0">
                            <div className="mb-1.5 flex flex-wrap items-center gap-2">
                              <Badge>{c.kinds[m.kind]}</Badge>
                              {m.priority !== "normal" && <Badge variant={m.priority === "high" ? "info" : "outline"}>{c.priorities[m.priority]}</Badge>}
                              {overdue && <Badge variant="destructive">{c.overdue}</Badge>}
                            </div>
                            <p className="font-semibold break-words">{m.title}</p>
                            {m.body && <p className="mt-1 text-sm break-words text-ink-2">{m.body}</p>}
                            <p className="mt-1.5 text-sm break-words text-muted-foreground">
                              {m.due_on && <span>{c.due(formatDate(m.due_on, lang))} · </span>}
                              {m.category}
                            </p>
                            {m.tags.length > 0 && (
                              <ul className="mt-2 flex flex-wrap gap-1.5">
                                {m.tags.map((t) => (
                                  <li key={t} className="max-w-full">
                                    <Badge variant="outline" className="max-w-full overflow-hidden text-ellipsis">
                                      {t}
                                    </Badge>
                                  </li>
                                ))}
                              </ul>
                            )}
                          </div>
                          <div className="flex flex-wrap gap-2">
                            {m.status !== "done" && (
                              <Button variant="outline" size="sm" className="h-10" disabled={busy === m.id} aria-label={c.completeLabel(m.title)} onClick={() => void setStatus(m, "done")}>
                                {c.complete}
                              </Button>
                            )}
                            {m.status !== "open" && (
                              <Button variant="outline" size="sm" className="h-10" disabled={busy === m.id} aria-label={c.reopenLabel(m.title)} onClick={() => void setStatus(m, "open")}>
                                {c.reopen}
                              </Button>
                            )}
                            {m.status !== "archived" && (
                              <Button variant="ghost" size="sm" className="h-10" disabled={busy === m.id} aria-label={c.archiveLabel(m.title)} onClick={() => void setStatus(m, "archived")}>
                                {c.archive}
                              </Button>
                            )}
                          </div>
                        </li>
                      );
                    })}
                  </ul>
                </div>
              );
            })}
          </div>
        ))}
    </section>
  );
}
