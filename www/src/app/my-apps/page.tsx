"use client"

import { useCallback, useEffect, useState } from "react"
import Link from "next/link"
import { AppStatusBadge } from "@/components/app-status-badge"
import { useAuth, type DeveloperApp } from "@/lib/auth-context"
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
        <p className="mt-3 text-sm text-zinc-500">Sign in to manage your published apps.</p>
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
        <p className="mt-3 text-sm text-zinc-500">Submit your developer application before publishing apps.</p>
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

export default function MyAppsPage() {
  return (
    <AuthGate>
      {(user) => <MyAppsContent key={user.id} />}
    </AuthGate>
  )
}

function MyAppsContent() {
  const { listDeveloperApps, delistDeveloperApp } = useAuth()
  const [apps, setApps] = useState<DeveloperApp[] | null>(null)
  const [error, setError] = useState("")
  const [delistingId, setDelistingId] = useState<number | null>(null)
  const [notice, setNotice] = useState("")

  const load = useCallback(async () => {
    setError("")
    try {
      setApps(await listDeveloperApps())
    } catch (err: any) {
      setError(err.message || "Could not load apps")
      setApps([])
    }
  }, [listDeveloperApps])

  useEffect(() => {
    load()
  }, [load])

  async function handleDelist(app: DeveloperApp) {
    if (!window.confirm(`Delist "${app.name}"? It will be removed from the marketplace immediately.`)) {
      return
    }
    setDelistingId(app.id)
    setNotice("")
    setError("")
    try {
      await delistDeveloperApp(app.id)
      setNotice(`"${app.name}" has been delisted.`)
      await load()
    } catch (err: any) {
      setError(err.message || "Could not delist app")
    } finally {
      setDelistingId(null)
    }
  }

  if (apps === null && !error) {
    return (
      <main className="mx-auto min-h-[60vh] max-w-5xl px-5 py-16 text-sm text-zinc-500">
        Loading your apps...
      </main>
    )
  }

  return (
    <main className="mx-auto min-h-[60vh] max-w-5xl px-5 py-12 sm:py-16">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight text-zinc-950 dark:text-white">My apps</h1>
          <p className="mt-3 text-sm text-zinc-500">Manage your published software and track review status.</p>
        </div>
        <Link
          href="/publish"
          className="inline-flex rounded-lg bg-zinc-950 px-4 py-2.5 text-sm font-semibold text-white transition hover:-translate-y-0.5 hover:shadow-lg dark:bg-white dark:text-zinc-950"
        >
          Publish new app
        </Link>
        <Link
          href="/my-apps/sales"
          className="inline-flex rounded-lg border border-zinc-200 px-4 py-2.5 text-sm font-medium text-zinc-800 transition hover:bg-zinc-100 dark:border-white/10 dark:text-zinc-200 dark:hover:bg-white/5"
        >
          Sales
        </Link>
      </div>

      {notice ? (
        <div className="mt-8 rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700 dark:border-emerald-800 dark:bg-emerald-950 dark:text-emerald-300">
          {notice}
        </div>
      ) : null}
      {error ? (
        <div className="mt-8 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-300">
          {error}
          <button
            onClick={load}
            className="ml-3 font-semibold underline underline-offset-2"
          >
            Retry
          </button>
        </div>
      ) : null}

      {apps && apps.length === 0 ? (
        <div className="mt-10 rounded-2xl border border-dashed border-zinc-300 px-6 py-20 text-center dark:border-white/15">
          <h2 className="text-lg font-semibold text-zinc-950 dark:text-white">No apps yet</h2>
          <p className="mx-auto mt-2 max-w-md text-sm leading-6 text-zinc-500">
            You haven&apos;t published any apps. Share your first software product with the marketplace.
          </p>
          <Link
            href="/publish"
            className="mt-6 inline-flex rounded-full bg-zinc-950 px-5 py-2.5 text-sm font-semibold text-white dark:bg-white dark:text-zinc-950"
          >
            Publish your first app
          </Link>
        </div>
      ) : null}

      {apps && apps.length > 0 ? (
        <ul className="mt-8 grid gap-4">
          {apps.map((app) => (
            <li
              key={app.id}
              className="flex flex-col gap-4 rounded-2xl border border-zinc-200 bg-white p-5 dark:border-white/10 dark:bg-zinc-950 sm:flex-row sm:items-center"
            >
              {app.iconUrl ? (
                <img
                  src={app.iconUrl}
                  alt=""
                  className="size-12 shrink-0 rounded-xl border border-zinc-200 object-cover dark:border-white/10"
                />
              ) : (
                <span className="flex size-12 shrink-0 items-center justify-center rounded-xl bg-zinc-100 text-lg font-semibold text-zinc-400 dark:bg-white/5">
                  {app.name.charAt(0).toUpperCase()}
                </span>
              )}
              <div className="min-w-0 flex-1">
                <div className="flex flex-wrap items-center gap-2">
                  <h2 className="truncate text-base font-semibold text-zinc-950 dark:text-white">
                    {app.name}
                  </h2>
                  <AppStatusBadge status={app.status} />
                </div>
                <p className="mt-1 truncate text-sm text-zinc-500">{app.tagline}</p>
                {app.status === "rejected" && app.reviewNote ? (
                  <p className="mt-2 rounded-lg bg-red-50 px-3 py-2 text-xs leading-5 text-red-700 dark:bg-red-950/50 dark:text-red-300">
                    Review note: {app.reviewNote}
                  </p>
                ) : null}
              </div>
              <div className="flex shrink-0 flex-wrap items-center gap-x-4 gap-y-1 text-sm text-zinc-500 sm:flex-col sm:items-end">
                <span>{formatCurrency(app.priceCents / 100)}</span>
                <span className="text-xs">Updated {formatDate(app.updatedAt)}</span>
              </div>
              <div className="flex shrink-0 flex-wrap items-center gap-2">
                {app.status === "approved" ? (
                  <Link
                    href={`/products/${app.slug}`}
                    className="rounded-lg border border-zinc-200 px-3 py-1.5 text-sm font-medium transition hover:bg-zinc-100 dark:border-white/10 dark:hover:bg-white/5"
                  >
                    View
                  </Link>
                ) : null}
                {app.status !== "delisted" ? (
                  <Link
                    href={`/my-apps/${app.id}/edit`}
                    className="rounded-lg border border-zinc-200 px-3 py-1.5 text-sm font-medium transition hover:bg-zinc-100 dark:border-white/10 dark:hover:bg-white/5"
                  >
                    Edit
                  </Link>
                ) : null}
                <Link
                  href={`/my-apps/${app.id}/deliverables`}
                  className="rounded-lg border border-zinc-200 px-3 py-1.5 text-sm font-medium transition hover:bg-zinc-100 dark:border-white/10 dark:hover:bg-white/5"
                >
                  Deliverables
                </Link>
                {app.status === "approved" ? (
                  <button
                    onClick={() => handleDelist(app)}
                    disabled={delistingId === app.id}
                    className="rounded-lg border border-red-200 px-3 py-1.5 text-sm font-medium text-red-600 transition hover:bg-red-50 disabled:opacity-50 dark:border-red-900 dark:text-red-400 dark:hover:bg-red-950/50"
                  >
                    {delistingId === app.id ? "Delisting..." : "Delist"}
                  </button>
                ) : null}
              </div>
            </li>
          ))}
        </ul>
      ) : null}
    </main>
  )
}
