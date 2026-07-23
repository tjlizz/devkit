import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import { notFound } from "next/navigation";
import {
  ArrowRightIcon,
  ArrowUpRightIcon,
  CheckIcon,
  ExternalLinkIcon,
  HeartIcon,
  StarIcon,
} from "@/components/icons";
import { PricingCard } from "@/components/pricing-card";
import ProductCard from "@/components/ProductCard";
import { ProductScreenshots } from "@/components/product-screenshots";
import { VerifiedBadge } from "@/components/ui/verified-badge";
import { formatDate, formatNumber } from "@/lib/format";
import { JsonLd } from "@/lib/json-ld";
import { createMetadata } from "@/lib/metadata";
import { categoryBySlug } from "@/lib/mock/categories";
import { developerByUsername } from "@/lib/mock/developers";
import { productBySlug, products } from "@/lib/mock/products";
import { absoluteUrl, siteConfig } from "@/lib/site";

interface ProductPageProps {
  params: Promise<{ slug: string }>;
}

export function generateStaticParams() {
  return products.map((product) => ({ slug: product.slug }));
}

export async function generateMetadata({ params }: ProductPageProps): Promise<Metadata> {
  const { slug } = await params;
  const product = productBySlug[slug];

  if (!product) {
    return createMetadata({
      title: "Product not found",
      description: "This DevKit product could not be found.",
      path: `/products/${slug}`,
      keywords: ["developer products"],
    });
  }

  const category = categoryBySlug[product.category];

  return createMetadata({
    title: `${product.name} — ${product.tagline}`,
    description: product.description,
    path: `/products/${product.slug}`,
    keywords: [
      product.name,
      category.name,
      ...product.tags,
      "developer-created software",
    ],
    image: product.coverImage,
  });
}

