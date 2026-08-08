"use client"

import { useCallback, useEffect, useState } from "react"
import Link from "next/link"
import { useAuth, type DeveloperSale, type DeveloperSalesSummary } from "@/lib/auth-context"
import { formatCurrency, formatDate } from "@/lib/format"

function AuthGate({
  children,
}: {
  children: (user: NonNullable<ReturnType<typeof useAuth>["user"]>) => React.ReactNode
}) {
  const { isAuthenticated, user, loading: authLoading } = useAuth()

  if (authLoading) {
    return (
      <main className="mx-auto min-h-[60vh] max-w-5xl px-5 py-16 text-sm text-zinc-500">
        Loading...
      </main>
    )
  }

  if (!isAuthenticated) {
    return (
      <main className="mx-auto min-h-[60vh] max-w-lg px-5 py-16 text-center">
        <h1 className="text-2xl font-semibold tracking-tight">Sign in required</h1>
        <p className="mt-3 text-sm text-zinc-500">Sign in to view your sales.</p>
        <Link
          className="mt-6 inline-flex rounded-lg bg-zinc-950 px-5 py-2.5 text-sm font-semibold text-white dark:bg-white dark:text-zinc-950"
          href="/login"
        >
          Sign in
        </Link>
      </main>
    )
  }

  if (!user?.isDeveloper) {
    return (
      <main className="mx-auto min-h-[60vh] max-w-lg px-5 py-16 text-center">
        <h1 className="text-2xl font-semibold tracking-tight">Developer approval required</h1>
        <p className="mt-3 text-sm text-zinc-500">Submit your developer application before viewing sales.</p>
        <Link
          className="mt-6 inline-flex rounded-lg bg-zinc-950 px-5 py-2.5 text-sm font-semibold text-white dark:bg-white dark:text-zinc-950"
          href="/become-developer"
        >
          Apply to publish
        </Link>
      </main>
    )
  }

  return <>{children(user)}</>
}

export default function SalesPage() {
  return (
    <AuthGate>
      {() => <SalesContent />}
    </AuthGate>
  )
}

