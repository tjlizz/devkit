import { developerByUsername } from "@/lib/mock/developers";
import { products } from "@/lib/mock/products";
import type { CategorySlug, Developer, Product } from "@/types";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL ||
  process.env.API_PROXY_TARGET ||
  "http://localhost:8080";

interface ApiApp {
  id: number;
  developerId: number;
  developerName: string;
  name: string;
  slug: string;
  tagline: string;
  description: string;
  category: CategorySlug;
  priceCents: number;
  currency: string;
  iconUrl: string;
  coverImageUrl: string;
  demoUrl: string;
  docsUrl: string;
  sourceUrl: string;
  supportUrl: string;
  tags: string[];
  version: string;
  releaseNotes: string;
  publishedAt?: string;
  createdAt: string;
  updatedAt: string;
}

function apiUrl(path: string) {
  return `${API_BASE.replace(/\/$/, "")}${path}`;
}

function realDeveloper(app: ApiApp): Developer {
  return {
    id: `developer_${app.developerId}`,
    name: app.developerName,
    username: `developer-${app.developerId}`,
    avatar: "/images/avatars/noah.svg",
    bio: "Verified DevKit marketplace developer.",
    longBio: "A verified DevKit developer publishing reviewed software for modern teams.",
    location: "",
    website: app.docsUrl || app.demoUrl || app.sourceUrl,
    socialLinks: [],
    verified: true,
    joinedAt: app.createdAt,
    publishedCount: 1,
    followers: 0,
    totalSales: 0,
    revenue: 0,
    specialties: app.tags.slice(0, 3),
  };
}

function toProduct(app: ApiApp): Product {
  return {
    id: String(app.id),
    name: app.name,
    slug: app.slug,
    tagline: app.tagline,
    description: app.tagline,
    longDescription: app.description,
    coverImage: app.coverImageUrl || app.iconUrl || "/images/products/terminal-cover.svg",
    screenshots: [app.coverImageUrl || app.iconUrl || "/images/products/terminal-cover.svg"],
    category: app.category,
    tags: app.tags,
    price: app.priceCents / 100,
    priceLabel: app.priceCents === 0 ? "free" : "one-time",
    authorUsername: `developer-${app.developerId}`,
    authorName: app.developerName,
    authorAvatar: "/images/avatars/noah.svg",
    createdAt: app.publishedAt || app.createdAt,
    updatedAt: app.updatedAt,
    sales: 0,
    favorites: 0,
    rating: 5,
    reviewCount: 0,
    featured: true,
    license: "Commercial",
    delivery: app.sourceUrl ? "Source access" : "Developer-provided delivery",
    support: app.supportUrl || "Developer support",
    features: [
      { title: "Reviewed listing", description: "Approved by DevKit before publication." },
      { title: "Current version", description: `Published version ${app.version}.` },
    ],
    techStack: app.tags,
    useCases: app.tags.length > 0 ? app.tags : [app.category],
    demoUrl: app.demoUrl,
    docsUrl: app.docsUrl,
    changelog: app.releaseNotes
      ? [{ version: app.version, date: app.updatedAt, title: "Marketplace release", notes: app.releaseNotes }]
      : [],
    faq: [],
  };
}

export async function getMarketplaceProducts(params: { category?: string; q?: string } = {}) {
  const searchParams = new URLSearchParams();
  if (params.category) searchParams.set("category", params.category);
  if (params.q) searchParams.set("q", params.q);
  try {
    const res = await fetch(apiUrl(`/api/v1/marketplace/apps?${searchParams.toString()}`), {
      cache: "no-store",
    });
    if (!res.ok) return products;
    const data = (await res.json()) as { apps: ApiApp[] };
    return data.apps.length > 0 ? data.apps.map(toProduct) : products;
  } catch {
    return products;
  }
}

export async function getMarketplaceProduct(slug: string) {
  try {
    const res = await fetch(apiUrl(`/api/v1/marketplace/apps/${encodeURIComponent(slug)}`), {
      cache: "no-store",
    });
    if (res.ok) {
      const data = (await res.json()) as { app: ApiApp };
      return { product: toProduct(data.app), developer: realDeveloper(data.app) };
    }
  } catch {
  }
  const product = products.find((item) => item.slug === slug);
  return {
    product,
    developer: product ? developerByUsername[product.authorUsername] : undefined,
  };
}
