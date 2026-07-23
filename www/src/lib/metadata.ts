import type { Metadata } from "next";
import { absoluteUrl, siteConfig } from "@/lib/site";

interface MetadataInput {
  title: string;
  description: string;
  path: string;
  keywords: string[];
  image?: string;
}

export function createMetadata({
  title,
  description,
  path,
  keywords,
  image = siteConfig.ogImage,
}: MetadataInput): Metadata {
  const canonical = absoluteUrl(path);
  const imageUrl = absoluteUrl(image);

  return {
    title,
    description,
    keywords,
    alternates: { canonical },
    openGraph: {
      type: "website",
      siteName: siteConfig.name,
      title,
      description,
      url: canonical,
      images: [{ url: imageUrl, width: 1200, height: 630, alt: title }],
    },
    twitter: {
      card: "summary_large_image",
      title,
      description,
      images: [imageUrl],
    },
  };
}
