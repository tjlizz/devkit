import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { ArrowRightIcon } from "@/components/icons";
import ProductCard from "@/components/ProductCard";
import { JsonLd } from "@/lib/json-ld";
import { getMarketplaceProducts } from "@/lib/marketplace";
import { createMetadata } from "@/lib/metadata";
import { products } from "@/lib/mock/products";
import { absoluteUrl, siteConfig } from "@/lib/site";
import { productsForTag, tagsFromProducts } from "@/lib/tags";

interface TagPageProps {
  params: Promise<{ slug: string }>;
}

const knownTags = tagsFromProducts(products);

export function generateStaticParams() {
  return knownTags.map(({ slug }) => ({ slug }));
}

function displayNameForTag(slug: string, productTags = knownTags) {
  return productTags.find((tag) => tag.slug === slug)?.name;
}

export async function generateMetadata({ params }: TagPageProps): Promise<Metadata> {
  const { slug } = await params;
  const marketplaceProducts = await getMarketplaceProducts();
  const tagName =
    displayNameForTag(slug, tagsFromProducts(marketplaceProducts)) ?? displayNameForTag(slug);

  if (!tagName) {
    return createMetadata({
      title: "Tag not found",
      description: "This DevKit marketplace tag could not be found.",
      path: `/tag/${slug}`,
      keywords: ["developer software marketplace"],
    });
  }

  return createMetadata({
    title: `Best ${tagName} Software by Independent Developers`,
    description: `Discover trusted ${tagName} software, tools, and digital products built and supported by independent developers on DevKit.`,
    path: `/tag/${slug}`,
    keywords: [tagName, `${tagName} software`, `${tagName} tools`, "developer marketplace"],
  });
}

export default async function TagPage({ params }: TagPageProps) {
  const { slug } = await params;
  const marketplaceProducts = await getMarketplaceProducts();
  const marketplaceTags = tagsFromProducts(marketplaceProducts);
  const tagName = displayNameForTag(slug, marketplaceTags) ?? displayNameForTag(slug);

  if (!tagName) notFound();

  const taggedProducts = productsForTag(marketplaceProducts, slug);
  const fallbackProducts = productsForTag(products, slug);
  const displayedProducts = taggedProducts.length > 0 ? taggedProducts : fallbackProducts;
  const relatedTags = tagsFromProducts(displayedProducts)
    .filter((tag) => tag.slug !== slug)
    .slice(0, 8);
  const description = `Explore ${tagName} products made by independent developers. Compare focused tools, transparent pricing, and products with ongoing maker support.`;

  const jsonLd = [
    {
      "@context": "https://schema.org",
      "@type": "CollectionPage",
      name: `${tagName} Software Marketplace`,
      description,
      url: absoluteUrl(`/tag/${slug}`),
      isPartOf: {
        "@type": "WebSite",
        name: siteConfig.name,
        url: siteConfig.url,
      },
    },
    {
      "@context": "https://schema.org",
      "@type": "ItemList",
      name: `${tagName} products`,
      numberOfItems: displayedProducts.length,
      itemListElement: displayedProducts.map((product, index) => ({
        "@type": "ListItem",
        position: index + 1,
        name: product.name,
        url: absoluteUrl(`/products/${product.slug}`),
      })),
    },
    {
      "@context": "https://schema.org",
      "@type": "BreadcrumbList",
      itemListElement: [
        {
          "@type": "ListItem",
          position: 1,
          name: "Home",
          item: siteConfig.url,
        },
        {
          "@type": "ListItem",
          position: 2,
          name: tagName,
          item: absoluteUrl(`/tag/${slug}`),
        },
      ],
    },
  ];

  return (
    <>
      <JsonLd data={jsonLd} />
      <section className="relative overflow-hidden border-b border-zinc-200 dark:border-white/10">
        <div className="hero-grid absolute inset-0 opacity-50 dark:opacity-15" />
        <div className="relative mx-auto max-w-7xl px-5 py-16 sm:px-6 lg:px-8 lg:py-24">
          <nav className="mb-10 flex items-center gap-2 text-xs text-zinc-500">
            <Link href="/">Marketplace</Link>
            <span>/</span>
            <span>Tags</span>
            <span>/</span>
            <span className="text-zinc-800 dark:text-zinc-300">{tagName}</span>
          </nav>
          <div className="max-w-3xl">
            <span className="inline-flex rounded-full border border-zinc-200 bg-white px-3 py-1.5 font-mono text-xs font-semibold text-zinc-700 shadow-sm dark:border-white/10 dark:bg-zinc-950 dark:text-zinc-300">
              #{slug}
            </span>
            <p className="mt-8 text-xs font-semibold uppercase tracking-[0.18em] text-zinc-500">
              Curated collection
            </p>
            <h1 className="mt-3 text-balance text-4xl font-semibold tracking-[-0.055em] text-zinc-950 sm:text-5xl dark:text-white">
              The best {tagName} products, built by developers
            </h1>
            <p className="mt-5 max-w-2xl text-base leading-7 text-zinc-600 sm:text-lg dark:text-zinc-400">
              {description}
            </p>
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-7xl px-5 py-16 sm:px-6 lg:px-8 lg:py-24">
        <div className="flex flex-col justify-between gap-5 sm:flex-row sm:items-end">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500">
              {displayedProducts.length} {displayedProducts.length === 1 ? "product" : "products"}
            </p>
            <h2 className="mt-3 text-3xl font-semibold tracking-[-0.045em] text-zinc-950 dark:text-white">
              Explore {tagName}
            </h2>
          </div>
          <Link
            href={`/search?q=${encodeURIComponent(tagName)}`}
            className="inline-flex items-center gap-2 text-sm font-semibold text-zinc-700 dark:text-zinc-300"
          >
            Search the marketplace
            <ArrowRightIcon className="size-4" />
          </Link>
        </div>
        <div className="mt-10 grid gap-x-6 gap-y-12 md:grid-cols-2 lg:grid-cols-3">
          {displayedProducts.map((product) => (
            <ProductCard key={product.id} product={product} />
          ))}
        </div>

        {relatedTags.length > 0 && (
          <div className="mt-20 border-t border-zinc-200 pt-10 dark:border-white/10">
            <h2 className="text-sm font-semibold text-zinc-950 dark:text-white">Related tags</h2>
            <div className="mt-5 flex flex-wrap gap-2">
              {relatedTags.map((tag) => (
                <Link
                  key={tag.slug}
                  href={`/tag/${tag.slug}`}
                  className="rounded-full border border-zinc-200 px-3 py-2 font-mono text-xs text-zinc-600 transition hover:border-zinc-400 hover:text-zinc-950 dark:border-white/10 dark:text-zinc-400 dark:hover:border-white/25 dark:hover:text-white"
                >
                  #{tag.name}
                </Link>
              ))}
            </div>
          </div>
        )}
      </section>
    </>
  );
}
