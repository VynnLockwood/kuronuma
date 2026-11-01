"use client";

import { Icon } from "@/components/icons";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useDash } from "./context";

/** The loading placeholder for one block. */
export function BlockSkeleton({ className = "h-32" }: { className?: string }) {
  const { copy } = useDash();
  return <Skeleton aria-busy="true" aria-label={copy.common.loading} className={`w-full rounded-3xl bg-muted ${className}`} />;
}

/** One block failed: its message and a retry, while the other blocks keep working. */
export function BlockError({ message, onRetry }: { message: string; onRetry: () => void }) {
  const { copy } = useDash();
  return (
    <Alert variant="destructive" className="items-start justify-between">
      <Icon name="alert" size={18} />
      <div className="min-w-0 flex-1">
        <p className="font-semibold">{copy.common.errorTitle}</p>
        <p className="mt-0.5 break-words opacity-90">{message}</p>
      </div>
      <Button variant="outline" size="sm" onClick={onRetry}>
        {copy.common.retry}
      </Button>
    </Alert>
  );
}
