import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import { notFound } from "next/navigation";
import {
  ArrowUpRightIcon,
  CalendarIcon,
  CheckIcon,
  ExternalLinkIcon,
  GlobeIcon,
  MapPinIcon,
} from "@/components/icons";
import ProductCard from "@/components/ProductCard";
import { formatCurrency, formatDate, formatNumber } from "@/lib/format";
import { JsonLd } from "@/lib/json-ld";
import { createMetadata } from "@/lib/metadata";
import { developerByUsername, developers } from "@/lib/mock/developers";
import { products } from "@/lib/mock/products";
import { absoluteUrl, siteConfig } from "@/lib/site";

interface DeveloperPageProps {
  params: Promise<{ username: string }>;
}

export function generateStaticParams() {
  return developers.map((developer) => ({ username: developer.username }));
}

export async function generateMetadata({
  params,
}: DeveloperPageProps): Promise<Metadata> {
  const { username } = await params;
  const developer = developerByUsername[username];

  if (!developer) {
    return createMetadata({
      title: "Developer not found",
      description: "This DevKit developer profile could not be found.",
      path: `/developers/${username}`,
      keywords: ["independent software developers"],
    });
  }

  return createMetadata({
    title: `${developer.name} (@${developer.username}) — Developer Profile`,
    description: developer.bio,
    path: `/developers/${developer.username}`,
    keywords: [
      developer.name,
      "independent developer",
      ...developer.specialties,
      "software maker",
    ],
    image: developer.avatar,
  });
}

