import Image from "next/image";

interface ProductScreenshotsProps {
  images: string[];
  productName: string;
}

export function ProductScreenshots({ images, productName }: ProductScreenshotsProps) {
  return (
    <section aria-labelledby="screenshots-heading">
      <div className="mb-5 flex items-end justify-between">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500">
            Inside the product
          </p>
          <h2
            id="screenshots-heading"
            className="mt-2 text-2xl font-semibold tracking-[-0.035em] text-zinc-950 dark:text-white"
          >
            Designed around your work
          </h2>
        </div>
        <p className="hidden text-xs text-zinc-500 sm:block">Scroll to explore →</p>
      </div>
      <div className="screenshot-scroll flex snap-x snap-mandatory gap-4 overflow-x-auto pb-4">
        {images.map((image, index) => (
          <figure
            key={image}
            className="relative aspect-[16/10] min-w-[88%] snap-center overflow-hidden rounded-2xl border border-zinc-200 bg-zinc-100 shadow-sm sm:min-w-[75%] dark:border-white/10 dark:bg-zinc-900"
          >
            <Image
              src={image}
              alt={`${productName} screenshot ${index + 1}`}
              fill
              sizes="(max-width: 768px) 88vw, 60vw"
              className="object-cover"
            />
          </figure>
        ))}
      </div>
    </section>
  );
}
