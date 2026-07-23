import type { Category } from "@/types";

export const categories: Category[] = [
  {
    name: "Software as a Service",
    shortName: "SaaS",
    slug: "saas",
    description: "Production-ready software for modern teams.",
    longDescription:
      "Discover focused SaaS products built by experienced independent developers—from analytics and operations to collaboration and customer success.",
    icon: "S",
    accent: "from-violet-500/20 to-indigo-500/5",
    productCount: 142,
  },
  {
    name: "AI Applications",
    shortName: "AI Apps",
    slug: "ai-applications",
    description: "Practical AI products that improve real workflows.",
    longDescription:
      "Browse thoughtfully designed AI applications for research, content, automation, coding, and knowledge work—made to deliver real outcomes.",
    icon: "A",
    accent: "from-cyan-500/20 to-blue-500/5",
    productCount: 98,
  },
  {
    name: "Developer Tools",
    shortName: "Dev Tools",
    slug: "developer-tools",
    description: "Build, debug, ship, and operate with confidence.",
    longDescription:
      "Explore polished developer tools for local development, testing, observability, deployment, and the daily craft of building software.",
    icon: "D",
    accent: "from-emerald-500/20 to-teal-500/5",
    productCount: 214,
  },
  {
    name: "Templates",
    shortName: "Templates",
    slug: "templates",
    description: "High-quality foundations for your next launch.",
    longDescription:
      "Start from robust, design-conscious templates for apps, dashboards, landing pages, documentation, and commerce experiences.",
    icon: "T",
    accent: "from-amber-500/20 to-orange-500/5",
    productCount: 187,
  },
  {
    name: "Plugins",
    shortName: "Plugins",
    slug: "plugins",
    description: "Powerful extensions for the tools you already use.",
    longDescription:
      "Extend your favorite design, development, productivity, and commerce platforms with carefully crafted plugins.",
    icon: "P",
    accent: "from-fuchsia-500/20 to-pink-500/5",
    productCount: 121,
  },
  {
    name: "APIs",
    shortName: "APIs",
    slug: "apis",
    description: "Reliable building blocks for ambitious products.",
    longDescription:
      "Integrate proven APIs for data, automation, media, security, and infrastructure—documented and supported by their creators.",
    icon: "{ }",
    accent: "from-sky-500/20 to-cyan-500/5",
    productCount: 76,
  },
  {
    name: "Open Source",
    shortName: "Open Source",
    slug: "open-source",
    description: "Sustainable open source with expert support.",
    longDescription:
      "Find maintained open-source software with commercial licenses, hosted options, and support plans that keep great projects thriving.",
    icon: "O",
    accent: "from-zinc-500/20 to-zinc-500/5",
    productCount: 164,
  },
];

export const categoryBySlug = Object.fromEntries(
  categories.map((category) => [category.slug, category]),
) as Record<string, Category>;
