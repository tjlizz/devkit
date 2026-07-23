import Link from "next/link";
import { ArrowRightIcon } from "@/components/icons";

export default function NotFound() {
  return (
    <section className="mx-auto flex min-h-[65vh] max-w-2xl flex-col items-center justify-center px-6 text-center">
      <p className="text-xs font-semibold uppercase tracking-[0.2em] text-zinc-500">404</p>
      <h1 className="mt-5 text-4xl font-semibold tracking-[-0.05em] text-zinc-950 dark:text-white">
        This page shipped somewhere else.
      </h1>
      <p className="mt-4 text-base leading-7 text-zinc-600 dark:text-zinc-400">
        The product or profile may have moved. The marketplace still has plenty worth
        exploring.
      </p>
      <Link
        href="/search"
        className="mt-8 inline-flex items-center gap-2 rounded-full bg-zinc-950 px-5 py-3 text-sm font-semibold text-white dark:bg-white dark:text-zinc-950"
      >
        Explore products
        <ArrowRightIcon className="size-4" />
      </Link>
    </section>
  );
}
