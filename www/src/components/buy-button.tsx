"use client"

import { useCallback, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth-context";
import type { AppPlan } from "@/types";

interface BuyButtonProps {
  appSlug: string;
  plan: AppPlan;
}

export function BuyButton({ appSlug, plan }: BuyButtonProps) {
  const { isAuthenticated, checkoutApp, confirmPayment } = useAuth();
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const free = plan.priceCents === 0;

  const handleBuy = useCallback(async () => {
    setError("");
    setBusy(true);
    try {
      const order = await checkoutApp(appSlug, plan.id);
      // Server-side confirm: entitlement is issued from the confirmed
      // payment state, never from a browser redirect alone.
      await confirmPayment(order.id);
      router.push("/purchases?notice=paid");
    } catch (err: any) {
      if (err?.status === 401) {
        router.push(`/login?next=/products/${appSlug}`);
        return;
      }
      setError(err?.message || "Checkout failed. Please try again.");
    } finally {
      setBusy(false);
    }
  }, [appSlug, plan.id, checkoutApp, confirmPayment, router]);

  if (!isAuthenticated) {
    return (
      <Link
        href={`/login?next=/products/${appSlug}`}
        className="flex h-10 w-full items-center justify-center rounded-lg bg-zinc-950 text-xs font-semibold text-white transition hover:bg-zinc-800 dark:bg-white dark:text-zinc-950 dark:hover:bg-zinc-200"
      >
        {free ? "Get it free" : "Buy now"}
      </Link>
    );
  }

  return (
    <div>
      <button
        type="button"
        onClick={handleBuy}
        disabled={busy}
        className="flex h-10 w-full items-center justify-center rounded-lg bg-zinc-950 text-xs font-semibold text-white transition hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-white dark:text-zinc-950 dark:hover:bg-zinc-200"
      >
        {busy ? "Processing..." : free ? "Get it free" : "Buy now"}
      </button>
      {error ? (
        <p className="mt-2 text-xs leading-5 text-red-600 dark:text-red-400">{error}</p>
      ) : null}
    </div>
  );
}
