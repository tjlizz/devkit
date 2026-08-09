"use client"

import { useCallback, useEffect, useState } from "react"
import Link from "next/link"
import { useSearchParams } from "next/navigation"
import { useAuth, type Delivery, type Entitlement, type Order } from "@/lib/auth-context"
import { formatBytes, formatCurrency, formatDate } from "@/lib/format"

const API_BASE = process.env.NEXT_PUBLIC_API_URL || ""

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

  if (!isAuthenticated || !user) {
    return (
      <main className="mx-auto min-h-[60vh] max-w-lg px-5 py-16 text-center">
        <h1 className="text-2xl font-semibold tracking-tight">Sign in required</h1>
        <p className="mt-3 text-sm text-zinc-500">Sign in to view your purchases and entitlements.</p>
        <Link
          className="mt-6 inline-flex rounded-lg bg-zinc-950 px-5 py-2.5 text-sm font-semibold text-white dark:bg-white dark:text-zinc-950"
          href="/login"
        >
          Sign in
        </Link>
      </main>
    )
  }

  return <>{children(user)}</>
}

const orderStatusLabel: Record<string, string> = {
  pending: "Pending payment",
  paid: "Paid",
  refunded: "Refunded",
  cancelled: "Cancelled",
}

export default function PurchasesPage() {
  return (
    <AuthGate>
      {() => <PurchasesContent />}
    </AuthGate>
  )
}

