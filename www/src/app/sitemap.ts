import type { MetadataRoute } from "next";
import { categories } from "@/lib/mock/categories";
import { developers } from "@/lib/mock/developers";
import { products } from "@/lib/mock/products";
import { absoluteUrl } from "@/lib/site";

export default function sitemap(): MetadataRoute.Sitemap {
  const now = new Date();

  return [
    {
      url: absoluteUrl("/"),
      lastModified: now,
      changeFrequency: "daily",
      priority: 1,
    },
    {
      url: absoluteUrl("/search"),
      lastModified: now,
      changeFrequency: "daily",
      priority: 0.8,
    },
    ...products.map((product) => ({
      url: absoluteUrl(`/products/${product.slug}`),
      lastModified: new Date(product.updatedAt),
      changeFrequency: "weekly" as const,
      priority: product.featured ? 0.9 : 0.8,
      images: product.screenshots.map(absoluteUrl),
    })),
    ...developers.map((developer) => ({
      url: absoluteUrl(`/developers/${developer.username}`),
      lastModified: now,
      changeFrequency: "weekly" as const,
      priority: 0.7,
      images: [absoluteUrl(developer.avatar)],
    })),
    ...categories.map((category) => ({
      url: absoluteUrl(`/category/${category.slug}`),
      lastModified: now,
      changeFrequency: "weekly" as const,
      priority: 0.75,
    })),
  ];
}
