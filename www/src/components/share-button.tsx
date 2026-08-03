"use client";

import { useEffect, useRef, useState } from "react";
import { ArrowUpRightIcon, CheckIcon } from "@/components/icons";

type ShareStatus = "idle" | "copied" | "error";

interface ShareButtonProps {
  title: string;
  text: string;
  iconOnly?: boolean;
}

async function copyToClipboard(value: string) {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(value);
    return;
  }

  const input = document.createElement("textarea");
  input.value = value;
  input.style.position = "fixed";
  input.style.opacity = "0";
  document.body.appendChild(input);
  input.select();
  const copied = document.execCommand("copy");
  input.remove();
  if (!copied) throw new Error("Copy failed");
}

export function ShareButton({ title, text, iconOnly = false }: ShareButtonProps) {
  const [status, setStatus] = useState<ShareStatus>("idle");
  const resetTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(
    () => () => {
      if (resetTimer.current) clearTimeout(resetTimer.current);
    },
    [],
  );

  function showTemporaryStatus(nextStatus: ShareStatus) {
    setStatus(nextStatus);
    if (resetTimer.current) clearTimeout(resetTimer.current);
    resetTimer.current = setTimeout(() => setStatus("idle"), 2200);
  }

  async function share() {
    const url = window.location.href;

    try {
      if (navigator.share) {
        await navigator.share({ title, text, url });
        return;
      }

      await copyToClipboard(url);
      showTemporaryStatus("copied");
    } catch (error) {
      if (error instanceof DOMException && error.name === "AbortError") return;
      try {
        await copyToClipboard(url);
        showTemporaryStatus("copied");
      } catch {
        showTemporaryStatus("error");
      }
    }
  }

  const label = status === "copied" ? "Link copied" : status === "error" ? "Copy failed" : "Share";

  return (
    <button
      type="button"
      onClick={share}
      aria-label={iconOnly ? `${label} ${title}` : undefined}
      title={iconOnly ? label : undefined}
      className={
        iconOnly
          ? "flex size-11 items-center justify-center rounded-full border border-zinc-200 text-zinc-600 transition hover:bg-zinc-50 dark:border-white/10 dark:text-zinc-300 dark:hover:bg-white/5"
          : "inline-flex h-10 items-center gap-2 rounded-full border border-zinc-200 px-4 text-sm font-medium text-zinc-700 transition hover:border-zinc-300 hover:bg-zinc-50 dark:border-white/10 dark:text-zinc-300 dark:hover:bg-white/5"
      }
    >
      {status === "copied" ? <CheckIcon className="size-4" /> : <ArrowUpRightIcon className="size-4" />}
      {!iconOnly && <span aria-live="polite">{label}</span>}
      {iconOnly && <span className="sr-only" aria-live="polite">{label}</span>}
    </button>
  );
}
