import Image from "next/image";
import Link from "next/link";
import { ArrowUpRightIcon, HeartIcon, StarIcon } from "@/components/icons";
import { formatCurrency, formatNumber } from "@/lib/format";
import { categoryBySlug } from "@/lib/mock/categories";
import { developerByUsername } from "@/lib/mock/developers";
import type { Product } from "@/types";

export default function ProductCard({ product }: { product: Product }) {
  const author = developerByUsername[product.authorUsername];
  const category = categoryBySlug[product.category];

  return (
    <article className="group flex h-full flex-col">
      <Link
        href={`/products/${product.slug}`}
        className="relative block aspect-[16/11] overflow-hidden rounded-2xl border border-zinc-200 bg-zinc-100 shadow-sm transition duration-500 group-hover:-translate-y-1 group-hover:border-zinc-300 group-hover:shadow-xl group-hover:shadow-zinc-950/10 dark:border-white/10 dark:bg-zinc-900 dark:group-hover:border-white/20"
      >
        <Image
          src={product.coverImage}
          alt={`${product.name} product preview`}
          fill
          sizes="(max-width: 768px) 100vw, (max-width: 1280px) 50vw, 33vw"
          className="object-cover transition-transform duration-700 group-hover:scale-[1.025]"
        />
        <span className="absolute top-3 right-3 flex size-9 translate-y-1 items-center justify-center rounded-full border border-white/30 bg-white/85 text-zinc-900 opacity-0 shadow-sm backdrop-blur transition-all group-hover:translate-y-0 group-hover:opacity-100 dark:border-white/15 dark:bg-zinc-950/80 dark:text-white">
          <ArrowUpRightIcon className="size-4" />
        </span>
      </Link>
      <div className="flex flex-1 flex-col pt-5">
        <div className="mb-2 flex items-start justify-between gap-4">
          <div>
            <p className="mb-1 text-[11px] font-semibold uppercase tracking-[0.12em] text-zinc-500">
              {category.shortName}
            </p>
            <h3 className="text-[17px] font-semibold tracking-[-0.025em] text-zinc-950 dark:text-white">
              <Link href={`/products/${product.slug}`} className="hover:underline">
                {product.name}
              </Link>
            </h3>
          </div>
          <p className="mt-5 shrink-0 text-sm font-semibold text-zinc-950 dark:text-white">
            {product.price === 0 ? "Free" : formatCurrency(product.price)}
          </p>
        </div>
        <p className="line-clamp-2 text-sm leading-6 text-zinc-600 dark:text-zinc-400">
          {product.description}
        </p>
        <div className="mt-auto flex items-center justify-between gap-3 pt-5 text-xs text-zinc-500">
          <Link
            href={`/developers/${author.username}`}
            className="flex min-w-0 items-center gap-2 font-medium text-zinc-700 transition hover:text-zinc-950 dark:text-zinc-300 dark:hover:text-white"
          >
            <Image
              src={author.avatar}
              alt=""
              width={24}
              height={24}
              className="rounded-full ring-1 ring-zinc-950/10 dark:ring-white/15"
            />
            <span className="truncate">{author.name}</span>
          </Link>
          <div className="flex shrink-0 items-center gap-3">
            <span className="flex items-center gap-1" title={`${product.rating} rating`}>
              <StarIcon className="size-3.5 fill-current text-amber-500" />
              {product.rating}
            </span>
            <span className="flex items-center gap-1" title={`${product.favorites} favorites`}>
              <HeartIcon className="size-3.5" />
              {formatNumber(product.favorites)}
            </span>
            <span className="hidden xl:inline">{formatNumber(product.sales)} sales</span>
          </div>
        </div>
      </div>
    </article>
  );
}
