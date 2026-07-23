import type { Developer } from "@/types";

export const developers: Developer[] = [
  {
    id: "dev_01",
    name: "Maya Chen",
    username: "mayachen",
    avatar: "/images/avatars/maya.svg",
    bio: "Independent maker building calm, powerful tools for product teams.",
    longBio:
      "I’m a product engineer and independent maker based in Vancouver. After a decade building analytics and collaboration products, I now focus on small, opinionated software that helps teams understand what matters and move with confidence.",
    location: "Vancouver, Canada",
    website: "https://maya.build",
    socialLinks: [
      { label: "GitHub", url: "https://github.com/mayachen" },
      { label: "X", url: "https://x.com/mayabuilds" },
    ],
    verified: true,
    joinedAt: "2023-04-12",
    publishedCount: 4,
    followers: 8420,
    totalSales: 3842,
    revenue: 186400,
    specialties: ["Product analytics", "TypeScript", "SaaS"],
  },
  {
    id: "dev_02",
    name: "Noah Williams",
    username: "noahw",
    avatar: "/images/avatars/noah.svg",
    bio: "Design engineer crafting beautiful infrastructure for developers.",
    longBio:
      "Design engineer, open-source maintainer, and terminal enthusiast. I create durable tools that make sophisticated engineering workflows feel simple, without hiding the details that experts care about.",
    location: "London, United Kingdom",
    website: "https://noahw.dev",
    socialLinks: [
      { label: "GitHub", url: "https://github.com/noahw" },
      { label: "LinkedIn", url: "https://linkedin.com/in/noahw" },
    ],
    verified: true,
    joinedAt: "2022-11-03",
    publishedCount: 6,
    followers: 12600,
    totalSales: 5186,
    revenue: 268900,
    specialties: ["Developer experience", "Go", "Open source"],
  },
  {
    id: "dev_03",
    name: "Sofia Ramirez",
    username: "sofiaramirez",
    avatar: "/images/avatars/sofia.svg",
    bio: "AI product builder turning complex research into useful software.",
    longBio:
      "I build approachable AI products for researchers and creative teams. My work combines applied machine learning with careful interaction design, privacy-conscious architecture, and excellent documentation.",
    location: "Austin, United States",
    website: "https://sofia.works",
    socialLinks: [
      { label: "GitHub", url: "https://github.com/sofiaramirez" },
      { label: "X", url: "https://x.com/sofiaworks" },
    ],
    verified: true,
    joinedAt: "2023-08-19",
    publishedCount: 3,
    followers: 6310,
    totalSales: 2940,
    revenue: 142700,
    specialties: ["Applied AI", "Python", "Product design"],
  },
  {
    id: "dev_04",
    name: "Eli Novak",
    username: "elinovak",
    avatar: "/images/avatars/eli.svg",
    bio: "Frontend systems, templates, and tiny details that make products sing.",
    longBio:
      "I’m a frontend architect based in Prague. I publish production-ready foundations, component systems, and practical resources for small teams that value both speed and craft.",
    location: "Prague, Czechia",
    website: "https://eli.codes",
    socialLinks: [
      { label: "GitHub", url: "https://github.com/elinovak" },
      { label: "X", url: "https://x.com/elicodes" },
    ],
    verified: true,
    joinedAt: "2024-01-22",
    publishedCount: 5,
    followers: 4780,
    totalSales: 1765,
    revenue: 98600,
    specialties: ["Design systems", "Next.js", "Templates"],
  },
];

export const developerByUsername = Object.fromEntries(
  developers.map((developer) => [developer.username, developer]),
) as Record<string, Developer>;
