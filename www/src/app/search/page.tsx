import type { Metadata } from "next";
import Link from "next/link";
import DeveloperCard from "@/components/DeveloperCard";
import { EmptyState } from "@/components/empty-state";
import { ArrowRightIcon } from "@/components/icons";
import ProductCard from "@/components/ProductCard";
import { SearchBar } from "@/components/search-bar";
import { JsonLd } from "@/lib/json-ld";
import { createMetadata } from "@/lib/metadata";
import { categoryBySlug } from "@/lib/mock/categories";
import { developers } from "@/lib/mock/developers";
import { products } from "@/lib/mock/products";
import { absoluteUrl, siteConfig } from "@/lib/site";
import type { CategorySlug } from "@/types";

interface SearchPageProps {
  searchParams: Promise<{
    q?: string;
    category?: string;
    type?: string;
  }>;
}

export async function generateMetadata({
  searchParams,
}: SearchPageProps): Promise<Metadata> {
  const { q = "" } = await searchParams;
  const query = q.trim();

  return createMetadata({
    title: query ? `Search results for “${query}”` : "Search Developer Software",
    description: query
      ? `Explore DevKit products, developers, and tags matching ${query}.`
      : "Search curated SaaS, AI apps, developer tools, templates, plugins, APIs, open source software, and independent makers.",
    path: "/search",
    keywords: [
      "search developer software",
      "software marketplace",
      "find developer tools",
      "independent makers",
      ...(query ? [query] : []),
    ],
  });
}

