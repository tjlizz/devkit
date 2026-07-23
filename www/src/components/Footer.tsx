import Link from "next/link";
import { ArrowUpRightIcon } from "@/components/icons";
import { Logo } from "@/components/ui/logo";
import { categories } from "@/lib/mock/categories";

export default function Footer() {
  return (
    <footer className="border-t border-zinc-200 bg-zinc-50/60 dark:border-white/10 dark:bg-white/[0.02]">
      <div className="mx-auto max-w-7xl px-5 py-14 sm:px-6 lg:px-8 lg:py-20">
        <div className="grid gap-12 lg:grid-cols-[1.5fr_1fr_1fr_1fr]">
          <div className="max-w-sm">
            <Logo />
            <p className="mt-5 text-sm leading-6 text-zinc-600 dark:text-zinc-400">
              Exceptional software, directly from the developers who build it.
              Discover useful products and support independent craft.
            </p>
          </div>
          <div>
            <p className="text-sm font-semibold text-zinc-950 dark:text-white">Marketplace</p>
            <ul className="mt-4 space-y-3">
              {categories.slice(0, 4).map((category) => (
                <li key={category.slug}>
                  <Link
                    href={`/category/${category.slug}`}
                    className="text-sm text-zinc-600 transition hover:text-zinc-950 dark:text-zinc-400 dark:hover:text-white"
                  >
                    {category.shortName}
                  </Link>
                </li>
              ))}
            </ul>
          </div>
          <div>
            <p className="text-sm font-semibold text-zinc-950 dark:text-white">Company</p>
            <ul className="mt-4 space-y-3 text-sm text-zinc-600 dark:text-zinc-400">
              {["About", "For developers", "Editorial", "Contact"].map((item) => (
                <li key={item}>
                  <Link href="/search" className="transition hover:text-zinc-950 dark:hover:text-white">
                    {item}
                  </Link>
                </li>
              ))}
            </ul>
          </div>
          <div>
            <p className="text-sm font-semibold text-zinc-950 dark:text-white">Follow</p>
            <ul className="mt-4 space-y-3">
              {["X / Twitter", "GitHub", "Newsletter"].map((item) => (
                <li key={item}>
                  <Link
                    href="/search"
                    className="inline-flex items-center gap-1 text-sm text-zinc-600 transition hover:text-zinc-950 dark:text-zinc-400 dark:hover:text-white"
                  >
                    {item}
                    <ArrowUpRightIcon className="size-3.5" />
                  </Link>
                </li>
              ))}
            </ul>
          </div>
        </div>
        <div className="mt-14 flex flex-col gap-4 border-t border-zinc-200 pt-6 text-xs text-zinc-500 sm:flex-row sm:items-center sm:justify-between dark:border-white/10">
          <p>© {new Date().getFullYear()} DevKit Marketplace. Made for people who build.</p>
          <div className="flex gap-5">
            <Link href="/search">Privacy</Link>
            <Link href="/search">Terms</Link>
            <Link href="/search">Seller policy</Link>
          </div>
        </div>
      </div>
    </footer>
  );
}
