"use client";

import { useEffect, useRef } from "react";
import { Button } from "@/components/ui/button";

// A confirmation on the browser's own <dialog>: it traps focus, closes on Escape and hands focus back to whatever
// opened it. The safe button (cancel) has the focus first, so Enter never deletes by accident.
export function Confirm({
  open,
  title,
  body,
  confirmLabel,
  cancelLabel,
  busy = false,
  onConfirm,
  onCancel,
}: {
  open: boolean;
  title: string;
  body: string;
  confirmLabel: string;
  cancelLabel: string;
  busy?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}) {
  const ref = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    const d = ref.current;
    if (!d) return;
    if (open && !d.open) d.showModal();
    if (!open && d.open) d.close();
  }, [open]);

  return (
    <dialog
      ref={ref}
      aria-labelledby="confirm-title"
      aria-describedby="confirm-body"
      onCancel={(e) => {
        e.preventDefault(); // Escape: let the parent decide, so state and dialog never disagree
        if (!busy) onCancel();
      }}
      onClick={(e) => {
        if (e.target === ref.current && !busy) onCancel(); // a click on the backdrop
      }}
      className="m-auto w-[min(92vw,420px)] rounded-[28px] border border-hairline bg-background p-7 text-foreground backdrop:bg-foreground/30"
    >
      <h2 id="confirm-title" className="text-xl font-bold tracking-[-0.02em]">
        {title}
      </h2>
      <p id="confirm-body" className="mt-2 text-[15px] leading-normal text-muted-foreground">
        {body}
      </p>
      <div className="mt-6 flex flex-wrap justify-end gap-3">
        <Button variant="outline" onClick={onCancel} disabled={busy} autoFocus>
          {cancelLabel}
        </Button>
        <Button variant="dark" onClick={onConfirm} disabled={busy}>
          {confirmLabel}
        </Button>
      </div>
    </dialog>
  );
}
