"use client"

import { useCallback, useEffect, useState } from "react"
import Image from "next/image"
import Link from "next/link"
import { useAuth, type FavoriteApp } from "@/lib/auth-context"
import { formatCurrency } from "@/lib/format"

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
        <p className="mt-3 text-sm text-zinc-500">Sign in to view apps you have saved.</p>
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

export default function FavoritesPage() {
  return (
    <AuthGate>
      {() => <FavoritesContent />}
    </AuthGate>
  )
}

function FavoritesContent() {
  const { listMyFavorites } = useAuth()
  const [apps, setApps] = useState<FavoriteApp[] | null>(null)
  const [error, setError] = useState("")

  const load = useCallback(async () => {
    setError("")
    try {
      const list = await listMyFavorites()
      setApps(list)
    } catch (err: any) {
      setError(err.message || "Could not load favorites")
      setApps([])
    }
  }, [listMyFavorites])

  useEffect(() => {
    load()
  }, [load])

  if (apps === null) {
    return (
      <main className="mx-auto min-h-[60vh] max-w-5xl px-5 py-16 text-sm text-zinc-500">
        Loading your favorites...
      </main>
    )
  }

  return (
    <main className="mx-auto min-h-[60vh] max-w-5xl px-5 py-12 sm:py-16">
      <div>
        <h1 className="text-3xl font-semibold tracking-tight text-zinc-950 dark:text-white">My favorites</h1>
        <p className="mt-3 text-sm text-zinc-500">Apps you have saved to evaluate or buy later.</p>
      </div>

      {error ? (
        <div className="mt-8 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-300">
          {error}
        </div>
      ) : null}

      {apps.length === 0 ? (
        <div className="mt-10 rounded-2xl border border-dashed border-zinc-300 px-6 py-20 text-center dark:border-white/15">
          <h2 className="text-lg font-semibold text-zinc-950 dark:text-white">No favorites yet</h2>
          <p className="mx-auto mt-2 max-w-md text-sm leading-6 text-zinc-500">
            Tap the heart on any marketplace listing to save it here for later.
          </p>
          <Link
            href="/search"
            className="mt-6 inline-flex rounded-full bg-zinc-950 px-5 py-2.5 text-sm font-semibold text-white dark:bg-white dark:text-zinc-950"
          >
            Explore products
          </Link>
        </div>
      ) : (
        <ul className="mt-10 grid gap-x-6 gap-y-10 md:grid-cols-2 lg:grid-cols-3">
          {apps.map((app) => {
            const cover = app.coverImageUrl || app.iconUrl || "/images/products/terminal-cover.svg"
            return (
              <li key={app.id} className="group">
                <Link
                  href={`/products/${app.slug}`}
                  className="relative block aspect-[16/10] overflow-hidden rounded-2xl border border-zinc-200 bg-zinc-100 shadow-sm transition duration-500 group-hover:-translate-y-1 group-hover:border-zinc-300 group-hover:shadow-xl group-hover:shadow-zinc-950/10 dark:border-white/10 dark:bg-zinc-900 dark:group-hover:border-white/20"
                >
                  <Image
                    src={cover}
                    alt={`${app.name} product preview`}
                    fill
                    sizes="(max-width: 768px) 100vw, (max-width: 1280px) 50vw, 33vw"
                    className="object-cover transition-transform duration-700 group-hover:scale-[1.025]"
                  />
                </Link>
                <div className="flex flex-col pt-4">
                  <div className="flex items-start justify-between gap-3">
                    <h2 className="text-[17px] font-semibold tracking-[-0.025em] text-zinc-950 dark:text-white">
                      <Link href={`/products/${app.slug}`} className="hover:underline">
                        {app.name}
                      </Link>
                    </h2>
                    <p className="shrink-0 text-sm font-semibold text-zinc-950 dark:text-white">
                      {app.priceCents === 0 ? "Free" : formatCurrency(app.priceCents / 100)}
                    </p>
                  </div>
                  <p className="mt-1 line-clamp-2 text-sm leading-6 text-zinc-600 dark:text-zinc-400">
                    {app.tagline || app.description}
                  </p>
                  <p className="mt-3 text-xs text-zinc-400">
                    by {app.developerName} · favorited {app.favoriteCount}
                  </p>
                </div>
              </li>
            )
          })}
        </ul>
      )}
    </main>
  )
}