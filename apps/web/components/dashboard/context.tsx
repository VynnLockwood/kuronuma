"use client";

import { createContext, useContext } from "react";
import { dashboardCopy, type DashboardCopy } from "@/lib/dashboard-copy";
import type { Lang } from "@/lib/web";

type Dash = { lang: Lang; copy: DashboardCopy; onSignedOut: () => void };

export const DashContext = createContext<Dash>({ lang: "en", copy: dashboardCopy.en, onSignedOut: () => {} });
export const useDash = () => useContext(DashContext);
