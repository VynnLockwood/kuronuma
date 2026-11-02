import { Icon } from "@/components/icons";
import { buttonVariants } from "@/components/ui/button";
import { REPO_URL } from "@/lib/brand";
import { cn } from "@/lib/utils";

const names = { en: "Kuronuma on GitHub", th: "Kuronuma บน GitHub" } as const;

// A quiet outline button in the header that opens the public repository in a new tab. The word "GitHub" shows from
// the sm breakpoint up; below that it is the icon alone (a 36 px tap target; WCAG 2.2 AA asks for 24). Hidden when REPO_URL is empty.
export function RepoLink({ lang = "en" }: { lang?: keyof typeof names }) {
  if (!REPO_URL) return null;
  return (
    <a
      href={REPO_URL}
      target="_blank"
      rel="noopener noreferrer"
      aria-label={names[lang]}
      className={cn(buttonVariants({ variant: "outline", size: "sm" }), "size-9 gap-1.5 px-0 sm:size-auto sm:h-8 sm:px-3")}
    >
      <Icon name="github" size={18} />
      <span className="hidden sm:inline">GitHub</span>
    </a>
  );
}
