import { notFound } from "next/navigation";
import type { Metadata } from "next";
import Link from "next/link";
import Navbar from "@/components/Navbar";
import Footer from "@/components/Footer";
import ProductCard from "@/components/ProductCard";
import { developerByUsername, developers } from "@/lib/mock/developers";
import { products } from "@/lib/mock/products";
import { createMetadata } from "@/lib/metadata";
import { formatNumber, formatDate, formatCurrency } from "@/lib/format";
import { JsonLd } from "@/lib/json-ld";
import { siteConfig } from "@/lib/site";

interface Props { params: Promise<{ username: string }> }

export async function generateStaticParams() {
  return developers.map((d) => ({ username: d.username }));
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { username } = await params;
  const dev = developerByUsername[username];
  if (!dev) return {};
  return createMetadata({
    title: `${dev.name} — Developer Profile`,
    description: dev.bio,
    path: `/developers/${dev.username}`,
    keywords: [dev.name, ...dev.specialties, "developer"],
  });
}

export default async function DeveloperPage({ params }: Props) {
  const { username } = await params;
  const dev = developerByUsername[username];
  if (!dev) notFound();

  const devProducts = products.filter((p) => p.authorUsername === dev.username);

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "Person",
    name: dev.name,
    description: dev.bio,
    url: `${siteConfig.url}/developers/${dev.username}`,
  };

  return (
    <>
      <JsonLd data={jsonLd} />
      <Navbar />
      <main className="mx-auto max-w-7xl px-6 py-12 lg:py-16">
        <div className="lg:flex lg:gap-16">
          {/* Sidebar */}
          <aside className="lg:w-72 shrink-0">
            <div className="flex items-center gap-4 lg:block">
              <img src={dev.avatar} alt={dev.name} className="h-16 w-16 lg:h-24 lg:w-24 rounded-full bg-surface-200" />
              <div className="lg:mt-4">
                <div className="flex items-center gap-1.5">
                  <h1 className="text-xl font-bold text-surface-900">{dev.name}</h1>
                  {dev.verified && (
                    <svg className="h-4 w-4 text-brand-500" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clipRule="evenodd" />
                    </svg>
                  )}
                </div>
                <p className="mt-1 text-sm text-surface-500">{dev.location}</p>
                <p className="mt-3 text-sm leading-relaxed text-surface-600">{dev.bio}</p>

                <div className="mt-6 space-y-4 text-sm">
                  <div className="flex justify-between"><span className="text-surface-500">Followers</span><span className="font-medium text-surface-900">{formatNumber(dev.followers)}</span></div>
                  <div className="flex justify-between"><span className="text-surface-500">Products</span><span className="font-medium text-surface-900">{dev.publishedCount}</span></div>
                  <div className="flex justify-between"><span className="text-surface-500">Sales</span><span className="font-medium text-surface-900">{formatNumber(dev.totalSales)}</span></div>
                  <div className="flex justify-between"><span className="text-surface-500">Revenue</span><span className="font-medium text-surface-900">{formatCurrency(dev.revenue)}</span></div>
                  <div className="flex justify-between"><span className="text-surface-500">Joined</span><span className="font-medium text-surface-900">{formatDate(dev.joinedAt)}</span></div>
                </div>

                {dev.website && (
                  <a href={dev.website} target="_blank" rel="noopener noreferrer" className="mt-6 block rounded-lg border border-surface-300 px-4 py-2 text-center text-sm font-medium text-surface-700 transition-colors hover:bg-surface-50">
                    Visit Website &rarr;
                  </a>
                )}

                <div className="mt-4 flex gap-2">
                  {dev.socialLinks.map((link) => (
                    <a key={link.label} href={link.url} target="_blank" rel="noopener noreferrer" className="flex-1 rounded-lg border border-surface-300 px-3 py-2 text-center text-xs font-medium text-surface-600 transition-colors hover:bg-surface-50">
                      {link.label}
                    </a>
                  ))}
                </div>
              </div>
            </div>
          </aside>

          {/* Main */}
          <div className="mt-10 flex-1 lg:mt-0">
            <p className="text-base leading-relaxed text-surface-600">{dev.longBio}</p>

            <div className="mt-3 flex flex-wrap gap-2">
              {dev.specialties.map((s) => (
                <span key={s} className="rounded-lg bg-surface-100 px-3 py-1 text-xs font-medium text-surface-600">{s}</span>
              ))}
            </div>

            <h2 className="mt-12 text-xl font-bold text-surface-900">Products by {dev.name}</h2>
            <div className="mt-6 grid gap-6 sm:grid-cols-2">
              {devProducts.map((product) => (
                <ProductCard key={product.id} product={product} />
              ))}
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}
