"use client"

import { useEffect, useState } from "react"
import Link from "next/link"
import { useAuth } from "@/lib/auth-context"
import { PricingPlansEditor, type PlanDraft } from "@/components/pricing-plans-editor"

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

export default function PublishPage() {
  const { isAuthenticated, user, loading: authLoading, refreshUser, submitApp } = useAuth()
  const [submitting, setSubmitting] = useState(false)
  const [success, setSuccess] = useState(false)
  const [error, setError] = useState("")
  const [form, setForm] = useState({
    name: "",
    slug: "",
    tagline: "",
    description: "",
    category: "developer-tools",
    price: "0",
    iconUrl: "",
    coverImageUrl: "",
    demoUrl: "",
    docsUrl: "",
    sourceUrl: "",
    supportUrl: "",
    tags: "",
    version: "1.0.0",
    releaseNotes: "",
    plans: [] as PlanDraft[],
  })

  useEffect(() => {
    if (isAuthenticated) {
      refreshUser().catch(() => undefined)
    }
  }, [isAuthenticated, refreshUser])

  if (authLoading) {
    return <main className="mx-auto min-h-[60vh] max-w-2xl px-5 py-16 text-sm text-zinc-500">Loading...</main>
  }

  if (!isAuthenticated) {
    return (
      <main className="mx-auto min-h-[60vh] max-w-lg px-5 py-16 text-center">
        <h1 className="text-2xl font-semibold tracking-tight">Sign in required</h1>
        <p className="mt-3 text-sm text-zinc-500">You need an approved developer account to publish software.</p>
        <Link className="mt-6 inline-flex rounded-lg bg-zinc-950 px-5 py-2.5 text-sm font-semibold text-white dark:bg-white dark:text-zinc-950" href="/login">
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
        <Link className="mt-6 inline-flex rounded-lg bg-zinc-950 px-5 py-2.5 text-sm font-semibold text-white dark:bg-white dark:text-zinc-950" href="/become-developer">
          Apply to publish
        </Link>
      </main>
    )
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSubmitting(true)
    setError("")
    try {
      await submitApp({
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
        plans: form.plans.map((plan) => ({
          name: plan.name,
          priceCents: Math.round(Number(plan.price || "0") * 100),
          currency: "USD",
          description: plan.description,
          features: plan.features.split(",").map((feature) => feature.trim()).filter(Boolean),
        })),
      })
      setSuccess(true)
    } catch (err: any) {
      setError(err.message || "App submission failed")
    } finally {
      setSubmitting(false)
    }
  }

  if (success) {
    return (
      <main className="mx-auto min-h-[60vh] max-w-lg px-5 py-16 text-center">
        <h1 className="text-2xl font-semibold tracking-tight">Submitted for review</h1>
        <p className="mt-3 text-sm text-zinc-500">An administrator will review the app before it appears in the marketplace.</p>
        <Link className="mt-6 inline-flex rounded-lg bg-zinc-950 px-5 py-2.5 text-sm font-semibold text-white dark:bg-white dark:text-zinc-950" href="/">
          Back to marketplace
        </Link>
      </main>
    )
  }

  return (
    <main className="mx-auto max-w-3xl px-5 py-12 sm:py-16">
      <h1 className="text-3xl font-semibold tracking-tight text-zinc-950 dark:text-white">Publish an app</h1>
      <p className="mt-3 text-sm text-zinc-500">Submit software for administrator review before it goes live.</p>

      <form onSubmit={handleSubmit} className="mt-8 grid gap-5">
        {error ? <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-300">{error}</div> : null}
        <label className="text-sm font-medium">
          Name
          <input className="mt-2 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm dark:border-white/10 dark:bg-zinc-950" value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value, slug: form.slug || slugify(event.target.value) })} required maxLength={80} />
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
        <PricingPlansEditor
          plans={form.plans}
          onChange={(plans) => setForm({ ...form, plans })}
        />
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
        <button type="submit" disabled={submitting} className="rounded-lg bg-zinc-950 px-4 py-2.5 text-sm font-semibold text-white disabled:opacity-50 dark:bg-white dark:text-zinc-950">
          {submitting ? "Submitting..." : "Submit for review"}
        </button>
      </form>
    </main>
  )
}