export default async function DeveloperPage({ params }: DeveloperPageProps) {
  const { username } = await params;
  const developer = developerByUsername[username];

  if (!developer) notFound();

  const developerProducts = products.filter(
    (product) => product.authorUsername === developer.username,
  );

  const jsonLd = [
    {
      "@context": "https://schema.org",
      "@type": "Person",
      name: developer.name,
      alternateName: `@${developer.username}`,
      description: developer.bio,
      image: absoluteUrl(developer.avatar),
      url: absoluteUrl(`/developers/${developer.username}`),
      sameAs: [
        developer.website,
        ...developer.socialLinks.map((link) => link.url),
      ],
      jobTitle: "Independent Software Developer",
      knowsAbout: developer.specialties,
    },
    {
      "@context": "https://schema.org",
      "@type": "ProfilePage",
      name: `${developer.name} on DevKit`,
      description: developer.bio,
      url: absoluteUrl(`/developers/${developer.username}`),
      dateCreated: developer.joinedAt,
      mainEntity: {
        "@type": "Person",
        name: developer.name,
      },
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
          name: developer.name,
          item: absoluteUrl(`/developers/${developer.username}`),
        },
      ],
    },
  ];

  return (
    <>
      <JsonLd data={jsonLd} />
      <section className="border-b border-zinc-200 dark:border-white/10">
        <div className="relative mx-auto max-w-7xl px-5 py-14 sm:px-6 lg:px-8 lg:py-20">
          <div className="absolute top-0 right-10 h-64 w-64 rounded-full bg-violet-500/10 blur-3xl" />
          <nav className="relative mb-10 flex items-center gap-2 text-xs text-zinc-500">
            <Link href="/">Marketplace</Link>
            <span>/</span>
            <span className="text-zinc-800 dark:text-zinc-300">Developers</span>
            <span>/</span>
            <span className="text-zinc-800 dark:text-zinc-300">@{developer.username}</span>
          </nav>
          <div className="relative grid gap-10 lg:grid-cols-[1fr_auto] lg:items-end">
            <div className="flex flex-col gap-7 sm:flex-row sm:items-start">
              <Image
                src={developer.avatar}
                alt={developer.name}
                width={112}
                height={112}
                priority
                className="rounded-3xl shadow-xl ring-1 ring-zinc-950/10 dark:ring-white/15"
              />
              <div className="max-w-2xl">
                <div className="flex flex-wrap items-center gap-3">
                  <h1 className="text-4xl font-semibold tracking-[-0.05em] text-zinc-950 sm:text-5xl dark:text-white">
                    {developer.name}
                  </h1>
                  {developer.verified ? (
                    <span
                      className="flex size-6 items-center justify-center rounded-full bg-blue-600 text-white"
                      title="Verified developer"
                    >
                      <CheckIcon className="size-3.5" />
                    </span>
                  ) : null}
                </div>
                <p className="mt-2 text-sm font-medium text-zinc-500">
                  @{developer.username}
                </p>
                <p className="mt-5 text-base leading-7 text-zinc-600 sm:text-lg dark:text-zinc-400">
                  {developer.bio}
                </p>
                <div className="mt-6 flex flex-wrap gap-x-5 gap-y-3 text-xs text-zinc-500">
                  <span className="flex items-center gap-1.5">
                    <MapPinIcon className="size-4" />
                    {developer.location}
                  </span>
                  <Link
                    href={developer.website}
                    className="flex items-center gap-1.5 transition hover:text-zinc-950 dark:hover:text-white"
                  >
                    <GlobeIcon className="size-4" />
                    {new URL(developer.website).hostname}
                  </Link>
                  <span className="flex items-center gap-1.5">
                    <CalendarIcon className="size-4" />
                    Joined {formatDate(developer.joinedAt)}
                  </span>
                </div>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <button className="inline-flex h-11 items-center rounded-full bg-zinc-950 px-5 text-sm font-semibold text-white transition hover:-translate-y-0.5 hover:bg-zinc-800 dark:bg-white dark:text-zinc-950 dark:hover:bg-zinc-200">
                Follow developer
              </button>
              <button
                className="flex size-11 items-center justify-center rounded-full border border-zinc-200 text-zinc-600 transition hover:bg-zinc-50 dark:border-white/10 dark:text-zinc-300 dark:hover:bg-white/5"
                aria-label="Share profile"
              >
                <ArrowUpRightIcon className="size-4" />
              </button>
            </div>
          </div>
        </div>
      </section>

      <div className="mx-auto grid max-w-7xl gap-12 px-5 py-14 sm:px-6 lg:grid-cols-[260px_minmax(0,1fr)] lg:px-8 lg:py-20">
        <aside>
          <div className="lg:sticky lg:top-24">
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500">
              Maker profile
            </p>
            <p className="mt-5 text-sm leading-7 text-zinc-600 dark:text-zinc-400">
              {developer.longBio}
            </p>
            <div className="mt-7 flex flex-wrap gap-2">
              {developer.specialties.map((specialty) => (
                <span
                  key={specialty}
                  className="rounded-full border border-zinc-200 px-3 py-1.5 text-xs text-zinc-600 dark:border-white/10 dark:text-zinc-400"
                >
                  {specialty}
                </span>
              ))}
            </div>
            <div className="mt-8 border-t border-zinc-200 pt-6 dark:border-white/10">
              <p className="text-xs font-semibold text-zinc-950 dark:text-white">Elsewhere</p>
              <div className="mt-3 space-y-2">
                {developer.socialLinks.map((link) => (
                  <Link
                    key={link.label}
                    href={link.url}
                    className="flex items-center justify-between rounded-lg py-1 text-sm text-zinc-600 transition hover:text-zinc-950 dark:text-zinc-400 dark:hover:text-white"
                  >
                    {link.label}
                    <ExternalLinkIcon className="size-3.5" />
                  </Link>
                ))}
              </div>
            </div>
          </div>
        </aside>

        <div>
          <section>
            <div className="flex flex-col justify-between gap-6 sm:flex-row sm:items-end">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500">
                  Storefront
                </p>
                <h2 className="mt-3 text-3xl font-semibold tracking-[-0.045em] text-zinc-950 dark:text-white">
                  Products by {developer.name.split(" ")[0]}
                </h2>
              </div>
              <p className="text-sm text-zinc-500">
                {developerProducts.length} published on DevKit
              </p>
            </div>
            {developerProducts.length ? (
              <div className="mt-10 grid gap-x-6 gap-y-12 md:grid-cols-2">
                {developerProducts.map((product) => (
                  <ProductCard key={product.id} product={product} />
                ))}
              </div>
            ) : (
              <div className="mt-10 rounded-2xl border border-dashed border-zinc-300 p-12 text-center text-sm text-zinc-500 dark:border-white/15">
                New work is on the way.
              </div>
            )}
          </section>

          <section className="mt-20">
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500">
              Track record
            </p>
            <h2 className="mt-3 text-3xl font-semibold tracking-[-0.045em] text-zinc-950 dark:text-white">
              Building in public, delivering in full
            </h2>
            <div className="mt-8 grid overflow-hidden rounded-2xl border border-zinc-200 sm:grid-cols-2 lg:grid-cols-4 dark:border-white/10">
              {[
                {
                  value: developer.publishedCount,
                  label: "Published products",
                },
                {
                  value: formatNumber(developer.totalSales),
                  label: "Customer sales",
                },
                {
                  value: formatNumber(developer.followers),
                  label: "Followers",
                },
                {
                  value: formatCurrency(developer.revenue, true),
                  label: "Maker revenue",
                },
              ].map((stat, index) => (
                <div
                  key={stat.label}
                  className={`p-6 ${index ? "border-t border-zinc-200 sm:border-t-0 sm:border-l dark:border-white/10" : ""}`}
                >
                  <p className="text-2xl font-semibold tracking-[-0.04em] text-zinc-950 dark:text-white">
                    {stat.value}
                  </p>
                  <p className="mt-2 text-xs text-zinc-500">{stat.label}</p>
                </div>
              ))}
            </div>
            <div className="mt-6 rounded-2xl bg-zinc-50 p-7 sm:p-9 dark:bg-white/[0.035]">
              <p className="text-lg leading-8 text-zinc-700 dark:text-zinc-300">
                “The best marketplace profiles feel like a transparent studio window:
                you can see the craft, the history, and the person responsible for making
                it last.”
              </p>
              <p className="mt-4 text-xs font-semibold uppercase tracking-[0.13em] text-zinc-500">
                DevKit editorial
              </p>
            </div>
          </section>
        </div>
      </div>
    </>
  );
}
