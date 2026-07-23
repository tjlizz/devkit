import Link from "next/link";
import { SearchIcon } from "@/components/icons";

export function EmptyState({ query }: { query?: string }) {
  return (
    <div className="rounded-2xl border border-dashed border-zinc-300 px-6 py-20 text-center dark:border-white/15">
      <span className="mx-auto flex size-12 items-center justify-center rounded-2xl bg-zinc-100 text-zinc-500 dark:bg-white/5">
        <SearchIcon className="size-5" />
      </span>
      <h2 className="mt-5 text-lg font-semibold text-zinc-950 dark:text-white">
        No results found
      </h2>
      <p className="mx-auto mt-2 max-w-md text-sm leading-6 text-zinc-500">
        We couldn&apos;t find anything{query ? ` matching “${query}”` : ""}. Try a
        broader term or explore every product.
      </p>
      <Link
        href="/search"
        className="mt-6 inline-flex rounded-full bg-zinc-950 px-5 py-2.5 text-sm font-semibold text-white dark:bg-white dark:text-zinc-950"
      >
        Clear search
      </Link>
    </div>
  );
}
