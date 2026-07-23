"use client"

import Link from "next/link";
import { MenuIcon, SearchIcon } from "@/components/icons";
import { ThemeToggle } from "@/components/theme-toggle";
import { Logo } from "@/components/ui/logo";
import { useAuth } from "@/lib/auth-context";

const navLinks = [
  { href: "/search", label: "Explore" },
  { href: "/category/saas", label: "Products" },
  { href: "/developers/mayachen", label: "Developers" },
];

export default function Navbar() {
  const { isAuthenticated, user, logout } = useAuth();

  return (
    <header className="sticky top-0 z-50 border-b border-zinc-200/70 bg-white/85 backdrop-blur-xl dark:border-white/10 dark:bg-zinc-950/85">
      <div className="mx-auto flex h-16 max-w-7xl items-center gap-6 px-5 sm:px-6 lg:px-8">
        <Logo />
        <nav className="hidden items-center gap-1 lg:flex" aria-label="Main navigation">
          {navLinks.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className="rounded-lg px-3 py-2 text-sm font-medium text-zinc-600 transition-colors hover:bg-zinc-100 hover:text-zinc-950 dark:text-zinc-400 dark:hover:bg-white/5 dark:hover:text-white"
            >
              {item.label}
            </Link>
          ))}
        </nav>
        <form
          action="/search"
          role="search"
          className="ml-auto hidden max-w-xs flex-1 items-center md:flex"
        >
          <SearchIcon className="pointer-events-none z-10 mr-[-28px] ml-3 size-4 text-zinc-400" />
          <input
            name="q"
            aria-label="Search products"
            placeholder="Search products..."
            className="h-9 w-full rounded-full border border-zinc-200 bg-zinc-50 pr-4 pl-9 text-sm outline-none transition placeholder:text-zinc-400 focus:border-zinc-400 focus:bg-white focus:ring-2 focus:ring-zinc-950/5 dark:border-white/10 dark:bg-white/5 dark:focus:border-white/20 dark:focus:bg-white/[0.07]"
          />
        </form>
        <ThemeToggle />
        <div className="flex items-center gap-2">
          {isAuthenticated && user ? (
            <details className="relative group">
              <summary className="flex cursor-pointer list-none items-center gap-2 rounded-full p-1.5 transition hover:bg-zinc-100 dark:hover:bg-white/5">
                <img
                  src={user.avatarUrl}
                  alt=""
                  className="size-7 rounded-full"
                />
                <span className="hidden text-sm font-medium sm:inline">
                  {user.displayName}
                </span>
              </summary>
              <nav className="absolute right-0 top-full mt-2 w-52 rounded-2xl border border-zinc-200 bg-white p-2 shadow-xl dark:border-white/10 dark:bg-zinc-900">
                <Link
                  href="/become-developer"
                  className="block rounded-xl px-3 py-2 text-sm font-medium hover:bg-zinc-100 dark:hover:bg-white/5"
                >
                  Become a developer
                </Link>
                <hr className="my-1 border-zinc-100 dark:border-white/10" />
                <button
                  onClick={logout}
                  className="block w-full rounded-xl px-3 py-2 text-left text-sm font-medium text-zinc-600 hover:bg-zinc-100 dark:text-zinc-400 dark:hover:bg-white/5"
                >
                  Sign out
                </button>
              </nav>
            </details>
          ) : (
            <>
              <Link
                href="/login"
                className="hidden h-9 items-center rounded-full px-4 text-sm font-medium text-zinc-600 transition hover:bg-zinc-100 dark:text-zinc-400 dark:hover:bg-white/5 sm:inline-flex"
              >
                Sign in
              </Link>
              <Link
                href="/register"
                className="hidden h-9 items-center rounded-full bg-zinc-950 px-4 text-sm font-semibold text-white transition hover:-translate-y-0.5 hover:bg-zinc-800 hover:shadow-lg dark:bg-white dark:text-zinc-950 dark:hover:bg-zinc-200 sm:inline-flex"
              >
                Get started
              </Link>
            </>
          )}
        </div>
        <details className="relative lg:hidden">
          <summary className="flex size-9 cursor-pointer list-none items-center justify-center rounded-full border border-zinc-200 dark:border-white/10">
            <MenuIcon className="size-4" />
            <span className="sr-only">Open menu</span>
          </summary>
          <nav className="absolute top-12 right-0 w-48 rounded-2xl border border-zinc-200 bg-white p-2 shadow-xl dark:border-white/10 dark:bg-zinc-900">
            {navLinks.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className="block rounded-xl px-3 py-2 text-sm font-medium hover:bg-zinc-100 dark:hover:bg-white/5"
              >
                {item.label}
              </Link>
            ))}
            <hr className="my-1 border-zinc-100 dark:border-white/10" />
            {isAuthenticated ? (
              <>
                <div className="px-3 py-2 text-sm font-medium text-zinc-500">{user?.displayName}</div>
                <Link
                  href="/become-developer"
                  className="block rounded-xl px-3 py-2 text-sm font-medium hover:bg-zinc-100 dark:hover:bg-white/5"
                >
                  Become a developer
                </Link>
                <button
                  onClick={logout}
                  className="block w-full rounded-xl px-3 py-2 text-left text-sm font-medium text-zinc-600 hover:bg-zinc-100 dark:text-zinc-400 dark:hover:bg-white/5"
                >
                  Sign out
                </button>
              </>
            ) : (
              <>
                <Link href="/login" className="block rounded-xl px-3 py-2 text-sm font-medium hover:bg-zinc-100 dark:hover:bg-white/5">Sign in</Link>
                <Link href="/register" className="block rounded-xl px-3 py-2 text-sm font-medium hover:bg-zinc-100 dark:hover:bg-white/5">Get started</Link>
              </>
            )}
          </nav>
        </details>
      </div>
    </header>
  );
}
