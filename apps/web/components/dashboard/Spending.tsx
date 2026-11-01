"use client";

import { useDash } from "./context";

// Replaced by the real Spending tab in the next commit.
export function Spending() {
  const { copy } = useDash();
  return <h2 className="text-2xl font-bold tracking-[-0.02em]">{copy.spending.title}</h2>;
}
