"use client";

import Link from "next/link";
import { useEffect, useMemo, useRef, useState } from "react";
import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button, buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { dueBucket, formatDate, localISODate } from "@/lib/format";
import { callTool } from "@/lib/web";
import { Confirm } from "./Confirm";
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

const textareaClass =
  "min-h-24 w-full min-w-0 rounded-2xl border border-input bg-background px-4 py-3 text-[15px] text-foreground transition-[border-color] focus-visible:border-ring";

/** "home, plants," becomes ["home", "plants"]: the tags box is comma-separated text. */
const parseTags = (text: string) => text.split(",").map((t) => t.trim()).filter(Boolean);
const sameTags = (a: string[], b: string[]) => JSON.stringify(a.map((x) => x.toLowerCase())) === JSON.stringify(b.map((x) => x.toLowerCase()));

type Editor = { mode: "add" } | { mode: "edit"; memory: Memory };
type Notice = { kind: "ok" | "info" | "error"; text: string };

const selectClass = "h-12 w-full min-w-0 rounded-2xl border border-input bg-background px-4 text-[15px] text-foreground focus-visible:border-ring";

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
  const [cats, reloadCats] = useTool<Categories>("list_categories", {});
  const names = cats.status === "ok" ? cats.data.categories.map((x) => x.name) : [];
  const filtered = Boolean(filters.kind || filters.status || filters.category || filters.tag || filters.q);
  const clear = () => {
    setFilters(empty);
    setTagText("");
    setQText("");
  };

  const [busy, setBusy] = useState<string | null>(null);
  const [notice, setNotice] = useState<Notice | null>(null);
  const heading = useRef<HTMLHeadingElement>(null);

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

  // The add and edit form.
  const [editor, setEditor] = useState<Editor | null>(null);
  const [kind, setKind] = useState<Memory["kind"]>("task");
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [due, setDue] = useState("");
  const [priority, setPriority] = useState<Memory["priority"]>("normal");
  const [category, setCategory] = useState("");
  const [tagsText, setTagsText] = useState("");
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const trigger = useRef<HTMLElement | null>(null);
  const titleInput = useRef<HTMLInputElement>(null);
  const focusAfterClose = useRef<"trigger" | "heading" | null>(null);

  const categoryField = (m: Memory) => (m.category === "uncategorized" ? "" : m.category);
  function openAdd(el: HTMLElement) {
    trigger.current = el;
    setKind("task");
    setTitle("");
    setBody("");
    setDue("");
    setPriority("normal");
    setCategory("");
    setTagsText("");
    setFormError(null);
    setNotice(null);
    setEditor({ mode: "add" });
  }
  function openEdit(m: Memory, el: HTMLElement) {
    trigger.current = el;
    setKind(m.kind);
    setTitle(m.title);
    setBody(m.body ?? "");
    setDue(m.due_on ?? "");
    setPriority(m.priority);
    setCategory(categoryField(m));
    setTagsText(m.tags.join(", "));
    setFormError(null);
    setNotice(null);
    setEditor({ mode: "edit", memory: m });
  }
  function closeEditor(focus: "trigger" | "heading") {
    focusAfterClose.current = focus;
    setEditor(null);
  }
  // Focus goes to the title when the form opens, and back where it came from when it closes. After a change the
  // list reloads and the row's button is replaced, so then it goes to the heading instead.
  useEffect(() => {
    if (editor) {
      titleInput.current?.focus();
      return;
    }
    const f = focusAfterClose.current;
    focusAfterClose.current = null;
    if (f === "trigger" && trigger.current && document.contains(trigger.current)) trigger.current.focus();
    else if (f) heading.current?.focus();
  }, [editor]);

  async function submit() {
    if (!editor || saving) return;
    const tags = parseTags(tagsText);
    if (editor.mode === "add") {
      const args: Record<string, unknown> = {
        kind,
        title: title.trim(),
        ...(body.trim() && { body: body.trim() }),
        ...(due && { due_on: due }),
        ...(priority !== "normal" && { priority }),
        ...(category.trim() && { category: category.trim() }),
        ...(tags.length > 0 && { tags }),
      };
      setSaving(true);
      setFormError(null);
      const r = await callTool<{ category_created?: boolean; memory?: Memory }>("add_memory", args, lang);
      setSaving(false);
      if (!r.ok) {
        if (r.kind === "signed-out") onSignedOut();
        else setFormError(r.message); // the form stays, with the typing
        return;
      }
      setNotice({ kind: "ok", text: r.data.category_created ? c.savedNewCategory(r.data.memory?.category ?? category.trim()) : c.saved });
      closeEditor("heading");
      reloadList();
      reloadCats();
      return;
    }
    // Edit: send only what changed.
    const o = editor.memory;
    const changes: Record<string, unknown> = {
      ...(title.trim() !== o.title && { title: title.trim() }),
      ...(body.trim() !== (o.body ?? "") && { body: body.trim() }),
      ...(due !== (o.due_on ?? "") && { due_on: due }),
      ...(priority !== o.priority && { priority }),
      ...(category.trim() !== categoryField(o) && { category: category.trim() }),
      ...(!sameTags(tags, o.tags) && { tags }),
    };
    if (Object.keys(changes).length === 0) {
      setNotice({ kind: "info", text: c.nothingToChange });
      closeEditor("trigger");
      return;
    }
    setSaving(true);
    setFormError(null);
    const r = await callTool<{ status?: string; memory?: Memory }>("update_memory", { id: o.id, ...changes }, lang);
    setSaving(false);
    if (!r.ok) {
      if (r.kind === "signed-out") onSignedOut();
      else setFormError(r.message);
      return;
    }
    setNotice(r.data.status === "not_found" ? { kind: "info", text: c.gone } : { kind: "ok", text: c.updated });
    closeEditor("heading");
    reloadList();
    reloadCats();
  }

  // Delete, after a confirmation. Archiving is the gentle option and sits next to it.
  const [toDelete, setToDelete] = useState<Memory | null>(null);
  const [deleting, setDeleting] = useState(false);
  const focusAfterDelete = useRef(false);
  async function confirmDelete() {
    if (!toDelete) return;
    setDeleting(true);
    const r = await callTool<{ status?: string }>("delete_memory", { id: toDelete.id }, lang);
    setDeleting(false);
    setToDelete(null);
    if (!r.ok) {
      if (r.kind === "signed-out") onSignedOut();
      else setNotice({ kind: "error", text: r.message });
      return;
    }
    setNotice(r.data.status === "not_found" ? { kind: "info", text: c.gone } : null);
    focusAfterDelete.current = true; // the row is about to go
    reloadList();
  }
  // Focus moves only after the dialog has closed (it hands focus back to the Delete button, which is then removed).
  useEffect(() => {
    if (toDelete === null && focusAfterDelete.current) {
      focusAfterDelete.current = false;
      heading.current?.focus();
    }
  }, [toDelete]);

  return (
    <section aria-labelledby="mem-title" className="flex flex-col gap-8">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <h2 id="mem-title" ref={heading} tabIndex={-1} className="text-[22px] font-bold tracking-[-0.02em] outline-none">
          {c.title}
        </h2>
        <Button className="h-11" onClick={(e) => openAdd(e.currentTarget)}>
          {c.form.addButton}
        </Button>
      </div>

      {editor && (
        <section aria-labelledby="mem-form-title">
          <Card className="rounded-[28px] bg-muted p-6 md:p-8">
            <h3 id="mem-form-title" className="mb-4 text-[19px] font-bold tracking-[-0.01em]">
              {editor.mode === "add" ? c.form.addTitle : c.form.editTitle}
            </h3>
            <form
              onSubmit={(e) => {
                e.preventDefault();
                void submit();
              }}
              className="grid grid-cols-[minmax(0,1fr)] gap-4 md:grid-cols-2"
            >
              {editor.mode === "add" ? (
                <Field id="mem-f-kind" label={c.form.kind}>
                  <select id="mem-f-kind" className={selectClass} value={kind} onChange={(e) => setKind(e.target.value as Memory["kind"])}>
                    {(["task", "todo", "note"] as const).map((k) => (
                      <option key={k} value={k}>
                        {c.kinds[k]}
                      </option>
                    ))}
                  </select>
                </Field>
              ) : (
                <div className="flex flex-col gap-1.5">
                  <p className="text-sm font-semibold text-ink-2">{c.form.kind}</p>
                  <p className="flex h-12 items-center text-[15px]">{c.kinds[kind]}</p>
                </div>
              )}
              <Field id="mem-f-priority" label={c.form.priority}>
                <select id="mem-f-priority" className={selectClass} value={priority} onChange={(e) => setPriority(e.target.value as Memory["priority"])}>
                  {(["low", "normal", "high"] as const).map((p) => (
                    <option key={p} value={p}>
                      {c.priorities[p]}
                    </option>
                  ))}
                </select>
              </Field>
              <div className="md:col-span-2">
                <Field id="mem-f-title" label={c.form.title}>
                  <Input id="mem-f-title" ref={titleInput} autoComplete="off" maxLength={200} value={title} onChange={(e) => setTitle(e.target.value)} />
                </Field>
              </div>
              <div className="md:col-span-2">
                <Field id="mem-f-body" label={c.form.body} hint={c.form.bodyHint}>
                  <textarea id="mem-f-body" className={textareaClass} maxLength={4000} value={body} onChange={(e) => setBody(e.target.value)} />
                </Field>
              </div>
              <Field id="mem-f-due" label={c.form.due}>
                <Input id="mem-f-due" type="date" value={due} onChange={(e) => setDue(e.target.value)} />
              </Field>
              <Field id="mem-f-cat" label={c.form.category} hint={c.form.categoryHint}>
                <Input id="mem-f-cat" list="mem-f-cat-list" autoComplete="off" maxLength={40} value={category} onChange={(e) => setCategory(e.target.value)} />
                <datalist id="mem-f-cat-list">
                  {names.map((n) => (
                    <option key={n} value={n} />
                  ))}
                </datalist>
              </Field>
              <div className="md:col-span-2">
                <Field id="mem-f-tags" label={c.form.tags} hint={c.form.tagsHint}>
                  <Input id="mem-f-tags" autoComplete="off" value={tagsText} onChange={(e) => setTagsText(e.target.value)} />
                </Field>
              </div>
              <div className="flex flex-wrap items-center gap-3 md:col-span-2">
                <Button type="submit" size="lg" disabled={saving}>
                  {saving ? c.form.saving : c.form.save}
                </Button>
                <Button type="button" variant="outline" size="lg" disabled={saving} onClick={() => closeEditor("trigger")}>
                  {copy.common.cancel}
                </Button>
                {formError && (
                  <p role="alert" className="text-sm font-medium break-words text-destructive">
                    {formError}
                  </p>
                )}
              </div>
            </form>
          </Card>
        </section>
      )}

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

      {notice?.kind === "error" && (
        <Alert variant="destructive">
          <div>
            <p className="font-semibold">{c.updateFailed}</p>
            <p className="break-words">{notice.text}</p>
          </div>
        </Alert>
      )}
      {notice && notice.kind !== "error" && (
        <p role="status" className="rounded-2xl bg-muted px-4 py-3 text-sm font-medium break-words text-ink-2">
          {notice.text}
        </p>
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
                            <Button variant="outline" size="sm" className="h-10" disabled={busy === m.id} aria-label={c.editLabel(m.title)} onClick={(e) => openEdit(m, e.currentTarget)}>
                              {c.edit}
                            </Button>
                            {m.status !== "archived" && (
                              <Button variant="ghost" size="sm" className="h-10" disabled={busy === m.id} aria-label={c.archiveLabel(m.title)} onClick={() => void setStatus(m, "archived")}>
                                {c.archive}
                              </Button>
                            )}
                            <Button variant="ghost" size="sm" className="h-10 text-destructive" disabled={busy === m.id} aria-label={c.deleteLabel(m.title)} onClick={() => setToDelete(m)}>
                              {copy.common.delete}
                            </Button>
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
      <Confirm
        open={toDelete !== null}
        title={c.deleteTitle}
        body={toDelete ? c.deleteBody(toDelete.title) : ""}
        confirmLabel={copy.common.delete}
        cancelLabel={copy.common.cancel}
        busy={deleting}
        onConfirm={() => void confirmDelete()}
        onCancel={() => setToDelete(null)}
      />
    </section>
  );
}
