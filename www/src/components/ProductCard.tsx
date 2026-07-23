import Link from "next/link";
import type { Product } from "@/types";
import { formatCurrency, formatNumber } from "@/lib/format";

export default function ProductCard({ product }: { product: Product }) {
  return (
    <Link
      href={`/products/${product.slug}`}
      className="group block overflow-hidden rounded-xl border border-surface-200 bg-white transition-all hover:shadow-lg hover:-translate-y-0.5"
    >
      <div className="aspect-[16/10] overflow-hidden bg-surface-100">
        <img
          src={product.coverImage}
          alt={product.name}
          className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-[1.02]"
        />
      </div>
      <div className="p-5">
        <div className="flex items-center gap-2">
          <span className="rounded-md bg-surface-100 px-2 py-0.5 text-[11px] font-medium uppercase tracking-wider text-surface-500">
            {product.category.replace("-", " & ")}
          </span>
          <span className="text-xs text-surface-400">&middot;</span>
          <span className="text-xs text-surface-400">{product.sales} sales</span>
        </div>
        <h3 className="mt-2 text-base font-semibold text-surface-900 group-hover:text-brand-600 transition-colors">
          {product.name}
        </h3>
        <p className="mt-1 text-sm leading-relaxed text-surface-500 line-clamp-2">
          {product.description}
        </p>
        <div className="mt-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="h-5 w-5 rounded-full bg-surface-300" />
            <span className="text-xs text-surface-500">{product.authorUsername}</span>
          </div>
          <div className="text-right">
            <span className="text-sm font-semibold text-surface-900">{formatCurrency(product.price)}</span>
            {product.priceLabel && (
              <span className="ml-0.5 text-[11px] text-surface-400">/{product.priceLabel.replace("per ", "")}</span>
            )}
          </div>
        </div>
      </div>
    </Link>
  );
}
