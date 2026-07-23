"use client";

import Link from "next/link";
import { useState } from "react";
import { useRouter } from "next/navigation";

export default function Navbar() {
  const [query, setQuery] = useState("");
  const router = useRouter();

  function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    if (query.trim()) router.push(`/search?q=${encodeURIComponent(query.trim())}`);
  }

  return (
    <header className="sticky top-0 z-50 border-b border-surface-200 bg-white/95 backdrop-blur supports-[backdrop-filter]:bg-white/80">
      <nav className="mx-auto flex h-16 max-w-7xl items-center justify-between gap-6 px-6">
        <Link href="/" className="flex items-center gap-2 text-lg font-bold tracking-tight text-surface-900">
          <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-brand-600 text-sm font-bold text-white">D</span>
          DevKit
        </Link>

        <div className="hidden items-center gap-6 md:flex">
          <Link href="/" className="text-sm font-medium text-surface-600 transition-colors hover:text-surface-900">Explore</Link>
          <Link href="/category/saas" className="text-sm font-medium text-surface-600 transition-colors hover:text-surface-900">Categories</Link>
          <Link href="/search" className="text-sm font-medium text-surface-600 transition-colors hover:text-surface-900">Search</Link>
        </div>

        <form onSubmit={handleSearch} className="hidden flex-1 max-w-xs sm:block">
          <input
            type="search"
            placeholder="Search products..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="w-full rounded-lg border border-surface-200 bg-surface-50 px-3 py-1.5 text-sm text-surface-900 placeholder-surface-400 outline-none transition-colors focus:border-brand-500 focus:bg-white"
          />
        </form>

        <div className="flex items-center gap-3">
          <Link
            href="/search"
            className="rounded-lg px-4 py-1.5 text-sm font-medium text-surface-600 transition-colors hover:bg-surface-100 hover:text-surface-900 sm:hidden"
          >
            Search
          </Link>
          <Link
            href="/developers/mayachen"
            className="rounded-lg bg-surface-900 px-4 py-1.5 text-sm font-medium text-white transition-all hover:bg-surface-800"
          >
            Become a Creator
          </Link>
        </div>
      </nav>
    </header>
  );
}
