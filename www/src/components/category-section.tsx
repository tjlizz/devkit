import Link from "next/link";
import { ArrowUpRightIcon } from "@/components/icons";
import { cn } from "@/lib/cn";
import { categories } from "@/lib/mock/categories";

export function CategorySection() {
  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
      {categories.map((category, index) => (
        <Link
          key={category.slug}
          href={`/category/${category.slug}`}
          className={cn(
            "group relative min-h-44 overflow-hidden rounded-2xl border border-zinc-200 bg-gradient-to-br p-6 transition duration-300 hover:-translate-y-1 hover:border-zinc-300 hover:shadow-lg dark:border-white/10 dark:hover:border-white/20",
            category.accent,
            index === 0 && "lg:col-span-2",
          )}
        >
          <div className="flex items-start justify-between">
            <span className="flex size-10 items-center justify-center rounded-xl border border-zinc-950/10 bg-white/80 font-mono text-xs font-bold text-zinc-800 shadow-sm backdrop-blur dark:border-white/10 dark:bg-zinc-950/70 dark:text-white">
              {category.icon}
            </span>
            <ArrowUpRightIcon className="size-4 text-zinc-400 transition duration-300 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 group-hover:text-zinc-950 dark:group-hover:text-white" />
          </div>
          <h3 className="mt-8 text-base font-semibold tracking-tight text-zinc-950 dark:text-white">
            {category.name}
          </h3>
          <p className="mt-1 text-xs text-zinc-500">{category.productCount} products</p>
        </Link>
      ))}
      <Link
        href="/search"
        className="group flex min-h-44 flex-col justify-between rounded-2xl bg-zinc-950 p-6 text-white transition duration-300 hover:-translate-y-1 hover:bg-zinc-800 hover:shadow-xl dark:bg-white dark:text-zinc-950 dark:hover:bg-zinc-200"
      >
        <span className="text-xs font-medium text-zinc-400 dark:text-zinc-500">Browse everything</span>
        <div className="flex items-end justify-between gap-4">
          <p className="text-lg font-semibold tracking-tight">Explore the full marketplace</p>
          <ArrowUpRightIcon className="size-5 shrink-0 transition-transform duration-300 group-hover:translate-x-1 group-hover:-translate-y-1" />
        </div>
      </Link>
    </div>
  );
}