function PurchasesContent() {
  const { listMyOrders, listMyEntitlements, getDelivery, refundOrder } = useAuth()
  const searchParams = useSearchParams()
  const [orders, setOrders] = useState<Order[] | null>(null)
  const [entitlements, setEntitlements] = useState<Entitlement[] | null>(null)
  const [error, setError] = useState("")
  const [notice, setNotice] = useState(searchParams.get("notice") === "paid" ? "Payment confirmed — your entitlement is active." : "")
  const [deliveryByEnt, setDeliveryByEnt] = useState<Record<number, Delivery>>({})
  const [loadingDelivery, setLoadingDelivery] = useState<number | null>(null)
  const [refundingId, setRefundingId] = useState<number | null>(null)

  const load = useCallback(async () => {
    setError("")
    try {
      const [orderList, entList] = await Promise.all([listMyOrders(), listMyEntitlements()])
      setOrders(orderList)
      setEntitlements(entList)
    } catch (err: any) {
      setError(err.message || "Could not load purchases")
      setOrders([])
      setEntitlements([])
    }
  }, [listMyOrders, listMyEntitlements])

  useEffect(() => {
    load()
  }, [load])

  async function handleDelivery(ent: Entitlement) {
    setLoadingDelivery(ent.id)
    setError("")
    try {
      const delivery = await getDelivery(ent.id)
      setDeliveryByEnt((prev) => ({ ...prev, [ent.id]: delivery }))
    } catch (err: any) {
      setError(err.message || "Could not get delivery")
    } finally {
      setLoadingDelivery(null)
    }
  }

  async function handleRefund(orderId: number) {
    if (!window.confirm("Request a refund for this order? Your access to the app will be revoked immediately.")) {
      return
    }
    setRefundingId(orderId)
    setError("")
    setNotice("")
    try {
      await refundOrder(orderId)
      setNotice("Refund processed. Your access has been revoked.")
      await load()
    } catch (err: any) {
      setError(err.message || "Could not request refund")
    } finally {
      setRefundingId(null)
    }
  }

  if (orders === null || entitlements === null) {
    return (
      <main className="mx-auto min-h-[60vh] max-w-5xl px-5 py-16 text-sm text-zinc-500">
        Loading your purchases...
      </main>
    )
  }

  const activeEntitlements = entitlements.filter((ent) => ent.status === "active")

  return (
    <main className="mx-auto min-h-[60vh] max-w-5xl px-5 py-12 sm:py-16">
      <div>
        <h1 className="text-3xl font-semibold tracking-tight text-zinc-950 dark:text-white">My purchases</h1>
        <p className="mt-3 text-sm text-zinc-500">Your orders, entitlements, and app deliveries.</p>
      </div>

      {notice ? (
        <div className="mt-8 rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700 dark:border-emerald-800 dark:bg-emerald-950 dark:text-emerald-300">
          {notice}
        </div>
      ) : null}
      {error ? (
        <div className="mt-8 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-300">
          {error}
        </div>
      ) : null}

      {activeEntitlements.length === 0 && orders.length === 0 ? (
        <div className="mt-10 rounded-2xl border border-dashed border-zinc-300 px-6 py-20 text-center dark:border-white/15">
          <h2 className="text-lg font-semibold text-zinc-950 dark:text-white">No purchases yet</h2>
          <p className="mx-auto mt-2 max-w-md text-sm leading-6 text-zinc-500">
            When you buy an app, your order and access will show up here.
          </p>
          <Link
            href="/search"
            className="mt-6 inline-flex rounded-full bg-zinc-950 px-5 py-2.5 text-sm font-semibold text-white dark:bg-white dark:text-zinc-950"
          >
            Explore products
          </Link>
        </div>
      ) : null}

      {activeEntitlements.length > 0 ? (
        <section className="mt-10">
          <h2 className="text-sm font-semibold uppercase tracking-[0.14em] text-zinc-500">
            Your apps
          </h2>
          <ul className="mt-4 grid gap-4">
            {activeEntitlements.map((ent) => {
              const delivery = deliveryByEnt[ent.id]
              return (
                <li
                  key={ent.id}
                  className="rounded-2xl border border-zinc-200 bg-white p-5 dark:border-white/10 dark:bg-zinc-950"
                >
                  <div className="flex flex-wrap items-start justify-between gap-4">
                    <div className="min-w-0">
                      <Link
                        href={`/products/${ent.appSlug}`}
                        className="text-base font-semibold text-zinc-950 hover:underline dark:text-white"
                      >
                        {ent.appName}
                      </Link>
                      <p className="mt-1 text-sm text-zinc-500">
                        {ent.planName ? `${ent.planName} plan` : "Standard license"} · Granted {formatDate(ent.grantedAt)}
                      </p>
                    </div>
                    <button
                      type="button"
                      onClick={() => handleDelivery(ent)}
                      disabled={loadingDelivery === ent.id}
                      className="rounded-lg bg-zinc-950 px-4 py-2 text-sm font-semibold text-white transition hover:bg-zinc-800 disabled:opacity-60 dark:bg-white dark:text-zinc-950 dark:hover:bg-zinc-200"
                    >
                      {loadingDelivery === ent.id ? "Loading..." : delivery ? "Refresh access" : "Access app"}
                    </button>
                  </div>

                  {delivery ? (
                    <div className="mt-4 space-y-3 rounded-xl bg-zinc-50 p-4 text-sm dark:bg-white/5">
                      <p className="text-xs text-zinc-500">
                        Access token expires {formatDate(delivery.expiresAt)}
                      </p>
                      {delivery.artifacts && delivery.artifacts.length > 0 ? (
                        <div className="space-y-2">
                          <p className="text-xs font-medium uppercase tracking-wide text-zinc-400">
                            Downloads
                          </p>
                          <ul className="space-y-2">
                            {delivery.artifacts.map((artifact) => (
                              <li
                                key={artifact.id}
                                className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-zinc-200 bg-white px-3 py-2 dark:border-white/10 dark:bg-zinc-900"
                              >
                                <div className="min-w-0">
                                  <p className="truncate font-medium text-zinc-800 dark:text-zinc-200">
                                    {artifact.fileName}
                                  </p>
                                  <p className="text-xs text-zinc-400">
                                    {formatBytes(artifact.sizeBytes)}
                                  </p>
                                </div>
                                <a
                                  href={`${API_BASE}/api/v1/artifacts/${artifact.id}/download?token=${delivery.deliveryToken}`}
                                  className="rounded-lg bg-zinc-950 px-3 py-1.5 text-sm font-semibold text-white transition hover:bg-zinc-800 dark:bg-white dark:text-zinc-950 dark:hover:bg-zinc-200"
                                >
                                  Download
                                </a>
                              </li>
                            ))}
                          </ul>
                        </div>
                      ) : null}
                      <div className="flex flex-wrap gap-2 pt-1">
                        {delivery.sourceUrl ? (
                          <a
                            href={delivery.sourceUrl}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="rounded-lg border border-zinc-200 px-3 py-1.5 font-medium text-zinc-800 transition hover:bg-zinc-100 dark:border-white/10 dark:text-zinc-200 dark:hover:bg-white/5"
                          >
                            Source
                          </a>
                        ) : null}
                        {delivery.docsUrl ? (
                          <a
                            href={delivery.docsUrl}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="rounded-lg border border-zinc-200 px-3 py-1.5 font-medium text-zinc-800 transition hover:bg-zinc-100 dark:border-white/10 dark:text-zinc-200 dark:hover:bg-white/5"
                          >
                            Docs
                          </a>
                        ) : null}
                        {delivery.demoUrl ? (
                          <a
                            href={delivery.demoUrl}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="rounded-lg border border-zinc-200 px-3 py-1.5 font-medium text-zinc-800 transition hover:bg-zinc-100 dark:border-white/10 dark:text-zinc-200 dark:hover:bg-white/5"
                          >
                            Live demo
                          </a>
                        ) : null}
                        <span className="ml-auto text-xs text-zinc-400">v{delivery.version}</span>
                      </div>
                    </div>
                  ) : null}
                </li>
              )
            })}
          </ul>
        </section>
      ) : null}

      {orders.length > 0 ? (
        <section className="mt-12">
          <h2 className="text-sm font-semibold uppercase tracking-[0.14em] text-zinc-500">
            Order history
          </h2>
          <ul className="mt-4 grid gap-3">
            {orders.map((order) => (
              <li
                key={order.id}
                className="flex flex-wrap items-center gap-4 rounded-2xl border border-zinc-200 bg-white p-5 text-sm dark:border-white/10 dark:bg-zinc-950"
              >
                <div className="min-w-0 flex-1">
                  <p className="font-semibold text-zinc-950 dark:text-white">{order.appName}</p>
                  <p className="mt-0.5 text-xs text-zinc-500">
                    {order.planName || "Standard"} · {formatDate(order.createdAt)}
                  </p>
                </div>
                <span className="font-medium text-zinc-800 dark:text-zinc-200">
                  {formatCurrency(order.priceCents / 100)}
                </span>
                <span
                  className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ring-1 ring-inset ${
                    order.status === "paid"
                      ? "bg-emerald-50 text-emerald-700 ring-emerald-600/20 dark:bg-emerald-950/50 dark:text-emerald-300"
                      : order.status === "refunded"
                        ? "bg-zinc-100 text-zinc-600 ring-zinc-500/20 dark:bg-white/10 dark:text-zinc-300"
                        : "bg-amber-50 text-amber-700 ring-amber-600/20 dark:bg-amber-950/50 dark:text-amber-300"
                  }`}
                >
                  {orderStatusLabel[order.status] ?? order.status}
                </span>
                {order.status === "paid" ? (
                  <button
                    type="button"
                    onClick={() => handleRefund(order.id)}
                    disabled={refundingId === order.id}
                    className="rounded-lg border border-zinc-200 px-3 py-1.5 text-sm font-medium text-zinc-700 transition hover:bg-zinc-100 disabled:opacity-50 dark:border-white/10 dark:text-zinc-300 dark:hover:bg-white/5"
                  >
                    {refundingId === order.id ? "Refunding..." : "Request refund"}
                  </button>
                ) : null}
              </li>
            ))}
          </ul>
        </section>
      ) : null}
    </main>
  )
}