function SalesContent() {
  const { listDeveloperSales } = useAuth()
  const [sales, setSales] = useState<DeveloperSale[] | null>(null)
  const [summary, setSummary] = useState<DeveloperSalesSummary | null>(null)
  const [error, setError] = useState("")

  const load = useCallback(async () => {
    setError("")
    try {
      const data = await listDeveloperSales()
      setSales(data.sales)
      setSummary(data.summary)
    } catch (err: any) {
      setError(err.message || "Could not load sales")
      setSales([])
      setSummary(null)
    }
  }, [listDeveloperSales])

  useEffect(() => {
    load()
  }, [load])

  if (sales === null && !error) {
    return (
      <main className="mx-auto min-h-[60vh] max-w-5xl px-5 py-16 text-sm text-zinc-500">
        Loading your sales...
      </main>
    )
  }

  const empty = sales !== null && sales.length === 0

  return (
    <main className="mx-auto min-h-[60vh] max-w-5xl px-5 py-12 sm:py-16">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <Link href="/my-apps" className="text-sm text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-200">
            ← My apps
          </Link>
          <h1 className="mt-2 text-3xl font-semibold tracking-tight text-zinc-950 dark:text-white">Sales</h1>
          <p className="mt-3 text-sm text-zinc-500">
            Orders, buyers, and revenue across your published apps.
          </p>
        </div>
      </div>

      {error ? (
        <div className="mt-8 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-300">
          {error}
          <button onClick={load} className="ml-3 font-semibold underline underline-offset-2">
            Retry
          </button>
        </div>
      ) : null}

      {summary ? (
        <section className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div className="rounded-2xl border border-zinc-200 bg-white p-5 dark:border-white/10 dark:bg-zinc-950">
            <p className="text-xs font-medium uppercase tracking-[0.14em] text-zinc-500">Revenue</p>
            <p className="mt-2 text-2xl font-semibold tracking-tight text-zinc-950 dark:text-white">
              {formatCurrency(summary.totalRevenueCents / 100)}
            </p>
          </div>
          <div className="rounded-2xl border border-zinc-200 bg-white p-5 dark:border-white/10 dark:bg-zinc-950">
            <p className="text-xs font-medium uppercase tracking-[0.14em] text-zinc-500">Sales</p>
            <p className="mt-2 text-2xl font-semibold tracking-tight text-zinc-950 dark:text-white">
              {summary.totalOrders}
            </p>
          </div>
          <div className="rounded-2xl border border-zinc-200 bg-white p-5 dark:border-white/10 dark:bg-zinc-950">
            <p className="text-xs font-medium uppercase tracking-[0.14em] text-zinc-500">Buyers</p>
            <p className="mt-2 text-2xl font-semibold tracking-tight text-zinc-950 dark:text-white">
              {summary.uniqueBuyers}
            </p>
          </div>
          <div className="rounded-2xl border border-zinc-200 bg-white p-5 dark:border-white/10 dark:bg-zinc-950">
            <p className="text-xs font-medium uppercase tracking-[0.14em] text-zinc-500">Apps sold</p>
            <p className="mt-2 text-2xl font-semibold tracking-tight text-zinc-950 dark:text-white">
              {summary.appsSold}
            </p>
          </div>
        </section>
      ) : null}

      <section className="mt-8">
        <h2 className="text-sm font-semibold uppercase tracking-[0.14em] text-zinc-500">Orders</h2>
        {empty ? (
          <div className="mt-4 rounded-2xl border border-dashed border-zinc-300 px-6 py-16 text-center dark:border-white/15">
            <h3 className="text-lg font-semibold text-zinc-950 dark:text-white">No sales yet</h3>
            <p className="mx-auto mt-2 max-w-md text-sm leading-6 text-zinc-500">
              When a buyer completes a purchase of one of your apps, the order will show up here.
            </p>
            <Link
              href="/my-apps"
              className="mt-6 inline-flex rounded-full bg-zinc-950 px-5 py-2.5 text-sm font-semibold text-white dark:bg-white dark:text-zinc-950"
            >
              Manage your apps
            </Link>
          </div>
        ) : null}

        {sales && sales.length > 0 ? (
          <div className="mt-4 overflow-x-auto rounded-2xl border border-zinc-200 bg-white dark:border-white/10 dark:bg-zinc-950">
            <table className="w-full min-w-[640px] text-left text-sm">
              <thead>
                <tr className="border-b border-zinc-200 text-xs uppercase tracking-wide text-zinc-500 dark:border-white/10">
                  <th className="px-5 py-3 font-medium">App</th>
                  <th className="px-5 py-3 font-medium">Plan</th>
                  <th className="px-5 py-3 font-medium">Buyer</th>
                  <th className="px-5 py-3 font-medium">Amount</th>
                  <th className="px-5 py-3 font-medium">Paid</th>
                </tr>
              </thead>
              <tbody>
                {sales.map((sale) => (
                  <tr
                    key={sale.orderId}
                    className="border-b border-zinc-100 last:border-0 dark:border-white/5"
                  >
                    <td className="px-5 py-4">
                      <Link
                        href={`/products/${sale.appSlug}`}
                        className="font-semibold text-zinc-950 hover:underline dark:text-white"
                      >
                        {sale.appName}
                      </Link>
                    </td>
                    <td className="px-5 py-4 text-zinc-600 dark:text-zinc-300">
                      {sale.planName || "Standard license"}
                    </td>
                    <td className="px-5 py-4 text-zinc-600 dark:text-zinc-300">{sale.buyerEmail}</td>
                    <td className="px-5 py-4 font-semibold text-zinc-950 dark:text-white">
                      {formatCurrency(sale.priceCents / 100)}
                    </td>
                    <td className="px-5 py-4 text-zinc-500">
                      {sale.paidAt ? formatDate(sale.paidAt) : formatDate(sale.createdAt)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : null}
      </section>
    </main>
  )
}