"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { useAuth } from "@/lib/auth-context"

export default function ProfileSettingsPage() {
  const router = useRouter()
  const { isAuthenticated, loading: authLoading, user, updateAvatar, resetAvatar } = useAuth()
  const [avatarFile, setAvatarFile] = useState<File | null>(null)
  const [previewUrl, setPreviewUrl] = useState("")
  const [loading, setLoading] = useState(false)
  const [resetting, setResetting] = useState(false)
  const [error, setError] = useState("")
  const [success, setSuccess] = useState("")

  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push("/login")
    }
  }, [authLoading, isAuthenticated, router])

  useEffect(() => {
    if (!avatarFile) {
      setPreviewUrl(user?.avatarUrl || "")
      return
    }
    const objectUrl = URL.createObjectURL(avatarFile)
    setPreviewUrl(objectUrl)
    return () => URL.revokeObjectURL(objectUrl)
  }, [avatarFile, user?.avatarUrl])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!avatarFile) {
      setError("Choose an image file first.")
      return
    }
    setLoading(true)
    setError("")
    setSuccess("")
    try {
      await updateAvatar(avatarFile)
      setAvatarFile(null)
      setSuccess("Avatar updated successfully.")
    } catch (err: any) {
      setError(err.message || "Avatar update failed")
    } finally {
      setLoading(false)
    }
  }

  async function handleReset() {
    setResetting(true)
    setError("")
    setSuccess("")
    try {
      await resetAvatar()
      setAvatarFile(null)
      setSuccess("Default avatar restored.")
    } catch (err: any) {
      setError(err.message || "Avatar reset failed")
    } finally {
      setResetting(false)
    }
  }

  if (authLoading || !isAuthenticated) {
    return (
      <div className="mx-auto flex min-h-[60vh] max-w-sm items-center px-4 py-16 text-sm text-zinc-500 dark:text-zinc-400">
        Loading...
      </div>
    )
  }

  return (
    <div className="mx-auto flex min-h-[60vh] max-w-lg items-center px-4 py-16">
      <div className="w-full">
        <h1 className="mb-2 text-2xl font-bold tracking-tight">Profile settings</h1>
        <p className="mb-8 text-sm text-zinc-500 dark:text-zinc-400">
          Upload the avatar people see next to your DevKit profile.
        </p>

        <form onSubmit={handleSubmit} className="space-y-5">
          {error && (
            <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-400">
              {error}
            </div>
          )}
          {success && (
            <div className="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700 dark:border-emerald-900 dark:bg-emerald-950 dark:text-emerald-300">
              {success}
            </div>
          )}

          <div className="flex items-center gap-4 rounded-lg border border-zinc-200 bg-zinc-50 p-4 dark:border-white/10 dark:bg-white/5">
            <img
              src={previewUrl}
              alt=""
              className="size-14 rounded-full border border-zinc-200 bg-white object-cover dark:border-white/10 dark:bg-zinc-900"
            />
            <div className="min-w-0">
              <p className="truncate text-sm font-medium">{user?.displayName}</p>
              <p className="truncate text-sm text-zinc-500 dark:text-zinc-400">{user?.email}</p>
            </div>
          </div>

          <div>
            <label htmlFor="avatarFile" className="mb-1.5 block text-sm font-medium">
              Avatar image
            </label>
            <input
              id="avatarFile"
              type="file"
              accept="image/png,image/jpeg,image/gif,image/webp"
              onChange={(e) => setAvatarFile(e.target.files?.[0] || null)}
              className="block w-full rounded-lg border border-zinc-200 bg-white px-4 py-2.5 text-sm outline-none focus:border-zinc-400 focus:ring-2 focus:ring-zinc-950/5 dark:border-white/10 dark:bg-zinc-900 dark:focus:border-white/20"
            />
            <p className="mt-2 text-xs text-zinc-500 dark:text-zinc-400">
              PNG, JPG, GIF, or WebP. Maximum size is 2MB.
            </p>
          </div>

          <div className="grid gap-3 sm:grid-cols-[1fr_auto]">
            <button
              type="submit"
              disabled={loading || !avatarFile}
              className="flex items-center justify-center rounded-lg bg-zinc-950 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-zinc-800 disabled:opacity-50 dark:bg-white dark:text-zinc-950 dark:hover:bg-zinc-200"
            >
              {loading ? "Uploading..." : "Upload avatar"}
            </button>
            <button
              type="button"
              disabled={resetting}
              onClick={handleReset}
              className="flex items-center justify-center rounded-lg border border-zinc-200 px-4 py-2.5 text-sm font-semibold text-zinc-700 transition hover:bg-zinc-50 disabled:opacity-50 dark:border-white/10 dark:text-zinc-300 dark:hover:bg-white/5"
            >
              {resetting ? "Resetting..." : "Use default"}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
