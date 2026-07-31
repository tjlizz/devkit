"use client"

import { useCallback, useEffect, useState } from "react"
import Link from "next/link"
import { useParams } from "next/navigation"
import { useAuth, type AppPublishPayload, type DeveloperApp } from "@/lib/auth-context"

const categories = [
  { label: "SaaS", value: "saas" },
  { label: "AI Applications", value: "ai-applications" },
  { label: "Developer Tools", value: "developer-tools" },
  { label: "Templates", value: "templates" },
  { label: "Plugins", value: "plugins" },
  { label: "APIs", value: "apis" },
  { label: "Open Source", value: "open-source" },
]

function slugify(value: string) {
  return value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 80)
}

function toForm(app: DeveloperApp) {
  return {
    name: app.name,
    slug: app.slug,
    tagline: app.tagline,
    description: app.description,
    category: app.category,
    price: (app.priceCents / 100).toString(),
    iconUrl: app.iconUrl,
    coverImageUrl: app.coverImageUrl,
    demoUrl: app.demoUrl,
    docsUrl: app.docsUrl,
    sourceUrl: app.sourceUrl,
    supportUrl: app.supportUrl,
    tags: app.tags.join(", "),
    version: app.version,
    releaseNotes: app.releaseNotes,
  }
}

function toPayload(form: ReturnType<typeof toForm>): AppPublishPayload {
  return {
    name: form.name,
    slug: form.slug || slugify(form.name),
    tagline: form.tagline,
    description: form.description,
    category: form.category,
    priceCents: Math.round(Number(form.price || "0") * 100),
    currency: "USD",
    iconUrl: form.iconUrl,
    coverImageUrl: form.coverImageUrl,
    demoUrl: form.demoUrl,
    docsUrl: form.docsUrl,
    sourceUrl: form.sourceUrl,
    supportUrl: form.supportUrl,
    tags: form.tags.split(",").map((tag) => tag.trim()).filter(Boolean),
    version: form.version,
    releaseNotes: form.releaseNotes,
  }
}

