import { notFound } from "next/navigation";
import type { Metadata } from "next";
import Link from "next/link";
import Navbar from "@/components/Navbar";
import Footer from "@/components/Footer";
import { productBySlug, products } from "@/lib/mock/products";
import { developerByUsername } from "@/lib/mock/developers";
import { createMetadata } from "@/lib/metadata";
import { formatCurrency, formatNumber, formatDate } from "@/lib/format";
import { JsonLd } from "@/lib/json-ld";
import { siteConfig } from "@/lib/site";

interface Props { params: Promise<{ slug: string }> }

export async function generateStaticParams() {
  return products.map((p) => ({ slug: p.slug }));
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  const product = productBySlug[slug];
  if (!product) return {};
  return createMetadata({
    title: product.name,
    description: product.tagline,
    path: `/products/${product.slug}`,
    keywords: [product.name, product.category, ...product.tags],
  });
}

export default async function ProductPage({ params }: Props) {
  const { slug } = await params;
  const product = productBySlug[slug];
  if (!product) notFound();

  const dev = developerByUsername[product.authorUsername];

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "Product",
    name: product.name,
    description: product.description,
    image: `${siteConfig.url}${product.coverImage}`,
    offers: {
      "@type": "Offer",
      price: product.price,
      priceCurrency: "USD",
    },
    author: dev ? { "@type": "Person", name: dev.name } : undefined,
  };

  return (
    <>
      <JsonLd data={jsonLd} />
      <Navbar />
      <main>
        {/* Product Hero */}
        <section className="border-b border-surface-200 bg-white">
          <div className="mx-auto max-w-7xl px-6 py-12 lg:py-16">
            <div className="lg:flex lg:gap-16">
              {/* Left: Info */}
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <span className="rounded-md bg-surface-100 px-2 py-0.5 text-xs font-medium uppercase tracking-wider text-surface-500">
                    {product.category.replace("-", " & ")}
                  </span>
                  {product.featured && (
                    <span className="rounded-md bg-brand-50 px-2 py-0.5 text-xs font-medium text-brand-600">Featured</span>
                  )}
                </div>
                <h1 className="mt-4 text-3xl font-bold tracking-tight text-surface-900 sm:text-4xl">{product.name}</h1>
                <p className="mt-2 text-lg leading-relaxed text-surface-500">{product.tagline}</p>

                {dev && (
                  <Link href={`/developers/${dev.username}`} className="mt-6 flex items-center gap-3 group">
                    <img src={dev.avatar} alt={dev.name} className="h-10 w-10 rounded-full bg-surface-200" />
                    <div>
                      <p className="text-sm font-medium text-surface-900 group-hover:text-brand-600 transition-colors">{dev.name}</p>
                      <p className="text-xs text-surface-500">{dev.specialties.slice(0, 2).join(" · ")}</p>
                    </div>
                  </Link>
                )}

                <div className="mt-8 flex flex-wrap gap-6 text-sm">
                  <div><span className="font-medium text-surface-900">{formatNumber(product.sales)}</span> <span className="text-surface-500">sales</span></div>
                  <div><span className="font-medium text-surface-900">{product.rating}</span> <span className="text-surface-500">rating ({product.reviewCount})</span></div>
                  <div><span className="font-medium text-surface-900">{formatNumber(product.favorites)}</span> <span className="text-surface-500">favorites</span></div>
                </div>
              </div>

              {/* Right: Price Card */}
              <div className="mt-8 lg:mt-0 lg:w-80 shrink-0">
                <div className="rounded-xl border border-surface-200 bg-surface-50 p-6">
                  <div className="flex items-baseline gap-1">
                    <span className="text-3xl font-bold text-surface-900">{formatCurrency(product.price)}</span>
                    {product.priceLabel && <span className="text-sm text-surface-500">/{product.priceLabel}</span>}
                  </div>
                  <button className="mt-4 w-full rounded-xl bg-surface-900 px-6 py-3 text-sm font-semibold text-white transition-all hover:bg-surface-800">
                    Buy Now
                  </button>
                  <p className="mt-3 text-xs text-center text-surface-400">Secure payment · Instant delivery</p>
                  <div className="mt-6 space-y-3 border-t border-surface-200 pt-4 text-xs text-surface-500">
                    <div className="flex justify-between"><span>License</span><span className="text-surface-700">{product.license}</span></div>
                    <div className="flex justify-between"><span>Delivery</span><span className="text-surface-700">{product.delivery}</span></div>
                    <div className="flex justify-between"><span>Support</span><span className="text-surface-700">{product.support}</span></div>
                    <div className="flex justify-between"><span>Updated</span><span className="text-surface-700">{formatDate(product.updatedAt)}</span></div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Screenshots */}
        <section className="py-12">
          <div className="mx-auto max-w-7xl px-6">
            <div className="grid gap-6">
              {product.screenshots.map((src, i) => (
                <img key={i} src={src} alt={`${product.name} screenshot ${i + 1}`} className="w-full rounded-xl border border-surface-200 bg-surface-100" />
              ))}
            </div>
          </div>
        </section>

        {/* Description */}
        <section className="border-t border-surface-200 py-16">
          <div className="mx-auto max-w-7xl px-6">
            <div className="max-w-3xl">
              <h2 className="text-xl font-bold text-surface-900">About {product.name}</h2>
              <p className="mt-4 text-base leading-relaxed text-surface-600">{product.longDescription}</p>
            </div>

            {/* Features */}
            <div className="mt-12 grid gap-8 sm:grid-cols-2 lg:grid-cols-3">
              {product.features.map((f, i) => (
                <div key={i}>
                  <h3 className="font-semibold text-surface-900">{f.title}</h3>
                  <p className="mt-1 text-sm leading-relaxed text-surface-500">{f.description}</p>
                </div>
              ))}
            </div>

            {/* Tech Stack */}
            <div className="mt-12">
              <h3 className="font-semibold text-surface-900">Tech Stack</h3>
              <div className="mt-3 flex flex-wrap gap-2">
                {product.techStack.map((t) => (
                  <span key={t} className="rounded-lg border border-surface-200 bg-white px-3 py-1 text-xs font-medium text-surface-600">{t}</span>
                ))}
              </div>
            </div>

            {/* Use Cases */}
            <div className="mt-10">
              <h3 className="font-semibold text-surface-900">Use Cases</h3>
              <ul className="mt-3 space-y-2">
                {product.useCases.map((u) => (
                  <li key={u} className="text-sm text-surface-600">&bull; {u}</li>
                ))}
              </ul>
            </div>

            {/* Links */}
            {(product.demoUrl || product.docsUrl) && (
              <div className="mt-10 flex gap-4">
                {product.demoUrl && (
                  <a href={product.demoUrl} target="_blank" rel="noopener noreferrer" className="rounded-lg border border-surface-300 px-5 py-2 text-sm font-medium text-surface-700 transition-colors hover:bg-surface-50">
                    Live Demo &rarr;
                  </a>
                )}
                {product.docsUrl && (
                  <a href={product.docsUrl} target="_blank" rel="noopener noreferrer" className="rounded-lg border border-surface-300 px-5 py-2 text-sm font-medium text-surface-700 transition-colors hover:bg-surface-50">
                    Documentation &rarr;
                  </a>
                )}
              </div>
            )}

            {/* Changelog */}
            {product.changelog.length > 0 && (
              <div className="mt-14">
                <h3 className="font-semibold text-surface-900">Changelog</h3>
                <div className="mt-4 space-y-4">
                  {product.changelog.map((entry) => (
                    <div key={entry.version} className="rounded-lg border border-surface-200 p-4">
                      <div className="flex items-center gap-2">
                        <span className="rounded bg-surface-100 px-2 py-0.5 text-xs font-mono text-surface-600">{entry.version}</span>
                        <span className="text-xs text-surface-400">{formatDate(entry.date)}</span>
                      </div>
                      <p className="mt-1 text-sm font-medium text-surface-900">{entry.title}</p>
                      <p className="mt-0.5 text-sm text-surface-500">{entry.notes}</p>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* FAQ */}
            {product.faq.length > 0 && (
              <div className="mt-14">
                <h3 className="font-semibold text-surface-900">FAQ</h3>
                <div className="mt-4 space-y-4">
                  {product.faq.map((item, i) => (
                    <div key={i}>
                      <p className="text-sm font-medium text-surface-900">Q: {item.question}</p>
                      <p className="mt-1 text-sm text-surface-600">A: {item.answer}</p>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </section>
      </main>
      <Footer />
    </>
  );
}
