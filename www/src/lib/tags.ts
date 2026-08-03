import type { Product } from "@/types";

export function tagToSlug(tag: string) {
  return tag
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

export function productsForTag(products: Product[], slug: string) {
  return products.filter((product) =>
    product.tags.some((tag) => tagToSlug(tag) === slug),
  );
}

export function tagsFromProducts(products: Product[]) {
  const tags = new Map<string, string>();

  for (const product of products) {
    for (const tag of product.tags) {
      const slug = tagToSlug(tag);
      if (slug && !tags.has(slug)) tags.set(slug, tag);
    }
  }

  return [...tags].map(([slug, name]) => ({ slug, name }));
}
