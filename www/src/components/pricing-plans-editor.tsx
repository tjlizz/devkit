"use client"

export interface PlanDraft {
  name: string
  price: string
  description: string
  features: string
}

const planPresets = [
  { name: "Free", price: "0" },
  { name: "Plus", price: "19" },
  { name: "Pro", price: "49" },
]

const inputClass =
  "mt-2 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm dark:border-white/10 dark:bg-zinc-950"

export function PricingPlansEditor({
  plans,
  onChange,
}: {
  plans: PlanDraft[]
  onChange: (plans: PlanDraft[]) => void
}) {
  function addPlan(preset?: { name: string; price: string }) {
    onChange([
      ...plans,
      { name: preset?.name ?? "", price: preset?.price ?? "0", description: "", features: "" },
    ])
  }

  function updatePlan(index: number, patch: Partial<PlanDraft>) {
    onChange(plans.map((plan, i) => (i === index ? { ...plan, ...patch } : plan)))
  }

  function removePlan(index: number) {
    onChange(plans.filter((_, i) => i !== index))
  }

  return (
    <section className="rounded-xl border border-zinc-200 p-5 dark:border-white/10">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-sm font-semibold text-zinc-950 dark:text-white">Pricing plans</h2>
          <p className="mt-1 text-xs text-zinc-500">
            Optional tiers like Free, Plus and Pro. They replace the single price on the product page.
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          {planPresets.map((preset) => (
            <button
              key={preset.name}
              type="button"
              onClick={() => addPlan(preset)}
              className="rounded-lg border border-zinc-200 px-3 py-1.5 text-xs font-semibold text-zinc-700 transition hover:bg-zinc-50 dark:border-white/10 dark:text-zinc-200 dark:hover:bg-white/5"
            >
              + {preset.name}
            </button>
          ))}
          <button
            type="button"
            onClick={() => addPlan()}
            className="rounded-lg border border-zinc-200 px-3 py-1.5 text-xs font-semibold text-zinc-700 transition hover:bg-zinc-50 dark:border-white/10 dark:text-zinc-200 dark:hover:bg-white/5"
          >
            + Custom
          </button>
        </div>
      </div>

      {plans.length === 0 ? (
        <p className="mt-4 text-xs text-zinc-400">No plans yet. Add a Free, Plus or Pro tier to get started.</p>
      ) : (
        <div className="mt-4 space-y-4">
          {plans.map((plan, index) => (
            <div key={index} className="rounded-lg border border-zinc-200 p-4 dark:border-white/10">
              <div className="flex items-center justify-between gap-3">
                <p className="text-xs font-semibold uppercase tracking-[0.14em] text-zinc-500">
                  Plan {index + 1}
                </p>
                <button
                  type="button"
                  onClick={() => removePlan(index)}
                  className="text-xs font-medium text-red-600 transition hover:text-red-500"
                >
                  Remove
                </button>
              </div>
              <div className="mt-3 grid gap-3 sm:grid-cols-3">
                <label className="text-xs font-medium text-zinc-700 dark:text-zinc-300">
                  Name
                  <input
                    className={inputClass}
                    value={plan.name}
                    onChange={(event) => updatePlan(index, { name: event.target.value })}
                    placeholder="Pro"
                    maxLength={60}
                    required
                  />
                </label>
                <label className="text-xs font-medium text-zinc-700 dark:text-zinc-300">
                  Price USD
                  <input
                    className={inputClass}
                    type="number"
                    min="0"
                    step="0.01"
                    value={plan.price}
                    onChange={(event) => updatePlan(index, { price: event.target.value })}
                  />
                </label>
                <label className="text-xs font-medium text-zinc-700 dark:text-zinc-300">
                  Features
                  <input
                    className={inputClass}
                    value={plan.features}
                    onChange={(event) => updatePlan(index, { features: event.target.value })}
                    placeholder="Unlimited projects, Priority support"
                  />
                </label>
              </div>
              <label className="mt-3 block text-xs font-medium text-zinc-700 dark:text-zinc-300">
                Description
                <input
                  className={inputClass}
                  value={plan.description}
                  onChange={(event) => updatePlan(index, { description: event.target.value })}
                  placeholder="For teams that ship."
                  maxLength={1000}
                />
              </label>
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
