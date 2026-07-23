import { SearchIcon, SlidersIcon } from "@/components/icons";
import { categories } from "@/lib/mock/categories";

interface SearchBarProps {
  query?: string;
  category?: string;
  type?: string;
}

export function SearchBar({ query = "", category = "", type = "all" }: SearchBarProps) {
  return (
    <form
      action="/search"
      className="rounded-2xl border border-zinc-200 bg-white p-3 shadow-sm dark:border-white/10 dark:bg-white/[0.025]"
    >
      <div className="flex flex-col gap-3 lg:flex-row">
        <label className="flex min-w-0 flex-1 items-center gap-3 rounded-xl bg-zinc-100 px-4 dark:bg-white/5">
          <SearchIcon className="size-5 shrink-0 text-zinc-400" />
          <span className="sr-only">Search marketplace</span>
          <input
            type="search"
            name="q"
            defaultValue={query}
            placeholder="Search products, developers, or tags..."
            className="h-12 w-full bg-transparent text-sm outline-none placeholder:text-zinc-400"
          />
        </label>
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-[minmax(160px,1fr)_minmax(160px,1fr)_auto]">
          <label className="relative">
            <span className="sr-only">Result type</span>
            <select
              name="type"
              defaultValue={type}
              className="h-12 w-full appearance-none rounded-xl border border-zinc-200 bg-white px-4 text-sm font-medium outline-none transition focus:border-zinc-400 dark:border-white/10 dark:bg-zinc-900"
            >
              <option value="all">All results</option>
              <option value="products">Products</option>
              <option value="developers">Developers</option>
              <option value="tags">Tags</option>
            </select>
          </label>
          <label className="relative">
            <span className="sr-only">Category</span>
            <select
              name="category"
              defaultValue={category}
              className="h-12 w-full appearance-none rounded-xl border border-zinc-200 bg-white px-4 text-sm font-medium outline-none transition focus:border-zinc-400 dark:border-white/10 dark:bg-zinc-900"
            >
              <option value="">All categories</option>
              {categories.map((item) => (
                <option key={item.slug} value={item.slug}>
                  {item.shortName}
                </option>
              ))}
            </select>
          </label>
          <button
            type="submit"
            className="col-span-2 inline-flex h-12 items-center justify-center gap-2 rounded-xl bg-zinc-950 px-6 text-sm font-semibold text-white transition hover:bg-zinc-800 sm:col-span-1 dark:bg-white dark:text-zinc-950 dark:hover:bg-zinc-200"
          >
            <SlidersIcon className="size-4" />
            Search
          </button>
        </div>
      </div>
    </form>
  );
}
