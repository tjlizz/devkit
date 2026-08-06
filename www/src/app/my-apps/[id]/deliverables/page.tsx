"use client"

import { useCallback, useEffect, useRef, useState } from "react"
import Link from "next/link"
import { useParams } from "next/navigation"
import { useAuth, type AppArtifact, type DeveloperApp } from "@/lib/auth-context"
import { formatBytes, formatDate } from "@/lib/format"

function AuthGate({
  children,
}: {
  children: (user: NonNullable<ReturnType<typeof useAuth>["user"]>) => React.ReactNode
}) {
  const { isAuthenticated, user, loading: authLoading } = useAuth()

  if (authLoading) {
    return <main className="mx-auto min-h-[60vh] max-w-5xl px-5 py-16 text-sm text-zinc-500">Loading...</main>
  }

  if (!isAuthenticated || !user) {
    return (
      <main className="mx-auto min-h-[60vh] max-w-lg px-5 py-16 text-center">
        <h1 className="text-2xl font-semibold tracking-tight">Sign in required</h1>
        <p className="mt-3 text-sm text-zinc-500">Sign in to manage deliverables.</p>
        <Link className="mt-6 inline-flex rounded-lg bg-zinc-950 px-5 py-2.5 text-sm font-semibold text-white dark:bg-white dark:text-zinc-950" href="/login">
          Sign in
        </Link>
      </main>
    )
  }

  if (!user.isDeveloper) {
    return (
      <main className="mx-auto min-h-[60vh] max-w-lg px-5 py-16 text-center">
        <h1 className="text-2xl font-semibold tracking-tight">Developer approval required</h1>
        <p className="mt-3 text-sm text-zinc-500">Submit your developer application before managing deliverables.</p>
        <Link className="mt-6 inline-flex rounded-lg bg-zinc-950 px-5 py-2.5 text-sm font-semibold text-white dark:bg-white dark:text-zinc-950" href="/become-developer">
          Apply to publish
        </Link>
      </main>
    )
  }

  return <>{children(user)}</>
}

export default function DeliverablesPage() {
  return (
    <AuthGate>
      {() => <DeliverablesContent />}
    </AuthGate>
  )
}

