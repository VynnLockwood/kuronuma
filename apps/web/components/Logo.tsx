import Link from "next/link";
import { BRAND } from "@/lib/brand";

// The full logo illustration, shown at 48px: large enough in the 72px header for its clouds and wings to read.
export function Logo({ href = "/" }: { href?: string }) {
  return (
    <Link href={href} className="logo" aria-label={`${BRAND.name} home`}>
      <img src="/brand/nemo-logo.png" width={48} height={48} alt="" />
      <span className="logo-word">{BRAND.name}</span>
    </Link>
  );
}
