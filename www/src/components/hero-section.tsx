import Link from "next/link";
import { ArrowRightIcon, CheckIcon, SearchIcon, SparklesIcon } from "@/components/icons";

const signals = [
  { value: "1,200+", label: "curated products" },
  { value: "48k", label: "happy customers" },
  { value: "$8.4m", label: "earned by makers" },
];

export function HeroSection() {
  return (
    <section className="relative overflow-hidden border-b border-zinc-200 dark:border-white/10">
      <div className="hero-grid absolute inset-0 opacity-60 dark:opacity-25" />
      <div className="pointer-events-none absolute top-[-10rem] left-1/2 h-[36rem] w-[50rem] -translate-x-1/2 rounded-full bg-violet-400/15 blur-3xl dark:bg-violet-500/10" />
      <div className="relative mx-auto max-w-7xl px-5 pt-20 pb-16 sm:px-6 sm:pt-28 sm:pb-20 lg:px-8 lg:pt-36">
        <div className="mx-auto max-w-4xl text-center">
          <div className="mb-7 inline-flex items-center gap-2 rounded-full border border-zinc-200 bg-white/80 px-3 py-1.5 text-xs font-medium text-zinc-600 shadow-sm backdrop-blur dark:border-white/10 dark:bg-white/5 dark:text-zinc-300">
            <SparklesIcon className="size-3.5 text-violet-500" />
            Independent software, thoughtfully curated
          </div>
          <h1 className="text-balance text-[2.7rem] leading-[1.03] font-semibold tracking-[-0.06em] text-zinc-950 sm:text-6xl lg:text-[4.65rem] dark:text-white">
            Discover and Buy
            <br />
            <span className="bg-gradient-to-r from-zinc-950 via-zinc-500 to-zinc-950 bg-[length:200%_100%] bg-clip-text text-transparent dark:from-white dark:via-zinc-500 dark:to-white">
              Developer-Created Software
            </span>
          </h1>
          <p className="mx-auto mt-7 max-w-2xl text-balance text-base leading-7 text-zinc-600 sm:text-lg dark:text-zinc-400">
            Exceptional SaaS, AI apps, tools, templates, and APIs—built and supported
            by the developers who know them best.
          </p>
          <div className="mt-9 flex flex-col items-center justify-center gap-3 sm:flex-row">
            <Link
              href="/search"
              className="inline-flex h-12 w-full items-center justify-center gap-2 rounded-full bg-zinc-950 px-6 text-sm font-semibold text-white shadow-lg shadow-zinc-950/15 transition hover:-translate-y-0.5 hover:bg-zinc-800 hover:shadow-xl sm:w-auto dark:bg-white dark:text-zinc-950 dark:hover:bg-zinc-200"
            >
              Explore marketplace
              <ArrowRightIcon className="size-4" />
            </Link>
            <Link
              href="/search"
              className="inline-flex h-12 w-full items-center justify-center rounded-full border border-zinc-200 bg-white/80 px-6 text-sm font-semibold text-zinc-800 shadow-sm backdrop-blur transition hover:-translate-y-0.5 hover:border-zinc-300 hover:bg-white sm:w-auto dark:border-white/10 dark:bg-white/5 dark:text-white dark:hover:bg-white/10"
            >
              Start selling
            </Link>
          </div>
          <form action="/search" className="mx-auto mt-10 max-w-2xl" role="search">
            <div className="flex items-center rounded-2xl border border-zinc-200 bg-white p-2 shadow-xl shadow-zinc-950/5 transition focus-within:border-zinc-400 focus-within:ring-4 focus-within:ring-zinc-950/5 dark:border-white/10 dark:bg-zinc-900 dark:shadow-black/20 dark:focus-within:border-white/25">
              <SearchIcon className="ml-3 size-5 shrink-0 text-zinc-400" />
              <input
                name="q"
                aria-label="Search the marketplace"
                placeholder="Search analytics, AI, templates..."
                className="h-11 min-w-0 flex-1 bg-transparent px-3 text-sm outline-none placeholder:text-zinc-400"
              />
              <button
                type="submit"
                className="hidden h-10 rounded-xl bg-zinc-100 px-4 text-xs font-semibold text-zinc-700 transition hover:bg-zinc-200 sm:block dark:bg-white/10 dark:text-zinc-200 dark:hover:bg-white/15"
              >
                Search
              </button>
            </div>
          </form>
          <div className="mt-5 flex flex-wrap justify-center gap-x-5 gap-y-2 text-xs text-zinc-500">
            {["Verified creators", "Secure checkout", "Lifetime access"].map((item) => (
              <span key={item} className="inline-flex items-center gap-1.5">
                <CheckIcon className="size-3.5 text-emerald-600 dark:text-emerald-400" />
                {item}
              </span>
            ))}
          </div>
        </div>
        <div className="mx-auto mt-16 grid max-w-3xl grid-cols-3 divide-x divide-zinc-200 border-y border-zinc-200 py-6 dark:divide-white/10 dark:border-white/10">
          {signals.map((signal) => (
            <div key={signal.label} className="px-2 text-center sm:px-6">
              <p className="text-lg font-semibold tracking-tight text-zinc-950 sm:text-2xl dark:text-white">
                {signal.value}
              </p>
              <p className="mt-1 text-[10px] text-zinc-500 sm:text-xs">{signal.label}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
