import Image from "next/image";
import Link from "next/link";
import { ArrowRightIcon } from "@/components/icons";
import { formatNumber } from "@/lib/format";
import type { Developer } from "@/types";

export default function DeveloperCard({ developer }: { developer: Developer }) {
  return (
    <article className="group flex flex-col rounded-2xl border border-zinc-200 bg-white p-6 transition duration-300 hover:-translate-y-1 hover:border-zinc-300 hover:shadow-xl hover:shadow-zinc-950/5 dark:border-white/10 dark:bg-white/[0.025] dark:hover:border-white/20">
      <div className="flex items-start justify-between gap-4">
        <Image
          src={developer.avatar}
          alt={`${developer.name} avatar`}
          width={52}
          height={52}
          className="rounded-2xl shadow-sm ring-1 ring-zinc-950/10 dark:ring-white/15"
        />
        <span className="rounded-full bg-emerald-500/10 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-wider text-emerald-700 dark:text-emerald-400">
          Verified
        </span>
      </div>
      <h3 className="mt-5 text-base font-semibold tracking-tight text-zinc-950 dark:text-white">
        <Link href={`/developers/${developer.username}`}>{developer.name}</Link>
      </h3>
      <p className="mt-1 text-xs text-zinc-500">@{developer.username}</p>
      <p className="mt-4 min-h-12 text-sm leading-6 text-zinc-600 dark:text-zinc-400">
        {developer.bio}
      </p>
      <div className="mt-6 grid grid-cols-2 border-y border-zinc-200 py-4 dark:border-white/10">
        <div>
          <p className="text-base font-semibold text-zinc-950 dark:text-white">
            {developer.publishedCount}
          </p>
          <p className="mt-0.5 text-xs text-zinc-500">Products</p>
        </div>
        <div className="border-l border-zinc-200 pl-5 dark:border-white/10">
          <p className="text-base font-semibold text-zinc-950 dark:text-white">
            {formatNumber(developer.followers)}
          </p>
          <p className="mt-0.5 text-xs text-zinc-500">Followers</p>
        </div>
      </div>
      <Link
        href={`/developers/${developer.username}`}
        className="mt-5 inline-flex items-center gap-2 text-sm font-medium text-zinc-700 transition-colors group-hover:text-zinc-950 dark:text-zinc-300 dark:group-hover:text-white"
      >
        View profile
        <ArrowRightIcon className="size-4 transition-transform group-hover:translate-x-1" />
      </Link>
    </article>
  );
}
