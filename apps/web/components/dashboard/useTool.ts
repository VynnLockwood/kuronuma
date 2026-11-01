"use client";

import { useCallback, useEffect, useState } from "react";
import { callTool } from "@/lib/web";
import { useDash } from "./context";

export type ToolState<T> = { status: "loading" } | { status: "ok"; data: T } | { status: "error"; message: string };

/**
 * Runs a tool through the web bridge whenever the arguments change, and again when reload() is called.
 * A 401 hands the person back to the sign-in card; any other failure becomes an error state with its message.
 * An answer that arrives after the arguments changed (or the tab closed) is dropped.
 */
export function useTool<T = Record<string, unknown>>(tool: string, args: Record<string, unknown> = {}): [ToolState<T>, () => void] {
  const { lang, onSignedOut } = useDash();
  const [state, setState] = useState<ToolState<T>>({ status: "loading" });
  const [tick, setTick] = useState(0);
  const key = JSON.stringify(args);

  useEffect(() => {
    let current = true;
    setState({ status: "loading" });
    callTool<T>(tool, JSON.parse(key) as Record<string, unknown>, lang).then((r) => {
      if (!current) return;
      if (r.ok) setState({ status: "ok", data: r.data });
      else if (r.kind === "signed-out") onSignedOut();
      else setState({ status: "error", message: r.message });
    });
    return () => {
      current = false;
    };
  }, [tool, key, tick, lang, onSignedOut]);

  const reload = useCallback(() => setTick((n) => n + 1), []);
  return [state, reload];
}
