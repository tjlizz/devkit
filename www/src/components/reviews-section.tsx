"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { StarIcon } from "@/components/icons";
import { useAuth, type Review } from "@/lib/auth-context";
import { formatDate } from "@/lib/format";

function Stars({ rating, size = "size-4" }: { rating: number; size?: string }) {
  return (
    <span className="flex items-center gap-0.5" aria-label={`${rating} out of 5 stars`}>
      {[1, 2, 3, 4, 5].map((value) => (
        <StarIcon
          key={value}
          className={`${size} ${
            value <= Math.round(rating)
              ? "fill-current text-amber-500"
              : "fill-none text-zinc-300 dark:text-zinc-600"
          }`}
        />
      ))}
    </span>
  );
}

export function ReviewsSection({ slug }: { slug: string }) {
  const {
    isAuthenticated,
    loading: authLoading,
    user,
    listMyEntitlements,
    listAppReviews,
    getMyReview,
    saveReview,
    deleteReview,
  } = useAuth();

  const [reviews, setReviews] = useState<Review[] | null>(null);
  const [myReview, setMyReview] = useState<Review | null>(null);
  const [canReview, setCanReview] = useState(false);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [rating, setRating] = useState(5);
  const [comment, setComment] = useState("");
  const [notice, setNotice] = useState("");

  const load = useCallback(async () => {
    setError("");
    try {
      const [list, mine] = await Promise.all([
        listAppReviews(slug),
        isAuthenticated ? getMyReview(slug).catch(() => null) : Promise.resolve(null),
      ]);
      setReviews(list);
      setMyReview(mine);
      if (mine) {
        setRating(mine.rating);
        setComment(mine.comment);
      }
      if (isAuthenticated) {
        const ents = await listMyEntitlements();
        setCanReview(ents.some((ent) => ent.appSlug === slug && ent.status === "active"));
      } else {
        setCanReview(false);
      }
    } catch (err: any) {
      setError(err?.message || "Could not load reviews");
    }
  }, [slug, isAuthenticated, listAppReviews, getMyReview, listMyEntitlements]);

  useEffect(() => {
    load();
  }, [load]);

  async function handleSubmit() {
    setSubmitting(true);
    setError("");
    setNotice("");
    try {
      const saved = await saveReview(slug, rating, comment.trim());
      setMyReview(saved);
      setReviews((current) => {
        const rest = (current || []).filter((item) => item.id !== saved.id);
        return [saved, ...rest];
      });
      setNotice("Review saved.");
    } catch (err: any) {
      setError(err?.message || "Could not save your review");
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete() {
    setSubmitting(true);
    setError("");
    setNotice("");
    try {
      await deleteReview(slug);
      const removedId = myReview?.id;
      setMyReview(null);
      setReviews((current) =>
        (current || []).filter((item) => item.id !== removedId),
      );
      setNotice("Review removed.");
    } catch (err: any) {
      setError(err?.message || "Could not delete your review");
    } finally {
      setSubmitting(false);
    }
  }

  const showForm = isAuthenticated && canReview;
  const loading = reviews === null;

  return (
    <section id="reviews" className="scroll-mt-24">
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500">
            Reviews
          </p>
          <h2 className="mt-3 text-3xl font-semibold tracking-[-0.045em] text-zinc-950 dark:text-white">
            What buyers say
          </h2>
        </div>
        {!loading && reviews && reviews.length > 0 && (
          <p className="text-sm text-zinc-500">
            {reviews.length} {reviews.length === 1 ? "review" : "reviews"}
          </p>
        )}
      </div>

      {error ? (
        <div className="mt-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-300">
          {error}
        </div>
      ) : null}
      {notice ? (
        <div className="mt-6 rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700 dark:border-emerald-800 dark:bg-emerald-950 dark:text-emerald-300">
          {notice}
        </div>
      ) : null}

      {loading ? (
        <p className="mt-6 text-sm text-zinc-500">Loading reviews...</p>
      ) : (
        <div className="mt-8 space-y-8">
          {showForm && (
            <form
              onSubmit={(event) => {
                event.preventDefault();
                handleSubmit();
              }}
              className="rounded-2xl border border-zinc-200 bg-zinc-50 p-6 sm:p-8 dark:border-white/10 dark:bg-white/[0.035]"
            >
              <div className="flex flex-wrap items-center justify-between gap-4">
                <div>
                  <p className="text-sm font-semibold text-zinc-950 dark:text-white">
                    {myReview ? "Update your review" : "Review this app"}
                  </p>
                  <p className="mt-1 text-xs text-zinc-500">
                    You&apos;re a verified buyer. Share your experience.
                  </p>
                </div>
                <div className="flex items-center gap-1" role="radiogroup" aria-label="Rating">
                  {[1, 2, 3, 4, 5].map((value) => (
                    <button
                      key={value}
                      type="button"
                      onClick={() => setRating(value)}
                      aria-label={`${value} star${value > 1 ? "s" : ""}`}
                      className="p-0.5"
                    >
                      <StarIcon
                        className={`size-6 transition ${
                          value <= rating
                            ? "fill-current text-amber-500"
                            : "fill-none text-zinc-300 dark:text-zinc-600"
                        }`}
                      />
                    </button>
                  ))}
                </div>
              </div>
              <textarea
                value={comment}
                onChange={(event) => setComment(event.target.value)}
                maxLength={2000}
                rows={4}
                placeholder="What did you like (or dislike)? What would you tell another buyer?"
                className="mt-5 w-full resize-y rounded-xl border border-zinc-200 bg-white px-4 py-3 text-sm text-zinc-900 placeholder:text-zinc-400 focus:border-zinc-400 focus:outline-none dark:border-white/10 dark:bg-zinc-950 dark:text-white dark:placeholder:text-zinc-500"
              />
              <div className="mt-5 flex items-center justify-between gap-4">
                <p className="text-xs text-zinc-400">{comment.length}/2000</p>
                <div className="flex items-center gap-3">
                  {myReview && (
                    <button
                      type="button"
                      onClick={handleDelete}
                      disabled={submitting}
                      className="text-sm font-medium text-zinc-500 transition hover:text-red-600 disabled:opacity-50"
                    >
                      Delete
                    </button>
                  )}
                  <button
                    type="submit"
                    disabled={submitting || comment.trim().length === 0}
                    className="inline-flex h-10 items-center rounded-lg bg-zinc-950 px-5 text-sm font-semibold text-white transition hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-white dark:text-zinc-950 dark:hover:bg-zinc-200"
                  >
                    {submitting ? "Saving..." : myReview ? "Update review" : "Post review"}
                  </button>
                </div>
              </div>
            </form>
          )}

          {isAuthenticated && !authLoading && !canReview && !myReview && (
            <div className="rounded-2xl border border-dashed border-zinc-300 px-6 py-8 text-center dark:border-white/15">
              <p className="text-sm text-zinc-600 dark:text-zinc-400">
                Buy this app to leave a verified review.
              </p>
            </div>
          )}

          {!isAuthenticated && (
            <div className="rounded-2xl border border-dashed border-zinc-300 px-6 py-8 text-center dark:border-white/15">
              <p className="text-sm text-zinc-600 dark:text-zinc-400">
                <Link href={`/login?next=/products/${slug}#reviews`} className="font-medium text-zinc-950 underline underline-offset-2 dark:text-white">
                  Sign in
                </Link>{" "}
                to leave a review.
              </p>
            </div>
          )}

          {reviews.length === 0 ? (
            <p className="text-sm text-zinc-500">No reviews yet. Be the first to share your experience.</p>
          ) : (
            <ul className="divide-y divide-zinc-200 dark:divide-white/10">
              {reviews.map((review) => (
                <li key={review.id} className="py-6 first:pt-0">
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <div className="flex items-center gap-3">
                      <span className="flex size-9 items-center justify-center rounded-full bg-zinc-100 text-sm font-semibold text-zinc-600 dark:bg-white/10 dark:text-zinc-300">
                        {review.buyerName.charAt(0).toUpperCase()}
                      </span>
                      <div>
                        <p className="flex items-center gap-2 text-sm font-semibold text-zinc-950 dark:text-white">
                          {review.buyerName}
                          {review.verifiedPurchase && (
                            <span className="rounded-full bg-emerald-50 px-2 py-0.5 text-[10px] font-medium text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-300">
                              Verified purchase
                            </span>
                          )}
                        </p>
                        <p className="mt-0.5 text-xs text-zinc-400">{formatDate(review.createdAt)}</p>
                      </div>
                    </div>
                    <Stars rating={review.rating} />
                  </div>
                  {review.comment ? (
                    <p className="mt-3 max-w-3xl text-sm leading-6 text-zinc-600 dark:text-zinc-400">
                      {review.comment}
                    </p>
                  ) : null}
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </section>
  );
}