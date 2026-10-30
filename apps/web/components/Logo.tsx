import Link from "next/link";
import { BRAND } from "@/lib/brand";

// The mark is a vector simplification of the full logo illustration, readable at small sizes.
export function Logo({ href = "/" }: { href?: string }) {
  return (
    <Link href={href} className="logo" aria-label={`${BRAND.name} home`}>
      <img src="/brand/nemo-mark.svg" width={36} height={36} alt="" />
      <span className="logo-word">{BRAND.name}</span>
    </Link>
  );
}
