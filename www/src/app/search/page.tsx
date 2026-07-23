import type { Metadata } from "next";
import Navbar from "@/components/Navbar";
import Footer from "@/components/Footer";
import ProductCard from "@/components/ProductCard";
import DeveloperCard from "@/components/DeveloperCard";
import { products } from "@/lib/mock/products";
import { developers } from "@/lib/mock/developers";
import { createMetadata } from "@/lib/metadata";

interface Props {
  searchParams: Promise<{ q?: string }>
}

export async function generateMetadata(): Promise<Metadata> {
  return createMetadata({
    title: "Search Products",
    description: "Search developer-created software, tools, templates, and more.",
    path: "/search",
    keywords: ["search", "developer marketplace"],
  });
}

export default async function SearchPage({ searchParams }: Props) {
  const { q } = await searchParams;
  const query = (q || "").trim().toLowerCase();

  const filteredProducts = query
    ? products.filter(
        (p) =>
          p.name.toLowerCase().includes(query) ||
          p.tagline.toLowerCase().includes(query) ||
          p.description.toLowerCase().includes(query) ||
          p.tags.some((t) => t.toLowerCase().includes(query)) ||
          p.category.toLowerCase().includes(query),
      )
    : products;

  const filteredDevelopers = query
    ? developers.filter(
        (d) =>
          d.name.toLowerCase().includes(query) ||
          d.bio.toLowerCase().includes(query) ||
          d.specialties.some((s) => s.toLowerCase().includes(query)),
      )
    : developers;

  return (
    <>
      <Navbar />
      <main className="mx-auto max-w-7xl px-6 py-12 lg:py-16">
        <h1 className="text-3xl font-bold tracking-tight text-surface-900">
          {query ? `Results for "${q}"` : "Browse All Products"}
        </h1>
        {query && (
          <p className="mt-2 text-sm text-surface-500">
            {filteredProducts.length} product{filteredProducts.length !== 1 ? "s" : ""} found
          </p>
        )}

        {filteredProducts.length > 0 ? (
          <>
            <h2 className="mt-10 text-lg font-semibold text-surface-900">Products</h2>
            <div className="mt-4 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {filteredProducts.map((product) => (
                <ProductCard key={product.id} product={product} />
              ))}
            </div>
          </>
        ) : query ? (
          <p className="mt-8 text-sm text-surface-500">No products found for "{q}". Try different keywords.</p>
        ) : null}

        {filteredDevelopers.length > 0 && query && (
          <>
            <h2 className="mt-14 text-lg font-semibold text-surface-900">Developers</h2>
            <div className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
              {filteredDevelopers.map((dev) => (
                <DeveloperCard key={dev.id} developer={dev} />
              ))}
            </div>
          </>
        )}
      </main>
      <Footer />
    </>
  );
}