function DeliverablesContent() {
  const { id } = useParams<{ id: string }>()
  const appId = Number(id)
  const { getDeveloperApp, listAppArtifacts, uploadAppArtifact, deleteAppArtifact } = useAuth()

  const [app, setApp] = useState<DeveloperApp | null>(null)
  const [artifacts, setArtifacts] = useState<AppArtifact[] | null>(null)
  const [error, setError] = useState("")
  const [notice, setNotice] = useState("")
  const [uploading, setUploading] = useState(false)
  const [deletingId, setDeletingId] = useState<number | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const load = useCallback(async () => {
    if (!appId) return
    setError("")
    try {
      const [appData, artifactList] = await Promise.all([getDeveloperApp(appId), listAppArtifacts(appId)])
      setApp(appData)
      setArtifacts(artifactList)
    } catch (err: any) {
      setError(err.message || "Could not load deliverables")
      setApp(null)
      setArtifacts([])
    }
  }, [appId, getDeveloperApp, listAppArtifacts])

  useEffect(() => {
    load()
  }, [load])

  async function handleUpload(file: File | undefined) {
    if (!file || !appId) return
    setUploading(true)
    setError("")
    setNotice("")
    try {
      await uploadAppArtifact(appId, file)
      setNotice(`Uploaded ${file.name}.`)
      await load()
    } catch (err: any) {
      setError(err.message || "Could not upload artifact")
    } finally {
      setUploading(false)
      if (fileInputRef.current) fileInputRef.current.value = ""
    }
  }

  async function handleDelete(artifact: AppArtifact) {
    if (!window.confirm(`Delete "${artifact.fileName}"? Buyers with entitlements will lose access to this file.`)) {
      return
    }
    setDeletingId(artifact.id)
    setError("")
    setNotice("")
    try {
      await deleteAppArtifact(appId, artifact.id)
      setNotice(`Deleted ${artifact.fileName}.`)
      await load()
    } catch (err: any) {
      setError(err.message || "Could not delete artifact")
    } finally {
      setDeletingId(null)
    }
  }

  if (app === null && artifacts === null && !error) {
    return (
      <main className="mx-auto min-h-[60vh] max-w-3xl px-5 py-16 text-sm text-zinc-500">
        Loading deliverables...
      </main>
    )
  }

  return (
    <main className="mx-auto min-h-[60vh] max-w-3xl px-5 py-12 sm:py-16">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <Link href="/my-apps" className="text-sm text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-200">
            ← My apps
          </Link>
          <h1 className="mt-2 text-3xl font-semibold tracking-tight text-zinc-950 dark:text-white">
            Deliverables{app ? `: ${app.name}` : ""}
          </h1>
          <p className="mt-3 text-sm text-zinc-500">
            Upload downloadable files that entitled buyers receive after purchase.
          </p>
        </div>
        <Link
          href={`/my-apps/${appId}/edit`}
          className="inline-flex rounded-lg border border-zinc-200 px-4 py-2.5 text-sm font-medium text-zinc-800 transition hover:bg-zinc-100 dark:border-white/10 dark:text-zinc-200 dark:hover:bg-white/5"
        >
          Edit listing
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
        </div>
      ) : null}

      <section className="mt-8 rounded-2xl border border-zinc-200 bg-white p-5 dark:border-white/10 dark:bg-zinc-950">
        <h2 className="text-sm font-semibold uppercase tracking-[0.14em] text-zinc-500">Upload a deliverable</h2>
        <p className="mt-1 text-xs text-zinc-500">Max 200 MB per file. Files are served only to verified buyers.</p>
        <div className="mt-4 flex flex-wrap items-center gap-3">
          <input
            ref={fileInputRef}
            type="file"
            disabled={uploading}
            onChange={(event) => handleUpload(event.target.files?.[0])}
            className="block w-full max-w-sm text-sm text-zinc-700 file:mr-3 file:rounded-lg file:border-0 file:bg-zinc-950 file:px-4 file:py-2 file:text-sm file:font-semibold file:text-white hover:file:bg-zinc-800 dark:text-zinc-300 dark:file:bg-white dark:file:text-zinc-950 dark:hover:file:bg-zinc-200"
          />
          {uploading ? <span className="text-sm text-zinc-500">Uploading...</span> : null}
        </div>
      </section>

      <section className="mt-8">
        <h2 className="text-sm font-semibold uppercase tracking-[0.14em] text-zinc-500">
          Uploaded files{artifacts ? ` (${artifacts.length})` : ""}
        </h2>
        {artifacts && artifacts.length === 0 ? (
          <p className="mt-3 text-sm text-zinc-500">No deliverables yet. Upload a file for buyers to download.</p>
        ) : null}
        <ul className="mt-4 grid gap-3">
          {(artifacts ?? []).map((artifact) => (
            <li
              key={artifact.id}
              className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-zinc-200 bg-white p-4 dark:border-white/10 dark:bg-zinc-950"
            >
              <div className="min-w-0">
                <p className="truncate font-semibold text-zinc-950 dark:text-white">{artifact.fileName}</p>
                <p className="mt-0.5 text-xs text-zinc-500">
                  {formatBytes(artifact.sizeBytes)} · Uploaded {formatDate(artifact.createdAt)}
                </p>
                {artifact.checksumSha256 ? (
                  <p className="mt-1 font-mono text-[10px] text-zinc-400">SHA-256 {artifact.checksumSha256.slice(0, 16)}…</p>
                ) : null}
              </div>
              <button
                type="button"
                onClick={() => handleDelete(artifact)}
                disabled={deletingId === artifact.id}
                className="rounded-lg border border-red-200 px-3 py-1.5 text-sm font-medium text-red-600 transition hover:bg-red-50 disabled:opacity-50 dark:border-red-900 dark:text-red-400 dark:hover:bg-red-950/50"
              >
                {deletingId === artifact.id ? "Deleting..." : "Delete"}
              </button>
            </li>
          ))}
        </ul>
      </section>
    </main>
  )
}