export default async function ProductPage({ params }: ProductPageProps) {
  const { slug } = await params;
  const product = productBySlug[slug];

  if (!product) notFound();

  const author = developerByUsername[product.authorUsername];
  const category = categoryBySlug[product.category];
  const related = products
    .filter((item) => item.slug !== product.slug && item.category === product.category)
    .concat(products.filter((item) => item.slug !== product.slug))
    .slice(0, 3);

  const jsonLd = [
    {
      "@context": "https://schema.org",
      "@type": "SoftwareApplication",
      name: product.name,
      applicationCategory: category.name,
      description: product.description,
      image: product.screenshots.map(absoluteUrl),
      url: absoluteUrl(`/products/${product.slug}`),
      author: {
        "@type": "Person",
        name: author.name,
        url: absoluteUrl(`/developers/${author.username}`),
      },
      offers: {
        "@type": "Offer",
        price: product.price,
        priceCurrency: "USD",
        availability: "https://schema.org/InStock",
        url: absoluteUrl(`/products/${product.slug}`),
      },
      aggregateRating: {
        "@type": "AggregateRating",
        ratingValue: product.rating,
        ratingCount: product.reviewCount,
        bestRating: 5,
      },
      datePublished: product.createdAt,
      dateModified: product.updatedAt,
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
        {
          "@type": "ListItem",
          position: 3,
          name: product.name,
          item: absoluteUrl(`/products/${product.slug}`),
        },
      ],
    },
    {
      "@context": "https://schema.org",
      "@type": "FAQPage",
      mainEntity: product.faq.map((item) => ({
        "@type": "Question",
        name: item.question,
        acceptedAnswer: {
          "@type": "Answer",
          text: item.answer,
        },
      })),
    },
  ];

  return (
    <>
      <JsonLd data={jsonLd} />
      <section className="border-b border-zinc-200 dark:border-white/10">
        <div className="mx-auto max-w-7xl px-5 py-12 sm:px-6 lg:px-8 lg:py-18">
          <nav className="mb-10 flex items-center gap-2 text-xs text-zinc-500">
            <Link href="/">Marketplace</Link>
            <span>/</span>
            <Link href={`/category/${category.slug}`}>{category.shortName}</Link>
            <span>/</span>
            <span className="text-zinc-800 dark:text-zinc-300">{product.name}</span>
          </nav>
          <div className="grid items-end gap-10 lg:grid-cols-[1fr_auto]">
            <div className="flex flex-col gap-7 sm:flex-row sm:items-start">
              <div className="relative size-20 shrink-0 overflow-hidden rounded-2xl border border-zinc-200 bg-zinc-100 shadow-sm sm:size-24 dark:border-white/10 dark:bg-zinc-900">
                <Image
                  src={product.coverImage}
                  alt={`${product.name} logo`}
                  fill
                  sizes="96px"
                  className="object-cover"
                />
              </div>
              <div>
                <div className="flex flex-wrap items-center gap-3">
                  <Link
                    href={`/category/${category.slug}`}
                    className="text-xs font-semibold uppercase tracking-[0.15em] text-zinc-500 transition hover:text-zinc-950 dark:hover:text-white"
                  >
                    {category.name}
                  </Link>
                  <span className="size-1 rounded-full bg-zinc-300 dark:bg-zinc-700" />
                  <span className="flex items-center gap-1 text-xs text-zinc-600 dark:text-zinc-400">
                    <StarIcon className="size-3.5 fill-current text-amber-500" />
                    {product.rating} ({product.reviewCount} reviews)
                  </span>
                </div>
                <h1 className="mt-3 text-4xl font-semibold tracking-[-0.055em] text-zinc-950 sm:text-5xl dark:text-white">
                  {product.name}
                </h1>
                <p className="mt-4 max-w-3xl text-lg leading-7 text-zinc-600 sm:text-xl dark:text-zinc-400">
                  {product.tagline}
                </p>
                <div className="mt-6 flex flex-wrap items-center gap-4">
                  <Link
                    href={`/developers/${author.username}`}
                    className="flex items-center gap-2.5"
                  >
                    <Image
                      src={author.avatar}
                      alt=""
                      width={30}
                      height={30}
                      className="rounded-full ring-1 ring-zinc-950/10 dark:ring-white/15"
                    />
                    <span className="text-sm font-medium text-zinc-700 dark:text-zinc-300">
                      by {author.name}
                    </span>
                  </Link>
                  <VerifiedBadge />
                  <span className="text-xs text-zinc-500">
                    Updated {formatDate(product.updatedAt)}
                  </span>
                </div>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <button className="inline-flex h-10 items-center gap-2 rounded-full border border-zinc-200 px-4 text-sm font-medium text-zinc-700 transition hover:border-zinc-300 hover:bg-zinc-50 dark:border-white/10 dark:text-zinc-300 dark:hover:bg-white/5">
                <HeartIcon className="size-4" />
                {formatNumber(product.favorites)}
              </button>
              <button className="inline-flex h-10 items-center gap-2 rounded-full border border-zinc-200 px-4 text-sm font-medium text-zinc-700 transition hover:border-zinc-300 hover:bg-zinc-50 dark:border-white/10 dark:text-zinc-300 dark:hover:bg-white/5">
                Share
                <ArrowUpRightIcon className="size-4" />
              </button>
            </div>
          </div>
        </div>
      </section>

      <div className="mx-auto grid max-w-7xl gap-12 px-5 py-12 sm:px-6 lg:grid-cols-[minmax(0,1fr)_340px] lg:px-8 lg:py-16">
        <div className="min-w-0 space-y-20">
          <ProductScreenshots images={product.screenshots} productName={product.name} />

          <section>
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500">
              Why {product.name}
            </p>
            <h2 className="mt-3 max-w-2xl text-3xl font-semibold tracking-[-0.045em] text-zinc-950 dark:text-white">
              Purpose-built software, without the usual compromise.
            </h2>
            <p className="mt-6 max-w-3xl text-base leading-8 text-zinc-600 dark:text-zinc-400">
              {product.longDescription}
            </p>
            <div className="mt-10 grid gap-px overflow-hidden rounded-2xl border border-zinc-200 bg-zinc-200 sm:grid-cols-2 dark:border-white/10 dark:bg-white/10">
              {product.features.map((feature, index) => (
                <div key={feature.title} className="bg-white p-6 sm:p-8 dark:bg-zinc-950">
                  <span className="flex size-7 items-center justify-center rounded-full bg-zinc-100 text-xs font-semibold text-zinc-600 dark:bg-white/5 dark:text-zinc-300">
                    {String(index + 1).padStart(2, "0")}
                  </span>
                  <h3 className="mt-6 text-base font-semibold text-zinc-950 dark:text-white">
                    {feature.title}
                  </h3>
                  <p className="mt-2 text-sm leading-6 text-zinc-600 dark:text-zinc-400">
                    {feature.description}
                  </p>
                </div>
              ))}
            </div>
          </section>

          <section className="grid gap-8 rounded-2xl bg-zinc-50 p-7 sm:grid-cols-2 sm:p-10 dark:bg-white/[0.035]">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.15em] text-zinc-500">
                Built with
              </p>
              <h2 className="mt-3 text-2xl font-semibold tracking-[-0.035em] text-zinc-950 dark:text-white">
                Modern, durable technology
              </h2>
              <div className="mt-6 flex flex-wrap gap-2">
                {product.techStack.map((tech) => (
                  <span
                    key={tech}
                    className="rounded-lg border border-zinc-200 bg-white px-3 py-2 font-mono text-xs text-zinc-700 dark:border-white/10 dark:bg-zinc-900 dark:text-zinc-300"
                  >
                    {tech}
                  </span>
                ))}
              </div>
            </div>
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.15em] text-zinc-500">
                Great for
              </p>
              <ul className="mt-5 space-y-3">
                {product.useCases.map((useCase) => (
                  <li
                    key={useCase}
                    className="flex items-center gap-2.5 text-sm text-zinc-700 dark:text-zinc-300"
                  >
                    <CheckIcon className="size-4 text-emerald-600 dark:text-emerald-400" />
                    {useCase}
                  </li>
                ))}
              </ul>
            </div>
          </section>

          <section id="documentation">
            <div className="flex flex-col justify-between gap-6 sm:flex-row sm:items-end">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500">
                  What&apos;s new
                </p>
                <h2 className="mt-3 text-3xl font-semibold tracking-[-0.045em] text-zinc-950 dark:text-white">
                  Changelog
                </h2>
              </div>
              <Link
                href={product.docsUrl}
                className="inline-flex items-center gap-2 text-sm font-semibold text-zinc-700 dark:text-zinc-300"
              >
                Read documentation
                <ExternalLinkIcon className="size-4" />
              </Link>
            </div>
            <div className="mt-8 divide-y divide-zinc-200 border-y border-zinc-200 dark:divide-white/10 dark:border-white/10">
              {product.changelog.map((entry) => (
                <article
                  key={entry.version}
                  className="grid gap-3 py-6 sm:grid-cols-[100px_1fr]"
                >
                  <div>
                    <span className="font-mono text-xs font-semibold text-zinc-800 dark:text-zinc-200">
                      v{entry.version}
                    </span>
                    <p className="mt-1 text-xs text-zinc-500">{formatDate(entry.date)}</p>
                  </div>
                  <div>
                    <h3 className="text-sm font-semibold text-zinc-950 dark:text-white">
                      {entry.title}
                    </h3>
                    <p className="mt-1 text-sm leading-6 text-zinc-600 dark:text-zinc-400">
                      {entry.notes}
                    </p>
                  </div>
                </article>
              ))}
            </div>
          </section>

          <section>
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500">
              Common questions
            </p>
            <h2 className="mt-3 text-3xl font-semibold tracking-[-0.045em] text-zinc-950 dark:text-white">
              Before you buy
            </h2>
            <div className="mt-8 divide-y divide-zinc-200 border-y border-zinc-200 dark:divide-white/10 dark:border-white/10">
              {product.faq.map((item) => (
                <details key={item.question} className="group py-5">
                  <summary className="flex cursor-pointer items-center justify-between gap-6 text-sm font-semibold text-zinc-950 dark:text-white">
                    {item.question}
                    <span className="text-lg font-normal text-zinc-400 transition-transform group-open:rotate-45">
                      +
                    </span>
                  </summary>
                  <p className="max-w-2xl pt-3 text-sm leading-6 text-zinc-600 dark:text-zinc-400">
                    {item.answer}
                  </p>
                </details>
              ))}
            </div>
          </section>

          <section className="flex flex-col gap-6 rounded-2xl border border-zinc-200 p-7 sm:flex-row sm:items-center sm:justify-between dark:border-white/10">
            <div className="flex items-center gap-4">
              <Image
                src={author.avatar}
                alt={author.name}
                width={56}
                height={56}
                className="rounded-2xl ring-1 ring-zinc-950/10 dark:ring-white/15"
              />
              <div>
                <p className="text-xs text-zinc-500">Built and supported by</p>
                <h2 className="mt-1 font-semibold text-zinc-950 dark:text-white">
                  {author.name}
                </h2>
              </div>
            </div>
            <Link
              href={`/developers/${author.username}`}
              className="inline-flex items-center gap-2 text-sm font-semibold text-zinc-700 dark:text-zinc-300"
            >
              Meet the developer
              <ArrowRightIcon className="size-4" />
            </Link>
          </section>
        </div>

        <div id="purchase" className="lg:sticky lg:top-24 lg:self-start">
          <PricingCard product={product} />
          <div className="mt-4 grid grid-cols-3 divide-x divide-zinc-200 rounded-2xl border border-zinc-200 px-2 py-4 text-center dark:divide-white/10 dark:border-white/10">
            <div>
              <p className="text-sm font-semibold text-zinc-950 dark:text-white">
                {formatNumber(product.sales)}
              </p>
              <p className="mt-1 text-[10px] text-zinc-500">Customers</p>
            </div>
            <div>
              <p className="text-sm font-semibold text-zinc-950 dark:text-white">
                {product.rating}
              </p>
              <p className="mt-1 text-[10px] text-zinc-500">Rating</p>
            </div>
            <div>
              <p className="text-sm font-semibold text-zinc-950 dark:text-white">
                {product.reviewCount}
              </p>
              <p className="mt-1 text-[10px] text-zinc-500">Reviews</p>
            </div>
          </div>
        </div>
      </div>

      <section className="border-t border-zinc-200 bg-zinc-50/70 dark:border-white/10 dark:bg-white/[0.02]">
        <div className="mx-auto max-w-7xl px-5 py-16 sm:px-6 lg:px-8 lg:py-24">
          <h2 className="text-2xl font-semibold tracking-[-0.035em] text-zinc-950 dark:text-white">
            More products worth exploring
          </h2>
          <div className="mt-8 grid gap-x-6 gap-y-12 md:grid-cols-2 lg:grid-cols-3">
            {related.map((item) => (
              <ProductCard key={item.id} product={item} />
            ))}
          </div>
        </div>
      </section>
    </>
  );
}