export default async function SearchPage({ searchParams }: SearchPageProps) {
  const {
    q = "",
    category = "",
    type = "all",
  } = await searchParams;
  const query = q.trim();
  const needle = query.toLowerCase();
  const validCategory = categoryBySlug[category as CategorySlug] ? category : "";

  const matchedProducts = products.filter((product) => {
    const inCategory = !validCategory || product.category === validCategory;
    const haystack = [
      product.name,
      product.tagline,
      product.description,
      product.category,
      ...product.tags,
    ]
      .join(" ")
      .toLowerCase();
    return inCategory && (!needle || haystack.includes(needle));
  });

  const matchedDevelopers = developers.filter((developer) => {
    const haystack = [
      developer.name,
      developer.username,
      developer.bio,
      ...developer.specialties,
    ]
      .join(" ")
      .toLowerCase();
    return !needle || haystack.includes(needle);
  });

  const matchedTags = Array.from(
    new Set(
      products
        .flatMap((product) => product.tags)
        .filter((tag) => !needle || tag.toLowerCase().includes(needle)),
    ),
  ).sort();

  const showProducts = type === "all" || type === "products";
  const showDevelopers = (type === "all" || type === "developers") && !validCategory;
  const showTags = (type === "all" || type === "tags") && !validCategory;
  const totalResults =
    (showProducts ? matchedProducts.length : 0) +
    (showDevelopers ? matchedDevelopers.length : 0) +
    (showTags ? matchedTags.length : 0);

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "SearchResultsPage",
    name: query ? `Search results for ${query}` : "Explore DevKit",
    description: "Search results from the DevKit developer software marketplace.",
    url: absoluteUrl(
      `/search${query ? `?q=${encodeURIComponent(query)}` : ""}`,
    ),
    isPartOf: {
      "@type": "WebSite",
      name: siteConfig.name,
      url: siteConfig.url,
    },
    mainEntity: {
      "@type": "ItemList",
      numberOfItems: matchedProducts.length,
      itemListElement: matchedProducts.map((product, index) => ({
        "@type": "ListItem",
        position: index + 1,
        name: product.name,
        url: absoluteUrl(`/products/${product.slug}`),
      })),
    },
  };

  return (
    <>
      <JsonLd data={jsonLd} />
      <section className="border-b border-zinc-200 bg-zinc-50/70 dark:border-white/10 dark:bg-white/[0.02]">
        <div className="mx-auto max-w-7xl px-5 py-14 sm:px-6 lg:px-8 lg:py-18">
          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-zinc-500">
            The marketplace
          </p>
          <h1 className="mt-4 text-4xl font-semibold tracking-[-0.055em] text-zinc-950 sm:text-5xl dark:text-white">
            {query ? `Results for “${query}”` : "Find your next essential tool"}
          </h1>
          <p className="mt-4 max-w-2xl text-base leading-7 text-zinc-600 dark:text-zinc-400">
            Search quality software, the developers behind it, and the technologies
            that connect them.
          </p>
          <div className="mt-9">
            <SearchBar query={query} category={validCategory} type={type} />
          </div>
        </div>
      </section>

      <div className="mx-auto max-w-7xl px-5 py-12 sm:px-6 lg:px-8 lg:py-16">
        <div className="mb-10 flex flex-wrap items-center justify-between gap-4 border-b border-zinc-200 pb-5 dark:border-white/10">
          <p className="text-sm text-zinc-500">
            <span className="font-semibold text-zinc-950 dark:text-white">
              {totalResults}
            </span>{" "}
            {totalResults === 1 ? "result" : "results"}
            {validCategory ? (
              <>
                {" "}
                in{" "}
                <span className="font-medium text-zinc-700 dark:text-zinc-300">
                  {categoryBySlug[validCategory as CategorySlug].name}
                </span>
              </>
            ) : null}
          </p>
          {(query || validCategory || type !== "all") && (
            <Link
              href="/search"
              className="text-xs font-semibold text-zinc-600 transition hover:text-zinc-950 dark:text-zinc-400 dark:hover:text-white"
            >
              Clear all filters
            </Link>
          )}
        </div>

        {totalResults === 0 ? (
          <EmptyState query={query} />
        ) : (
          <div className="space-y-20">
            {showProducts && matchedProducts.length > 0 ? (
              <section>
                <div className="flex items-end justify-between">
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500">
                      Products
                    </p>
                    <h2 className="mt-3 text-2xl font-semibold tracking-[-0.035em] text-zinc-950 dark:text-white">
                      Software for your shortlist
                    </h2>
                  </div>
                  <p className="text-xs text-zinc-500">{matchedProducts.length} matches</p>
                </div>
                <div className="mt-8 grid gap-x-6 gap-y-12 md:grid-cols-2 lg:grid-cols-3">
                  {matchedProducts.map((product) => (
                    <ProductCard key={product.id} product={product} />
                  ))}
                </div>
              </section>
            ) : null}

            {showDevelopers && matchedDevelopers.length > 0 ? (
              <section>
                <div className="flex items-end justify-between">
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500">
                      Developers
                    </p>
                    <h2 className="mt-3 text-2xl font-semibold tracking-[-0.035em] text-zinc-950 dark:text-white">
                      People making exceptional work
                    </h2>
                  </div>
                  <p className="text-xs text-zinc-500">
                    {matchedDevelopers.length} matches
                  </p>
                </div>
                <div className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                  {matchedDevelopers.map((developer) => (
                    <DeveloperCard key={developer.id} developer={developer} />
                  ))}
                </div>
              </section>
            ) : null}

            {showTags && matchedTags.length > 0 ? (
              <section className="rounded-2xl border border-zinc-200 p-7 sm:p-9 dark:border-white/10">
                <div className="flex flex-col justify-between gap-6 sm:flex-row sm:items-end">
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500">
                      Popular tags
                    </p>
                    <h2 className="mt-3 text-2xl font-semibold tracking-[-0.035em] text-zinc-950 dark:text-white">
                      Explore by technology and workflow
                    </h2>
                  </div>
                  <ArrowRightIcon className="hidden size-5 text-zinc-400 sm:block" />
                </div>
                <div className="mt-7 flex flex-wrap gap-2">
                  {matchedTags.map((tag) => (
                    <Link
                      key={tag}
                      href={`/search?q=${encodeURIComponent(tag)}&type=products`}
                      className="rounded-full border border-zinc-200 bg-zinc-50 px-4 py-2 text-sm font-medium text-zinc-700 transition hover:-translate-y-0.5 hover:border-zinc-300 hover:bg-white dark:border-white/10 dark:bg-white/5 dark:text-zinc-300 dark:hover:bg-white/10"
                    >
                      {tag}
                    </Link>
                  ))}
                </div>
              </section>
            ) : null}
          </div>
        )}
      </div>
    </>
  );
}
