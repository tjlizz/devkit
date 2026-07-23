import Link from "next/link";
import { siteConfig } from "@/lib/site";

export default function Footer() {
  return (
    <footer className="border-t border-surface-200 bg-surface-50">
      <div className="mx-auto max-w-7xl px-6 py-16">
        <div className="grid gap-12 sm:grid-cols-2 lg:grid-cols-4">
          <div>
            <Link href="/" className="flex items-center gap-2 text-lg font-bold tracking-tight text-surface-900">
              <span className="flex h-7 w-7 items-center justify-center rounded-md bg-brand-600 text-xs font-bold text-white">D</span>
              DevKit
            </Link>
            <p className="mt-3 text-sm leading-relaxed text-surface-500">
              A curated marketplace for exceptional software built by independent developers.
            </p>
          </div>
          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-surface-500">Browse</h3>
            <ul className="mt-4 space-y-3">
              <li><Link href="/" className="text-sm text-surface-600 transition-colors hover:text-surface-900">Featured</Link></li>
              <li><Link href="/category/saas" className="text-sm text-surface-600 transition-colors hover:text-surface-900">SaaS</Link></li>
              <li><Link href="/category/developer-tools" className="text-sm text-surface-600 transition-colors hover:text-surface-900">Dev Tools</Link></li>
              <li><Link href="/category/templates" className="text-sm text-surface-600 transition-colors hover:text-surface-900">Templates</Link></li>
            </ul>
          </div>
          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-surface-500">For Creators</h3>
            <ul className="mt-4 space-y-3">
              <li><Link href="/developers/mayachen" className="text-sm text-surface-600 transition-colors hover:text-surface-900">Become a Creator</Link></li>
              <li><Link href="/developers/mayachen" className="text-sm text-surface-600 transition-colors hover:text-surface-900">Developer Dashboard</Link></li>
            </ul>
          </div>
          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-surface-500">Connect</h3>
            <ul className="mt-4 space-y-3">
              <li><a href={siteConfig.links.x} target="_blank" rel="noopener noreferrer" className="text-sm text-surface-600 transition-colors hover:text-surface-900">X (Twitter)</a></li>
              <li><a href={siteConfig.links.github} target="_blank" rel="noopener noreferrer" className="text-sm text-surface-600 transition-colors hover:text-surface-900">GitHub</a></li>
            </ul>
          </div>
        </div>
        <div className="mt-12 border-t border-surface-200 pt-8 text-center text-xs text-surface-400">
          &copy; {new Date().getFullYear()} {siteConfig.name}. All rights reserved.
        </div>
      </div>
    </footer>
  );
}
