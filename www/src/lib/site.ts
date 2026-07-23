export const siteConfig = {
  name: "DevKit",
  title: "DevKit — Developer Software Marketplace",
  description:
    "Discover, buy, and sell exceptional software built by independent developers. Curated SaaS, AI apps, developer tools, templates, plugins, and APIs.",
  url: process.env.NEXT_PUBLIC_SITE_URL ?? "https://devkit.market",
  ogImage: "/images/og-devkit.svg",
  links: {
    github: "https://github.com/devkit-market",
    x: "https://x.com/devkitmarket",
  },
};

export function absoluteUrl(path: string) {
  return new URL(path, siteConfig.url).toString();
}
