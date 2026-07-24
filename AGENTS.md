# DevKit — Developer Software Marketplace

## Project Overview

DevKit is a developer software marketplace platform with three layers:

- **server/** — Go HTTP API with SQLite (admin backend)
- **web/** — Vue 3 + TypeScript + Ant Design Vue (admin panel)
- **www/** — Next.js public-facing marketplace (App Router, SSR, SEO-optimized)

## Frontend Goals (www/)

Build a professional developer works showcase and trading platform (类似 Product Hunt + Gumroad + GitHub Marketplace 的结合). Developers can onboard, create portfolios, publish software products (SaaS, source code, plugins, templates, tools, AI apps, APIs), and sell them.

**Core design values**: Premium, professional, trustworthy, commercial-grade. Think Stripe/Vercel/Linear/Product Hunt/Gumroad visual quality.

**Do NOT design as**:
- A tool navigation site
- A cheap SaaS template site
- A software download site
- A Chinese resource aggregation site

## Tech Stack (www/)

- Next.js (latest, App Router)
- TypeScript
- Tailwind CSS
- Server Components first
- SSR for SEO
- Static export support where applicable

## SEO Requirements

Every page must be an independent SEO landing page:

- Dynamic `metadata` (`title`, `description`, `keywords`, `openGraph`, `twitter`, `canonical`)
- `sitemap.xml`
- `robots.txt`
- JSON-LD structured data
- Indexable: developer profiles, product detail pages, category pages, tag pages

## Page Specifications

### Homepage (`/`)
- Hero: title "Discover and Buy Developer-Created Software", subtitle, CTA buttons
- Featured Products (cards: logo, name, description, author, category, price, sales, favorites)
- Trending Developers (avatar, name, bio, published count, followers)
- Categories (SaaS, AI Apps, Dev Tools, Templates, Plugins, APIs, Open Source)
- Why choose (buyer-facing: verified devs, secure purchase, lifetime updates, pro support; seller-facing: publish, reach customers, manage sales)

### Product Detail (`/products/[slug]`)
- SaaS-product-website style landing page
- Logo, name, tagline, author info, price, Buy Now
- Screenshots showcase, detailed description, features, tech stack, use cases, demo, docs, changelog, FAQ
- Sidebar: price, buy button, license info, support channels

### Developer Profile (`/developers/[username]`)
- GitHub Profile + Indie Hacker style
- Avatar, bio, website, social links, product list, sales data, revenue stats

### Category Page (`/category/[slug]`)
- SEO-friendly with intro text, featured products, latest products

### Search Page (`/search`)
- Product search, developer search, tag search

## Design Language

**Colors**: Black/white/gray primary, minimal brand accent
**Typography**: Modern sans-serif
**Layout**: Max 1280px, generous whitespace, rounded cards, micro-animations, high-quality images
**Interactions**: Hover effects, smooth animations, favorites, sharing, ratings, follow

## Data Models (frontend mock phase)

### User
- Developer / Buyer roles

### Product
- name, slug, description, coverImage, screenshots[], category, tags[], price, author, createdAt, updatedAt

## Payment (reserved for future)
- Stripe, PayPal, WeChat Pay, Alipay

## Phase 1
Frontend-only with mock data. Clean architecture, scalable to production SaaS.

## Deployment

Current production deployment on this host:

- Public marketplace: `http://devkit.280303.xyz`
- Developer admin: `http://admin.devkit.280303.xyz`
- Nginx vhost: `/etc/nginx/sites-available/devkit`, enabled via `/etc/nginx/sites-enabled/devkit`
- Public Next.js service: `devkit-www.service`, running from `/root/devkit/www` on `127.0.0.1:13000`
- Public Next.js build env: `/root/devkit/www/.env.production`
- Go API service: `devkit-api.service`, running `/root/devkit/server/bin/devkit-server` on port `18080`
- API environment: `/etc/devkit/server.env`
- SMTP is configured in `/etc/devkit/server.env` with `DEVKIT_SMTP_HOST=send.ahasend.com`, `DEVKIT_SMTP_PORT=587`, `DEVKIT_SMTP_USERNAME`, `DEVKIT_SMTP_PASSWORD`, and `DEVKIT_SMTP_FROM`
- SQLite database: `/var/lib/devkit/devkit.db`
- Admin static files served by nginx from `/var/www/devkit/admin`

Update deployment after code changes:

```sh
cd /root/devkit
make build
cp -a web/dist/. /var/www/devkit/admin/
systemctl restart devkit-api.service
systemctl restart devkit-www.service
/usr/sbin/nginx -t
systemctl reload nginx
```

Nginx routing:

- `devkit.280303.xyz` proxies all traffic to the Next.js service on `127.0.0.1:13000`.
- `admin.devkit.280303.xyz` serves the Vue build from `/var/www/devkit/admin`.
- `admin.devkit.280303.xyz/api/` proxies to the Go API on `127.0.0.1:18080`.

Do not edit unrelated nginx server blocks when updating DevKit. Keep future DevKit nginx changes inside `/etc/nginx/sites-available/devkit`.

---

## Prompt (for reference when implementing)

你是一名资深产品设计师和 Next.js 全栈工程师，请设计并实现一个专业的开发者作品交易平台。

项目定位：这是一个面向全球开发者的作品展示与销售平台（类似 Product Hunt + Gumroad + GitHub Marketplace 的结合），开发者可以入驻平台，创建个人主页，发布自己的软件、源码、SaaS、插件、模板、工具、AI 应用等数字作品，并进行销售。

注意：
- 不要设计成普通工具导航站。
- 不要有廉价 SaaS 模板感。
- 不要像软件下载站。
- 整体应该像一个专业的软件产品交易市场。

核心设计目标：高端、专业、信任感、商业化、面向开发者和企业用户、类似成熟海外 SaaS Marketplace 的视觉体验。
