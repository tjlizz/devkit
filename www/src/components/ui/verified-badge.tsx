import { CheckIcon } from "@/components/icons";

export function VerifiedBadge({ label = "Verified developer" }: { label?: string }) {
  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-blue-500/10 px-2 py-0.5 text-[11px] font-semibold text-blue-600 dark:text-blue-400">
      <CheckIcon className="size-3" />
      {label}
    </span>
  );
}
