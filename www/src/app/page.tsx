import type { Metadata } from "next";
import Link from "next/link";
import {
  ArrowRightIcon,
  CheckIcon,
  PackageIcon,
  ShieldIcon,
  SparklesIcon,
  UsersIcon,
} from "@/components/icons";
import { CategorySection } from "@/components/category-section";
import DeveloperCard from "@/components/DeveloperCard";
import { HeroSection } from "@/components/hero-section";
import ProductCard from "@/components/ProductCard";
import { SectionHeading } from "@/components/ui/section-heading";
import { JsonLd } from "@/lib/json-ld";
import { createMetadata } from "@/lib/metadata";
import { developers } from "@/lib/mock/developers";
import { featuredProducts } from "@/lib/mock/products";
import { absoluteUrl, siteConfig } from "@/lib/site";

export const metadata: Metadata = createMetadata({
  title: "Developer-Created Software Marketplace",
  description:
    "Discover and buy exceptional SaaS, AI apps, developer tools, templates, plugins, and APIs directly from verified independent developers.",
  path: "/",
  keywords: [
    "developer software marketplace",
    "buy SaaS products",
    "independent developer tools",
    "software templates",
    "AI applications",
  ],
});

const buyerBenefits = [
  "Products reviewed for quality and clarity",
  "Secure checkout with buyer protection",
  "Direct updates from the developer",
  "Responsive, expert product support",
];

const sellerBenefits = [
  "A premium home for your best work",
  "Reach teams ready to buy software",
  "Simple publishing and sales analytics",
  "Keep ownership of your audience",
];

