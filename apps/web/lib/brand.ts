export const BRAND = {
  name: "Kuronuma",
  thai: "คุโรนุมะ",
  full: "Kuronuma",
} as const;

// Public links. The repository link (header button and footer) defaults to the public repository.
// Set NEXT_PUBLIC_REPO_URL at build time to point it at a fork, or to an empty value to hide the link.
export const PUBLIC_REPO_URL = "https://github.com/VynnLockwood/kuronuma";
export const REPO_URL = process.env.NEXT_PUBLIC_REPO_URL ?? PUBLIC_REPO_URL;

// Absolute site URL, needed for social-preview (Open Graph) image URLs. Optional.
export const SITE_URL = process.env.NEXT_PUBLIC_SITE_URL ?? "";
