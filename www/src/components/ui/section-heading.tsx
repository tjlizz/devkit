import Link from "next/link";
import { ArrowRightIcon } from "@/components/icons";

interface SectionHeadingProps {
  eyebrow?: string;
  title: string;
  description?: string;
  link?: { label: string; href: string };
}

export function SectionHeading({
  eyebrow,
  title,
  description,
  link,
}: SectionHeadingProps) {
  return (
    <div className="mb-8 flex items-end justify-between gap-6 md:mb-10">
      <div className="max-w-2xl">
        {eyebrow ? (
          <p className="mb-3 text-xs font-semibold uppercase tracking-[0.18em] text-zinc-500">
            {eyebrow}
          </p>
        ) : null}
        <h2 className="text-2xl font-semibold tracking-[-0.035em] text-zinc-950 sm:text-3xl dark:text-white">
          {title}
        </h2>
        {description ? (
          <p className="mt-3 text-sm leading-6 text-zinc-600 sm:text-base dark:text-zinc-400">
            {description}
          </p>
        ) : null}
      </div>
      {link ? (
        <Link
          href={link.href}
          className="hidden shrink-0 items-center gap-2 text-sm font-medium text-zinc-700 transition-colors hover:text-zinc-950 sm:flex dark:text-zinc-300 dark:hover:text-white"
        >
          {link.label}
          <ArrowRightIcon className="size-4" />
        </Link>
      ) : null}
    </div>
  );
}
