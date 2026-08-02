"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { HeartIcon } from "@/components/icons";
import { formatNumber } from "@/lib/format";
import { useAuth } from "@/lib/auth-context";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";

interface FavoriteState {
  favorited: boolean;
  favoriteCount: number;
}

export function FavoriteButton({
  slug,
  initialCount,
  interactive,
}: {
  slug: string;
  initialCount: number;
  interactive: boolean;
}) {
  const router = useRouter();
  const { isAuthenticated, loading: authLoading, token } = useAuth();
  const [state, setState] = useState<FavoriteState>({ favorited: false, favoriteCount: initialCount });
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    setState((current) => ({ ...current, favoriteCount: initialCount }));
  }, [initialCount]);

  useEffect(() => {
    if (!interactive || authLoading || !isAuthenticated || !token) return;
    const controller = new AbortController();
    fetch(`${API_BASE}/api/v1/marketplace/apps/${encodeURIComponent(slug)}/favorite`, {
      headers: { Authorization: `Bearer ${token}` },
      signal: controller.signal,
    })
      .then(async (response) => {
        if (!response.ok) throw new Error("Could not load favorite");
        return response.json() as Promise<FavoriteState>;
      })
      .then(setState)
      .catch((requestError) => {
        if (requestError.name !== "AbortError") setError("Could not load favorite status");
      });
    return () => controller.abort();
  }, [authLoading, interactive, isAuthenticated, slug, token]);

  async function toggleFavorite() {
    if (!interactive) return;
    if (!isAuthenticated || !token) {
      router.push(`/login?next=${encodeURIComponent(`/products/${slug}`)}`);
      return;
    }
    setPending(true);
    setError("");
    try {
      const response = await fetch(`${API_BASE}/api/v1/marketplace/apps/${encodeURIComponent(slug)}/favorite`, {
        method: "POST",
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!response.ok) throw new Error("Could not update favorite");
      setState((await response.json()) as FavoriteState);
    } catch {
      setError("Could not update favorite");
    } finally {
      setPending(false);
    }
  }

  return (
    <button
      type="button"
      onClick={toggleFavorite}
      disabled={pending || !interactive}
      aria-pressed={interactive ? state.favorited : undefined}
      aria-label={state.favorited ? "Remove from favorites" : "Add to favorites"}
      title={error || (!interactive ? "Sign in favorites are available on published marketplace listings" : undefined)}
      className={`inline-flex h-10 items-center gap-2 rounded-full border px-4 text-sm font-medium transition disabled:cursor-default ${
        state.favorited
          ? "border-rose-200 bg-rose-50 text-rose-700 dark:border-rose-900 dark:bg-rose-950/40 dark:text-rose-300"
          : "border-zinc-200 text-zinc-700 hover:border-zinc-300 hover:bg-zinc-50 dark:border-white/10 dark:text-zinc-300 dark:hover:bg-white/5"
      }`}
    >
      <HeartIcon className="size-4" fill={state.favorited ? "currentColor" : "none"} />
      {formatNumber(state.favoriteCount)}
    </button>
  );
}
