export type CategorySlug =
  | "saas"
  | "ai-applications"
  | "developer-tools"
  | "templates"
  | "plugins"
  | "apis"
  | "open-source";

export interface SocialLink {
  label: "GitHub" | "X" | "LinkedIn";
  url: string;
}

export interface Developer {
  id: string;
  name: string;
  username: string;
  avatar: string;
  bio: string;
  longBio: string;
  location: string;
  website: string;
  socialLinks: SocialLink[];
  verified: boolean;
  joinedAt: string;
  publishedCount: number;
  followers: number;
  totalSales: number;
  revenue: number;
  specialties: string[];
}

export interface ProductFeature {
  title: string;
  description: string;
}

export interface ProductFaq {
  question: string;
  answer: string;
}

export interface ChangelogEntry {
  version: string;
  date: string;
  title: string;
  notes: string;
}

export interface Product {
  id: string;
  name: string;
  slug: string;
  tagline: string;
  description: string;
  longDescription: string;
  coverImage: string;
  screenshots: string[];
  category: CategorySlug;
  tags: string[];
  price: number;
  priceLabel?: string;
  authorUsername: string;
  createdAt: string;
  updatedAt: string;
  sales: number;
  favorites: number;
  rating: number;
  reviewCount: number;
  featured: boolean;
  license: string;
  delivery: string;
  support: string;
  features: ProductFeature[];
  techStack: string[];
  useCases: string[];
  demoUrl: string;
  docsUrl: string;
  changelog: ChangelogEntry[];
  faq: ProductFaq[];
}

export interface Category {
  name: string;
  shortName: string;
  slug: CategorySlug;
  description: string;
  longDescription: string;
  icon: string;
  accent: string;
  productCount: number;
}
