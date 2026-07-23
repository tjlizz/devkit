import Link from "next/link";

export function Logo() {
  return (
    <Link
      href="/"
      className="group inline-flex items-center gap-2.5 font-semibold tracking-tight"
      aria-label="DevKit home"
    >
      <span className="relative flex size-7 items-center justify-center overflow-hidden rounded-[9px] bg-zinc-950 text-white shadow-sm ring-1 ring-black/10 dark:bg-white dark:text-zinc-950">
        <span className="text-[11px] font-black tracking-[-0.08em]">DK</span>
        <span className="absolute inset-x-1 bottom-0.5 h-px origin-left scale-x-0 bg-current transition-transform duration-300 group-hover:scale-x-100" />
      </span>
      <span className="text-[15px]">DevKit</span>
    </Link>
  );
}