export default function EditAppPage() {
  const params = useParams<{ id: string }>()
  const appId = Number(params?.id)
  const { isAuthenticated, user, loading: authLoading, getDeveloperApp, updateDeveloperApp } = useAuth()

  const [app, setApp] = useState<DeveloperApp | null>(null)
  const [loadError, setLoadError] = useState("")
  const [form, setForm] = useState<ReturnType<typeof toForm> | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [saved, setSaved] = useState(false)
  const [error, setError] = useState("")

  const load = useCallback(async () => {
    if (!appId || Number.isNaN(appId)) {
      setLoadError("Invalid app")
      return
    }
    setLoadError("")
    try {
      const loaded = await getDeveloperApp(appId)
      setApp(loaded)
      setForm(toForm(loaded))
    } catch (err: any) {
      setLoadError(err.message || "Could not load app")
    }
  }, [appId, getDeveloperApp])

  useEffect(() => {
    if (isAuthenticated) {
      load()
    }
  }, [isAuthenticated, load])

  if (authLoading) {
    return <main className="mx-auto min-h-[60vh] max-w-3xl px-5 py-16 text-sm text-zinc-500">Loading...</main>
  }

  if (!isAuthenticated) {
    return (
      <main className="mx-auto min-h-[60vh] max-w-lg px-5 py-16 text-center">
        <h1 className="text-2xl font-semibold tracking-tight">Sign in required</h1>
        <p className="mt-3 text-sm text-zinc-500">Sign in to edit your app.</p>
        <Link className="mt-6 inline-flex rounded-lg bg-zinc-950 px-5 py-2.5 text-sm font-semibold text-white dark:bg-white dark:text-zinc-950" href="/login">
          Sign in
        </Link>
      </main>
    )
  }

  if (loadError) {
    return (
      <main className="mx-auto min-h-[60vh] max-w-lg px-5 py-16 text-center">
        <h1 className="text-2xl font-semibold tracking-tight">App not found</h1>
        <p className="mt-3 text-sm text-zinc-500">{loadError}</p>
        <Link className="mt-6 inline-flex rounded-lg bg-zinc-950 px-5 py-2.5 text-sm font-semibold text-white dark:bg-white dark:text-zinc-950" href="/my-apps">
          Back to my apps
        </Link>
      </main>
    )
  }

  if (!form || !app) {
    return <main className="mx-auto min-h-[60vh] max-w-3xl px-5 py-16 text-sm text-zinc-500">Loading app...</main>
  }

  if (app.status === "delisted") {
    return (
      <main className="mx-auto min-h-[60vh] max-w-lg px-5 py-16 text-center">
        <h1 className="text-2xl font-semibold tracking-tight">App is delisted</h1>
        <p className="mt-3 text-sm text-zinc-500">Delisted apps cannot be edited. Publish a new app if you want to bring it back.</p>
        <Link className="mt-6 inline-flex rounded-lg bg-zinc-950 px-5 py-2.5 text-sm font-semibold text-white dark:bg-white dark:text-zinc-950" href="/my-apps">
          Back to my apps
        </Link>
      </main>
    )
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!app || !form) return
    setSubmitting(true)
    setError("")
    try {
      await updateDeveloperApp(app.id, toPayload(form))
      setSaved(true)
    } catch (err: any) {
      setError(err.message || "Could not update app")
    } finally {
      setSubmitting(false)
    }
  }

  if (saved) {
    return (
      <main className="mx-auto min-h-[60vh] max-w-lg px-5 py-16 text-center">
        <h1 className="text-2xl font-semibold tracking-tight">Changes submitted</h1>
        <p className="mt-3 text-sm text-zinc-500">
          {app.status === "approved"
            ? "Your changes are pending review and the app is temporarily removed from the marketplace."
            : "Your changes have been saved."}
        </p>
        <Link className="mt-6 inline-flex rounded-lg bg-zinc-950 px-5 py-2.5 text-sm font-semibold text-white dark:bg-white dark:text-zinc-950" href="/my-apps">
          Back to my apps
        </Link>
      </main>
    )
  }

  return (
    <main className="mx-auto max-w-3xl px-5 py-12 sm:py-16">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight text-zinc-950 dark:text-white">Edit app</h1>
          <p className="mt-3 text-sm text-zinc-500">Update your listing details. Changes to an approved app require a new review.</p>
        </div>
        <Link href="/my-apps" className="text-sm font-medium text-zinc-500 transition hover:text-zinc-950 dark:hover:text-white">
          Back to my apps
        </Link>
      </div>

      {app.status === "approved" ? (
        <div className="mt-6 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-800 dark:bg-amber-950/50 dark:text-amber-300">
          This app is currently live. Saving changes will move it back to <strong>pending review</strong> and remove it from the marketplace until approved again.
        </div>
      ) : null}

      <form onSubmit={handleSubmit} className="mt-8 grid gap-5">
        {error ? <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-300">{error}</div> : null}
        <label className="text-sm font-medium">
          Name
          <input className="mt-2 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm dark:border-white/10 dark:bg-zinc-950" value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value, slug: form.slug === slugify(form.name) ? slugify(event.target.value) : form.slug })} required maxLength={80} />
        </label>
        <label className="text-sm font-medium">
          Slug
          <input className="mt-2 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm dark:border-white/10 dark:bg-zinc-950" value={form.slug} onChange={(event) => setForm({ ...form, slug: slugify(event.target.value) })} required />
        </label>
        <label className="text-sm font-medium">
          Tagline
          <input className="mt-2 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm dark:border-white/10 dark:bg-zinc-950" value={form.tagline} onChange={(event) => setForm({ ...form, tagline: event.target.value })} required maxLength={140} />
        </label>
        <label className="text-sm font-medium">
          Description
          <textarea className="mt-2 min-h-36 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm dark:border-white/10 dark:bg-zinc-950" value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} required />
        </label>
        <div className="grid gap-5 sm:grid-cols-2">
          <label className="text-sm font-medium">
            Category
            <select className="mt-2 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm dark:border-white/10 dark:bg-zinc-950" value={form.category} onChange={(event) => setForm({ ...form, category: event.target.value })}>
              {categories.map((category) => <option key={category.value} value={category.value}>{category.label}</option>)}
            </select>
          </label>
          <label className="text-sm font-medium">
            Price USD
            <input className="mt-2 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm dark:border-white/10 dark:bg-zinc-950" type="number" min="0" step="0.01" value={form.price} onChange={(event) => setForm({ ...form, price: event.target.value })} />
          </label>
        </div>
        <div className="grid gap-5 sm:grid-cols-2">
          {(["iconUrl", "coverImageUrl", "demoUrl", "docsUrl", "sourceUrl", "supportUrl"] as const).map((field) => (
            <label key={field} className="text-sm font-medium">
              {field}
              <input className="mt-2 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm dark:border-white/10 dark:bg-zinc-950" value={form[field]} onChange={(event) => setForm({ ...form, [field]: event.target.value })} />
            </label>
          ))}
        </div>
        <label className="text-sm font-medium">
          Tags
          <input className="mt-2 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm dark:border-white/10 dark:bg-zinc-950" value={form.tags} onChange={(event) => setForm({ ...form, tags: event.target.value })} placeholder="release, analytics, observability" />
        </label>
        <label className="text-sm font-medium">
          Version
          <input className="mt-2 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm dark:border-white/10 dark:bg-zinc-950" value={form.version} onChange={(event) => setForm({ ...form, version: event.target.value })} required />
        </label>
        <label className="text-sm font-medium">
          Release notes
          <textarea className="mt-2 min-h-24 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm dark:border-white/10 dark:bg-zinc-950" value={form.releaseNotes} onChange={(event) => setForm({ ...form, releaseNotes: event.target.value })} />
        </label>
        <div className="flex items-center gap-3">
          <button type="submit" disabled={submitting} className="rounded-lg bg-zinc-950 px-4 py-2.5 text-sm font-semibold text-white disabled:opacity-50 dark:bg-white dark:text-zinc-950">
            {submitting ? "Saving..." : "Save changes"}
          </button>
          <Link href="/my-apps" className="text-sm font-medium text-zinc-500 transition hover:text-zinc-950 dark:hover:text-white">
            Cancel
          </Link>
        </div>
      </form>
    </main>
  )
}
