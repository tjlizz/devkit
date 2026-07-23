import { notFound } from "next/navigation";
import type { Metadata } from "next";
import Link from "next/link";
import Navbar from "@/components/Navbar";
import Footer from "@/components/Footer";
import ProductCard from "@/components/ProductCard";
import { categories, categoryBySlug } from "@/lib/mock/categories";
import { products } from "@/lib/mock/products";
import { createMetadata } from "@/lib/metadata";

interface Props { params: Promise<{ slug: string }> }

export async function generateStaticParams() {
  return categories.map((c) => ({ slug: c.slug }));
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  const cat = categoryBySlug[slug];
  if (!cat) return {};
  return createMetadata({
    title: `${cat.shortName} — Browse Products`,
    description: cat.longDescription,
    path: `/category/${cat.slug}`,
    keywords: [cat.shortName, cat.name, "category"],
  });
}

export default async function CategoryPage({ params }: Props) {
  const { slug } = await params;
  const cat = categoryBySlug[slug];
  if (!cat) notFound();

  const catProducts = products.filter((p) => p.category === cat.slug);

  return (
    <>
      <Navbar />
      <main className="mx-auto max-w-7xl px-6 py-12 lg:py-16">
        <div className="flex items-center gap-2 text-sm text-surface-500">
          <Link href="/" className="hover:text-surface-700">Home</Link>
          <span>/</span>
          <span className="text-surface-900">{cat.shortName}</span>
        </div>

        <div className="mt-6 max-w-2xl">
          <h1 className="text-3xl font-bold tracking-tight text-surface-900 sm:text-4xl">{cat.shortName}</h1>
          <p className="mt-3 text-base leading-relaxed text-surface-500">{cat.longDescription}</p>
          <p className="mt-2 text-sm text-surface-400">{catProducts.length} {catProducts.length === 1 ? "product" : "products"}</p>
        </div>

        <div className="mt-10 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {catProducts.map((product) => (
            <ProductCard key={product.id} product={product} />
          ))}
        </div>
      </main>
      <Footer />
    </>
  );
}
