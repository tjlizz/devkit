const statusConfig: Record<string, { label: string; className: string }> = {
  pending_review: {
    label: "Pending review",
    className:
      "bg-amber-50 text-amber-700 ring-amber-600/20 dark:bg-amber-950/50 dark:text-amber-300",
  },
  approved: {
    label: "Approved",
    className:
      "bg-emerald-50 text-emerald-700 ring-emerald-600/20 dark:bg-emerald-950/50 dark:text-emerald-300",
  },
  rejected: {
    label: "Rejected",
    className:
      "bg-red-50 text-red-700 ring-red-600/20 dark:bg-red-950/50 dark:text-red-300",
  },
  delisted: {
    label: "Delisted",
    className:
      "bg-zinc-100 text-zinc-600 ring-zinc-600/10 dark:bg-white/5 dark:text-zinc-400",
  },
}

export function AppStatusBadge({ status }: { status: string }) {
  const config = statusConfig[status] ?? {
    label: status,
    className: "bg-zinc-100 text-zinc-600 ring-zinc-600/10 dark:bg-white/5 dark:text-zinc-400",
  }
  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ring-1 ring-inset ${config.className}`}
    >
      {config.label}
    </span>
  )
}
