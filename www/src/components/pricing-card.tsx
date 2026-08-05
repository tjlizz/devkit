import Link from "next/link";
import { BuyButton } from "@/components/buy-button";
import { CheckIcon, ShieldIcon } from "@/components/icons";
import { formatCurrency } from "@/lib/format";
import type { Product } from "@/types";

export function PricingCard({ product }: { product: Product }) {
  const plans = product.plans && product.plans.length > 0 ? product.plans : null;

  return (
    <aside className="rounded-2xl border border-zinc-200 bg-white p-6 shadow-xl shadow-zinc-950/5 dark:border-white/10 dark:bg-zinc-900 dark:shadow-black/20">
      {plans ? (
        <>
          <p className="text-xs font-semibold uppercase tracking-[0.14em] text-zinc-500">
            Choose your plan
          </p>
          <div className="mt-4 space-y-3">
            {plans.map((plan) => (
              <div
                key={plan.name}
                className="rounded-xl border border-zinc-200 p-5 dark:border-white/10"
              >
                <div className="flex items-center justify-between gap-3">
                  <p className="text-sm font-semibold text-zinc-950 dark:text-white">{plan.name}</p>
                  <div className="flex items-baseline gap-1">
                    <span className="text-2xl font-semibold tracking-[-0.04em] text-zinc-950 dark:text-white">
                      {plan.priceCents === 0 ? "Free" : formatCurrency(plan.priceCents / 100)}
                    </span>
                    {plan.priceCents > 0 ? <span className="text-xs text-zinc-500">USD</span> : null}
                  </div>
                </div>
                {plan.description ? (
                  <p className="mt-2 text-xs leading-5 text-zinc-500">{plan.description}</p>
                ) : null}
                {plan.features.length > 0 ? (
                  <ul className="mt-3 space-y-2 text-xs text-zinc-600 dark:text-zinc-400">
                    {plan.features.map((feature) => (
                      <li key={feature} className="flex gap-2">
                        <CheckIcon className="mt-0.5 size-3.5 shrink-0 text-emerald-600 dark:text-emerald-400" />
                        <span>{feature}</span>
                      </li>
                    ))}
                  </ul>
                ) : null}
                <BuyButton appSlug={product.slug} plan={plan} />
              </div>
            ))}
          </div>
        </>
      ) : (
        <>
          <p className="text-xs font-semibold uppercase tracking-[0.14em] text-zinc-500">
            {product.priceLabel}
          </p>
          <div className="mt-3 flex items-end gap-2">
            <span className="text-4xl font-semibold tracking-[-0.05em] text-zinc-950 dark:text-white">
              {product.price === 0 ? "Free" : formatCurrency(product.price)}
            </span>
            {product.price > 0 ? <span className="mb-1.5 text-xs text-zinc-500">USD</span> : null}
          </div>
          <Link
            href="#purchase"
            className="mt-6 flex h-12 w-full items-center justify-center rounded-xl bg-zinc-950 text-sm font-semibold text-white shadow-lg transition hover:-translate-y-0.5 hover:bg-zinc-800 dark:bg-white dark:text-zinc-950 dark:hover:bg-zinc-200"
          >
            {product.price === 0 ? "Get it free" : "Buy now"}
          </Link>
        </>
      )}
      <Link
        href={product.demoUrl}
        className="mt-2 flex h-11 w-full items-center justify-center rounded-xl border border-zinc-200 text-sm font-semibold text-zinc-700 transition hover:bg-zinc-50 dark:border-white/10 dark:text-zinc-200 dark:hover:bg-white/5"
      >
        View live demo
      </Link>
      {!plans ? (
        <>
          <div className="my-6 h-px bg-zinc-200 dark:bg-white/10" />
          <ul className="space-y-3 text-sm text-zinc-600 dark:text-zinc-400">
            {[product.license, product.delivery, product.support].map((item) => (
              <li key={item} className="flex gap-2.5">
                <CheckIcon className="mt-0.5 size-4 shrink-0 text-emerald-600 dark:text-emerald-400" />
                <span>{item}</span>
              </li>
            ))}
          </ul>
        </>
      ) : null}
      <div className="my-6 h-px bg-zinc-200 dark:bg-white/10" />
      <div className="mt-6 flex items-start gap-3 rounded-xl bg-zinc-50 p-4 dark:bg-white/5">
        <ShieldIcon className="mt-0.5 size-5 shrink-0 text-zinc-500" />
        <div>
          <p className="text-xs font-semibold text-zinc-800 dark:text-zinc-200">
            Buyer protection
          </p>
          <p className="mt-1 text-xs leading-5 text-zinc-500">
            Secure checkout and a 14-day product-quality guarantee.
          </p>
        </div>
      </div>
    </aside>
  );
}
