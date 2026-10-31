export const BRAND = {
  name: "Kuronuma",
  thai: "คุโรนุมะ",
  full: "Kuronuma",
} as const;

// Public links. The source link only renders when NEXT_PUBLIC_REPO_URL is set at build time
// (the repository is private for now, so there is nothing to link to yet).
export const REPO_URL = process.env.NEXT_PUBLIC_REPO_URL ?? "";

// Absolute site URL, needed for social-preview (Open Graph) image URLs. Optional.
export const SITE_URL = process.env.NEXT_PUBLIC_SITE_URL ?? "";
