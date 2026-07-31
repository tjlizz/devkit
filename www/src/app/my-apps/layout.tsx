import type { Metadata } from "next";
import { createMetadata } from "@/lib/metadata";

export const metadata: Metadata = createMetadata({
  title: "My apps",
  description: "Manage your published apps, track review status, and update listings on DevKit.",
  path: "/my-apps",
  keywords: ["devkit", "my apps", "developer marketplace"],
});

export default function MyAppsLayout({ children }: { children: React.ReactNode }) {
  return children;
}