export default function HomePage() {
  const jsonLd = [
    {
      "@context": "https://schema.org",
      "@type": "WebSite",
      name: siteConfig.name,
      url: siteConfig.url,
      description: siteConfig.description,
      potentialAction: {
        "@type": "SearchAction",
        target: `${absoluteUrl("/search")}?q={search_term_string}`,
        "query-input": "required name=search_term_string",
      },
    },
    {
      "@context": "https://schema.org",
      "@type": "ItemList",
      name: "Featured developer products",
      itemListElement: featuredProducts.map((product, index) => ({
        "@type": "ListItem",
        position: index + 1,
        url: absoluteUrl(`/products/${product.slug}`),
        name: product.name,
      })),
    },
  ];

  return (
    <>
      <JsonLd data={jsonLd} />
      <HeroSection />

      <section className="mx-auto max-w-7xl px-5 py-20 sm:px-6 lg:px-8 lg:py-28">
        <SectionHeading
          eyebrow="Editor’s selection"
          title="Software worth building on"
          description="Original products with real depth—selected for thoughtful design, durable technology, and developer-led support."
          link={{ label: "View all products", href: "/search" }}
        />
        <div className="grid gap-x-6 gap-y-12 md:grid-cols-2 lg:grid-cols-3">
          {featuredProducts.map((product) => (
            <ProductCard key={product.id} product={product} />
          ))}
        </div>
      </section>

      <section className="border-y border-zinc-200 bg-zinc-50/70 dark:border-white/10 dark:bg-white/[0.02]">
        <div className="mx-auto max-w-7xl px-5 py-20 sm:px-6 lg:px-8 lg:py-28">
          <SectionHeading
            eyebrow="Explore by craft"
            title="Built for how modern teams work"
            description="From focused micro-SaaS to open infrastructure, find a product shaped by someone who cares about the details."
          />
          <CategorySection />
        </div>
      </section>

      <section className="mx-auto max-w-7xl px-5 py-20 sm:px-6 lg:px-8 lg:py-28">
        <SectionHeading
          eyebrow="People behind the products"
          title="Meet developers worth following"
          description="Independent builders with deep domain expertise, responsive support, and a track record of shipping."
          link={{
            label: "Discover developers",
            href: "/developers/mayachen",
          }}
        />
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {developers.map((developer) => (
            <DeveloperCard key={developer.id} developer={developer} />
          ))}
        </div>
      </section>

      <section className="border-t border-zinc-200 dark:border-white/10">
        <div className="mx-auto max-w-7xl px-5 py-20 sm:px-6 lg:px-8 lg:py-28">
          <SectionHeading
            eyebrow="A better marketplace"
            title="Built for trust on both sides"
            description="DevKit gives buyers confidence and gives developers a serious place to grow a software business."
          />
          <div className="grid overflow-hidden rounded-3xl border border-zinc-200 lg:grid-cols-2 dark:border-white/10">
            <div className="relative bg-white p-7 sm:p-10 lg:p-12 dark:bg-zinc-950">
              <div className="absolute top-0 right-0 h-40 w-40 rounded-full bg-blue-500/10 blur-3xl" />
              <ShieldIcon className="size-8 text-blue-600 dark:text-blue-400" />
              <p className="mt-8 text-xs font-semibold uppercase tracking-[0.15em] text-zinc-500">
                For buyers
              </p>
              <h3 className="mt-3 text-2xl font-semibold tracking-[-0.04em] text-zinc-950 dark:text-white">
                Buy with informed confidence
              </h3>
              <p className="mt-4 max-w-md text-sm leading-6 text-zinc-600 dark:text-zinc-400">
                Understand exactly what you&apos;re getting, who built it, and how it
                will be supported before you purchase.
              </p>
              <ul className="mt-8 grid gap-4 sm:grid-cols-2">
                {buyerBenefits.map((benefit) => (
                  <li
                    key={benefit}
                    className="flex items-start gap-2.5 text-sm text-zinc-700 dark:text-zinc-300"
                  >
                    <CheckIcon className="mt-0.5 size-4 shrink-0 text-blue-600 dark:text-blue-400" />
                    {benefit}
                  </li>
                ))}
              </ul>
            </div>
            <div className="relative border-t border-zinc-200 bg-zinc-950 p-7 text-white sm:p-10 lg:border-t-0 lg:border-l lg:p-12 dark:border-white/10 dark:bg-white dark:text-zinc-950">
              <div className="absolute top-0 right-0 h-40 w-40 rounded-full bg-violet-500/20 blur-3xl" />
              <PackageIcon className="size-8 text-violet-400 dark:text-violet-600" />
              <p className="mt-8 text-xs font-semibold uppercase tracking-[0.15em] text-zinc-500">
                For developers
              </p>
              <h3 className="mt-3 text-2xl font-semibold tracking-[-0.04em]">
                Turn your craft into a business
              </h3>
              <p className="mt-4 max-w-md text-sm leading-6 text-zinc-400 dark:text-zinc-600">
                Present your work beautifully, reach thoughtful customers, and
                manage sales without surrendering your identity.
              </p>
              <ul className="mt-8 grid gap-4 sm:grid-cols-2">
                {sellerBenefits.map((benefit) => (
                  <li
                    key={benefit}
                    className="flex items-start gap-2.5 text-sm text-zinc-300 dark:text-zinc-700"
                  >
                    <CheckIcon className="mt-0.5 size-4 shrink-0 text-violet-400 dark:text-violet-600" />
                    {benefit}
                  </li>
                ))}
              </ul>
            </div>
          </div>
          <div className="mt-6 grid gap-4 sm:grid-cols-3">
            {[
              {
                icon: SparklesIcon,
                title: "Curated, never crowded",
                text: "Every listing is reviewed for originality, quality, and honest positioning.",
              },
              {
                icon: UsersIcon,
                title: "Creator-direct support",
                text: "Questions go to the people who understand the product all the way down.",
              },
              {
                icon: ShieldIcon,
                title: "Commercial-grade trust",
                text: "Clear licenses, secure payments, verified makers, and transparent updates.",
              },
            ].map((item) => (
              <div
                key={item.title}
                className="rounded-2xl border border-zinc-200 p-6 dark:border-white/10"
              >
                <item.icon className="size-5 text-zinc-500" />
                <h3 className="mt-5 text-sm font-semibold text-zinc-950 dark:text-white">
                  {item.title}
                </h3>
                <p className="mt-2 text-sm leading-6 text-zinc-500">{item.text}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-7xl px-5 pb-20 sm:px-6 lg:px-8 lg:pb-28">
        <div className="relative overflow-hidden rounded-3xl bg-zinc-950 px-6 py-16 text-center text-white sm:px-12 dark:bg-white dark:text-zinc-950">
          <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,_rgba(139,92,246,0.25),_transparent_55%)]" />
          <div className="relative">
            <p className="text-xs font-semibold uppercase tracking-[0.17em] text-zinc-400 dark:text-zinc-500">
              Your next favorite tool is here
            </p>
            <h2 className="mx-auto mt-4 max-w-2xl text-balance text-3xl font-semibold tracking-[-0.05em] sm:text-4xl">
              Find software made with care by people who ship.
            </h2>
            <Link
              href="/search"
              className="mt-8 inline-flex h-11 items-center gap-2 rounded-full bg-white px-5 text-sm font-semibold text-zinc-950 transition hover:-translate-y-0.5 dark:bg-zinc-950 dark:text-white"
            >
              Explore the marketplace
              <ArrowRightIcon className="size-4" />
            </Link>
          </div>
        </div>
      </section>
    </>
  );
}
