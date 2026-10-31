import type { ReactNode } from "react";

const paths = {
  receipt: (
    <>
      <path d="M6 3h12v18l-3-2-3 2-3-2-3 2V3z" />
      <path d="M9 8h6M9 12h6" />
    </>
  ),
  chat: <path d="M4 5h16v11H9l-5 4V5z" />,
  shield: (
    <>
      <path d="M12 3l7 3v5c0 5-3 8-7 10-4-2-7-5-7-10V6l7-3z" />
      <path d="M9 12l2 2 4-4" />
    </>
  ),
  layers: (
    <>
      <path d="M12 3l9 5-9 5-9-5 9-5z" />
      <path d="M3 13l9 5 9-5" />
    </>
  ),
  plug: (
    <>
      <path d="M9 3v5M15 3v5" />
      <path d="M6 8h12v3a6 6 0 0 1-12 0V8z" />
      <path d="M12 17v4" />
    </>
  ),
  code: <path d="M8 8l-5 4 5 4M16 8l5 4-5 4M14 5l-4 14" />,
  lock: (
    <>
      <rect x="5" y="11" width="14" height="10" rx="2" />
      <path d="M8 11V8a4 4 0 0 1 8 0v3" />
    </>
  ),
  sparkle: <path d="M12 3l2 5 5 2-5 2-2 5-2-5-5-2 5-2 2-5z" />,
  check: <path d="M5 12l4 4 10-10" />,
  cross: <path d="M6 6l12 12M18 6L6 18" />,
  arrow: <path d="M5 12h14M13 6l6 6-6 6" />,
  chevron: <path d="M6 9l6 6 6-6" />,
  globe: (
    <>
      <circle cx="12" cy="12" r="9" />
      <path d="M3 12h18M12 3c3 3 3 15 0 18M12 3c-3 3-3 15 0 18" />
    </>
  ),
  clock: (
    <>
      <circle cx="12" cy="12" r="9" />
      <path d="M12 7v5l3 2" />
    </>
  ),
  copy: (
    <>
      <rect x="8" y="8" width="12" height="12" rx="2" />
      <path d="M16 8V5a1 1 0 0 0-1-1H5a1 1 0 0 0-1 1v10a1 1 0 0 0 1 1h3" />
    </>
  ),
  alert: (
    <>
      <circle cx="12" cy="12" r="9" />
      <path d="M12 8v5M12 16h.01" />
    </>
  ),
  unplug: (
    <>
      <path d="M9 3v5M15 3v5" />
      <path d="M6 8h12v3a6 6 0 0 1-12 0V8z" />
      <path d="M12 17v4M4 4l16 16" />
    </>
  ),
  google: (
    <path d="M21 12.2c0-.6-.1-1.3-.2-1.9H12v3.6h5a4.3 4.3 0 0 1-1.9 2.8v2.3h3C19.9 17.4 21 15 21 12.2zM12 21c2.5 0 4.6-.8 6.1-2.3l-3-2.3c-.8.6-1.9.9-3.1.9-2.4 0-4.4-1.6-5.1-3.8H3.8v2.4A9 9 0 0 0 12 21zM6.9 13.5a5.4 5.4 0 0 1 0-3.4V7.7H3.8a9 9 0 0 0 0 8.2zM12 6.6c1.3 0 2.5.5 3.5 1.4l2.6-2.6A9 9 0 0 0 3.8 7.7l3.1 2.4C7.6 8.1 9.6 6.6 12 6.6z" />
  ),
  camera: (
    <>
      <path d="M4 8h3l2-3h6l2 3h3v11H4V8z" />
      <circle cx="12" cy="13" r="3.5" />
    </>
  ),
} satisfies Record<string, ReactNode>;

export type IconName = keyof typeof paths;

// Solid glyphs (brand marks) are filled; every other icon is a 1.8 px stroke.
const filled: ReadonlySet<IconName> = new Set<IconName>(["google"]);

export function Icon({ name, size = 24, className }: { name: IconName; size?: number; className?: string }) {
  const solid = filled.has(name);
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill={solid ? "currentColor" : "none"}
      stroke={solid ? "none" : "currentColor"}
      strokeWidth={1.8}
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
      aria-hidden="true"
      focusable="false"
    >
      {paths[name]}
    </svg>
  );
}
