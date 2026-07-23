import Link from "next/link";
import type { Developer } from "@/types";
import { formatNumber } from "@/lib/format";

export default function DeveloperCard({ developer }: { developer: Developer }) {
  return (
    <Link
      href={`/developers/${developer.username}`}
      className="group block rounded-xl border border-surface-200 bg-white p-6 transition-all hover:shadow-md hover:-translate-y-0.5"
    >
      <div className="flex items-center gap-4">
        <img src={developer.avatar} alt={developer.name} className="h-12 w-12 rounded-full bg-surface-200" />
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.5">
            <h3 className="text-sm font-semibold text-surface-900 group-hover:text-brand-600 transition-colors truncate">
              {developer.name}
            </h3>
            {developer.verified && (
              <svg className="h-3.5 w-3.5 shrink-0 text-brand-500" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clipRule="evenodd" />
              </svg>
            )}
          </div>
          <p className="mt-0.5 text-xs text-surface-500">{developer.bio}</p>
        </div>
      </div>
      <div className="mt-4 flex gap-4 text-xs text-surface-500">
        <span className="font-medium text-surface-700">{formatNumber(developer.followers)} followers</span>
        <span>{developer.publishedCount} products</span>
      </div>
    </Link>
  );
}
