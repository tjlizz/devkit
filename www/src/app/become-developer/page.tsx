"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { useAuth } from "@/lib/auth-context"

export default function BecomeDeveloperPage() {
  const router = useRouter()
  const { isAuthenticated, user, upgradeToDeveloper, refreshUser, loading: authLoading } = useAuth()
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState(false)
  const [error, setError] = useState("")
  const [form, setForm] = useState({
    displayName: user?.displayName || "",
    profileUrl: "",
    reason: "",
  })

  useEffect(() => {
    if (isAuthenticated) {
      refreshUser().catch(() => undefined)
    }
  }, [isAuthenticated, refreshUser])

  useEffect(() => {
    setForm((current) => ({
      ...current,
      displayName: current.displayName || user?.displayName || "",
    }))
  }, [user?.displayName])

  if (authLoading) {
    return (
      <div className="mx-auto flex min-h-[60vh] max-w-lg items-center px-4 py-16">
        <p className="text-sm text-zinc-500">Loading...</p>
      </div>
    )
  }

  if (!isAuthenticated) {
    return (
      <div className="mx-auto flex min-h-[60vh] max-w-lg items-center px-4 py-16">
        <div className="w-full text-center">
          <h1 className="mb-2 text-2xl font-bold tracking-tight">Sign in required</h1>
          <p className="mb-6 text-sm text-zinc-500 dark:text-zinc-400">
            You need to sign in before becoming a developer.
          </p>
          <a
            href="/login"
            className="inline-flex items-center justify-center rounded-lg bg-zinc-950 px-6 py-2.5 text-sm font-semibold text-white hover:bg-zinc-800 dark:bg-white dark:text-zinc-950"
          >
            Sign in
          </a>
        </div>
      </div>
    )
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError("")
    try {
      await upgradeToDeveloper(form)
      setSuccess(true)
    } catch (err: any) {
      if (err.status === 409) {
        setError(err.message || "You already have a developer application.")
      } else if (err.status === 403) {
        setError("Your email is not verified. Please check your activation link.")
      } else {
        setError(err.message || "Upgrade failed")
      }
    } finally {
      setLoading(false)
    }
  }

  if (success) {
    return (
      <div className="mx-auto flex min-h-[60vh] max-w-lg items-center px-4 py-16">
        <div className="w-full text-center">
          <div className="mx-auto mb-4 flex size-12 items-center justify-center rounded-full bg-green-100 dark:bg-green-900">
            <svg className="size-6 text-green-600 dark:text-green-400" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <h1 className="mb-2 text-2xl font-bold tracking-tight">Application submitted</h1>
          <p className="mb-6 text-sm text-zinc-500 dark:text-zinc-400">
            An administrator will review your developer application before publishing access is enabled.
          </p>
          <a
            href="/"
            className="inline-flex items-center justify-center rounded-lg bg-zinc-950 px-6 py-2.5 text-sm font-semibold text-white hover:bg-zinc-800 dark:bg-white dark:text-zinc-950"
          >
            Go to marketplace
          </a>
        </div>
      </div>
    )
  }

  if (user?.isDeveloper || user?.developerApplicationStatus === "pending" || user?.developerApplicationStatus === "approved") {
    const status = user.isDeveloper || user.developerApplicationStatus === "approved" ? "approved" : "pending"
    const copy = {
      pending: "Your developer application is pending administrator review.",
      approved: "Your developer profile is approved. You can publish and manage applications.",
    }[status || "pending"]

    return (
      <div className="mx-auto flex min-h-[60vh] max-w-lg items-center px-4 py-16">
        <div className="w-full text-center">
          <p className="mb-3 text-xs font-semibold uppercase tracking-wide text-zinc-500">
            Developer application
          </p>
          <h1 className="mb-2 text-2xl font-bold tracking-tight capitalize">{status}</h1>
          <p className="mb-6 text-sm text-zinc-500 dark:text-zinc-400">{copy}</p>
          <a
            href="/"
            className="inline-flex items-center justify-center rounded-lg bg-zinc-950 px-6 py-2.5 text-sm font-semibold text-white hover:bg-zinc-800 dark:bg-white dark:text-zinc-950"
          >
            Go to marketplace
          </a>
        </div>
      </div>
    )
  }

  return (
    <div className="mx-auto flex min-h-[60vh] max-w-lg items-center px-4 py-16">
      <div className="w-full">
        <h1 className="mb-2 text-2xl font-bold tracking-tight">Become a developer</h1>
        <p className="mb-8 text-sm text-zinc-500 dark:text-zinc-400">
          Publish and sell your software on the DevKit marketplace
        </p>

        <form onSubmit={handleSubmit} className="space-y-5">
          {user?.developerApplicationStatus === "rejected" && (
            <div className="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-800 dark:bg-amber-950 dark:text-amber-300">
              Your last application was rejected. Submit an improved application for another review.
            </div>
          )}

          {error && (
            <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-400">
              {error}
            </div>
          )}

          <div className="rounded-lg border border-zinc-200 bg-zinc-50 p-5 text-sm dark:border-white/10 dark:bg-white/5">
            <h2 className="mb-2 font-semibold">What you get:</h2>
            <ul className="space-y-1.5 text-zinc-600 dark:text-zinc-400">
              <li>✓ Publish and sell your software products</li>
              <li>✓ Built-in payment and licensing system</li>
              <li>✓ Developer dashboard and analytics</li>
              <li>✓ Community of developers and buyers</li>
            </ul>
          </div>

          <div className="rounded-lg border border-zinc-200 bg-zinc-50 p-5 text-sm dark:border-white/10 dark:bg-white/5">
            <div className="flex items-center gap-3">
              <img
                src={user?.avatarUrl || `https://www.gravatar.com/avatar/00000000000000000000000000000000?d=mp&s=60`}
                alt=""
                className="size-10 rounded-full"
              />
              <div>
                <p className="font-medium">{user?.displayName}</p>
                <p className="text-zinc-500">{user?.email}</p>
              </div>
            </div>
          </div>

          <label className="block text-sm font-medium">
            Display name
            <input
              className="mt-2 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-white/10 dark:bg-zinc-950"
              value={form.displayName}
              onChange={(e) => setForm({ ...form, displayName: e.target.value })}
              required
              maxLength={80}
            />
          </label>

          <label className="block text-sm font-medium">
            Profile URL
            <input
              className="mt-2 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-white/10 dark:bg-zinc-950"
              type="url"
              value={form.profileUrl}
              onChange={(e) => setForm({ ...form, profileUrl: e.target.value })}
              placeholder="https://github.com/yourname"
            />
          </label>

          <label className="block text-sm font-medium">
            Why should DevKit approve you?
            <textarea
              className="mt-2 min-h-32 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-white/10 dark:bg-zinc-950"
              value={form.reason}
              onChange={(e) => setForm({ ...form, reason: e.target.value })}
              required
              minLength={20}
              maxLength={1200}
            />
          </label>

          <button
            type="submit"
            disabled={loading}
            className="flex w-full items-center justify-center rounded-lg bg-zinc-950 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-zinc-800 disabled:opacity-50 dark:bg-white dark:text-zinc-950 dark:hover:bg-zinc-200"
          >
            {loading ? "Processing..." : "Submit application"}
          </button>
        </form>
      </div>
    </div>
  )
}
