import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { ArrowRightIcon } from "@/components/icons";
import ProductCard from "@/components/ProductCard";
import { JsonLd } from "@/lib/json-ld";
import { createMetadata } from "@/lib/metadata";
import { categories, categoryBySlug } from "@/lib/mock/categories";
import { products, productsByCategory } from "@/lib/mock/products";
import { absoluteUrl, siteConfig } from "@/lib/site";
import type { CategorySlug } from "@/types";

interface CategoryPageProps {
  params: Promise<{ slug: string }>;
}

export function generateStaticParams() {
  return categories.map((category) => ({ slug: category.slug }));
}

export async function generateMetadata({ params }: CategoryPageProps): Promise<Metadata> {
  const { slug } = await params;
  const category = categoryBySlug[slug as CategorySlug];

  if (!category) {
    return createMetadata({
      title: "Category not found",
      description: "This DevKit marketplace category could not be found.",
      path: `/category/${slug}`,
      keywords: ["software marketplace"],
    });
  }

  return createMetadata({
    title: `Best ${category.name} by Independent Developers`,
    description: category.longDescription,
    path: `/category/${category.slug}`,
    keywords: [
      category.name,
      `${category.shortName} marketplace`,
      "independent developer software",
      "buy software",
    ],
  });
}

export default async function CategoryPage({ params }: CategoryPageProps) {
  const { slug } = await params;
  const category = categoryBySlug[slug as CategorySlug];

  if (!category) notFound();

  const categoryProducts = productsByCategory[category.slug] ?? [];
  const showcasedProducts =
    categoryProducts.length >= 3
      ? categoryProducts
      : categoryProducts.concat(
          products.filter((product) => !categoryProducts.includes(product)).slice(0, 3),
        );

  const jsonLd = [
    {
      "@context": "https://schema.org",
      "@type": "CollectionPage",
      name: `${category.name} Marketplace`,
      description: category.longDescription,
      url: absoluteUrl(`/category/${category.slug}`),
      isPartOf: {
        "@type": "WebSite",
        name: siteConfig.name,
        url: siteConfig.url,
      },
    },
    {
      "@context": "https://schema.org",
      "@type": "ItemList",
      name: `Featured ${category.name}`,
      itemListElement: categoryProducts.map((product, index) => ({
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
          name: category.name,
          item: absoluteUrl(`/category/${category.slug}`),
        },
      ],
    },
  ];

  return (
    <>
      <JsonLd data={jsonLd} />
      <section className="relative overflow-hidden border-b border-zinc-200 dark:border-white/10">
        <div
          className={`absolute inset-0 bg-gradient-to-br ${category.accent} opacity-70`}
        />
        <div className="hero-grid absolute inset-0 opacity-40 dark:opacity-15" />
        <div className="relative mx-auto max-w-7xl px-5 py-16 sm:px-6 lg:px-8 lg:py-24">
          <nav className="mb-10 flex items-center gap-2 text-xs text-zinc-500">
            <Link href="/">Marketplace</Link>
            <span>/</span>
            <span className="text-zinc-800 dark:text-zinc-300">{category.name}</span>
          </nav>
          <div className="grid items-end gap-10 lg:grid-cols-[1fr_280px]">
            <div className="max-w-3xl">
              <span className="flex size-12 items-center justify-center rounded-2xl border border-zinc-950/10 bg-white/80 font-mono text-sm font-bold text-zinc-800 shadow-sm backdrop-blur dark:border-white/10 dark:bg-zinc-950/70 dark:text-white">
                {category.icon}
              </span>
              <p className="mt-8 text-xs font-semibold uppercase tracking-[0.18em] text-zinc-500">
                Curated category
              </p>
              <h1 className="mt-3 text-balance text-4xl font-semibold tracking-[-0.055em] text-zinc-950 sm:text-5xl dark:text-white">
                Exceptional {category.name}
              </h1>
              <p className="mt-5 max-w-2xl text-base leading-7 text-zinc-600 sm:text-lg dark:text-zinc-400">
                {category.longDescription}
              </p>
            </div>
            <div className="rounded-2xl border border-zinc-200/80 bg-white/70 p-6 backdrop-blur dark:border-white/10 dark:bg-zinc-950/60">
              <p className="text-3xl font-semibold tracking-[-0.04em] text-zinc-950 dark:text-white">
                {category.productCount}
              </p>
              <p className="mt-1 text-sm text-zinc-500">products in this category</p>
              <Link
                href={`/search?category=${category.slug}`}
                className="mt-5 inline-flex items-center gap-2 text-sm font-semibold text-zinc-800 dark:text-zinc-200"
              >
                Search category
                <ArrowRightIcon className="size-4" />
              </Link>
            </div>
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-7xl px-5 py-16 sm:px-6 lg:px-8 lg:py-24">
        <div className="flex items-end justify-between gap-6">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500">
              Featured
            </p>
            <h2 className="mt-3 text-3xl font-semibold tracking-[-0.045em] text-zinc-950 dark:text-white">
              Standout {category.shortName}
            </h2>
          </div>
          <Link
            href={`/search?category=${category.slug}`}
            className="hidden items-center gap-2 text-sm font-semibold text-zinc-700 sm:flex dark:text-zinc-300"
          >
            View all
            <ArrowRightIcon className="size-4" />
          </Link>
        </div>
        <div className="mt-10 grid gap-x-6 gap-y-12 md:grid-cols-2 lg:grid-cols-3">
          {showcasedProducts.slice(0, 3).map((product) => (
            <ProductCard key={product.id} product={product} />
          ))}
        </div>
      </section>

      <section className="border-y border-zinc-200 bg-zinc-50/70 dark:border-white/10 dark:bg-white/[0.02]">
        <div className="mx-auto max-w-7xl px-5 py-16 sm:px-6 lg:px-8 lg:py-24">
          <div className="grid gap-12 lg:grid-cols-[0.8fr_1.2fr]">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500">
                Buyer&apos;s guide
              </p>
              <h2 className="mt-3 text-3xl font-semibold tracking-[-0.045em] text-zinc-950 dark:text-white">
                Choosing {category.shortName} that lasts
              </h2>
            </div>
            <div className="grid gap-8 sm:grid-cols-2">
              {[
                {
                  title: "Understand the fit",
                  text: "Look for a specific workflow, a clear user, and honest boundaries—not a feature list that promises everything.",
                },
                {
                  title: "Evaluate the maker",
                  text: "Review the developer’s expertise, update history, documentation, and support commitments.",
                },
                {
                  title: "Check the economics",
                  text: "Compare the license, included updates, team limits, and true long-term cost for your use case.",
                },
                {
                  title: "Start with confidence",
                  text: "Use demos, documentation, and DevKit buyer protection to validate the product before committing.",
                },
              ].map((item, index) => (
                <article key={item.title}>
                  <span className="font-mono text-xs text-zinc-400">
                    {String(index + 1).padStart(2, "0")}
                  </span>
                  <h3 className="mt-3 text-sm font-semibold text-zinc-950 dark:text-white">
                    {item.title}
                  </h3>
                  <p className="mt-2 text-sm leading-6 text-zinc-600 dark:text-zinc-400">
                    {item.text}
                  </p>
                </article>
              ))}
            </div>
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-7xl px-5 py-16 sm:px-6 lg:px-8 lg:py-24">
        <div className="flex items-end justify-between">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500">
              Recently updated
            </p>
            <h2 className="mt-3 text-3xl font-semibold tracking-[-0.045em] text-zinc-950 dark:text-white">
              Fresh from the makers
            </h2>
          </div>
        </div>
        <div className="mt-10 grid gap-x-6 gap-y-12 md:grid-cols-2 lg:grid-cols-3">
          {showcasedProducts.slice(0, 3).map((product) => (
            <ProductCard key={`latest-${product.id}`} product={product} />
          ))}
        </div>
      </section>
    </>
  );
}
