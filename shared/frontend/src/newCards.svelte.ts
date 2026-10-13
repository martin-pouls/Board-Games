import { untrack } from 'svelte';
import { SvelteSet } from 'svelte/reactivity';

/** How long a card that just came into the hand stays marked. */
export const NEW_CARD_MS = 8000;

/**
 * Remembers which cards just came into a hand (drawn, received as a gift, taken from a
 * neighbour), so the hand can mark them for a few seconds. Only the latest cards to come in are
 * marked: a new draw ends the marks of the one before. Cards are identified by keys that are
 * unique within the hand.
 *
 * Nothing is marked when the whole hand is new (first view after a page load, a new deal, a hand
 * swapped for another one), and a card that comes back right after it left (an undone play) is
 * not new either.
 */
export class NewCards {
	#marked = new SvelteSet<string>();
	// The rest is bookkeeping that nothing renders.
	/* eslint-disable svelte/prefer-svelte-reactivity */
	#hand: Set<string> | null = null;
	#deal: unknown = undefined;
	/** Cards that left the hand since cards were last marked. */
	#gone = new Set<string>();
	#timers = new Map<string, ReturnType<typeof setTimeout>>();
	/* eslint-enable svelte/prefer-svelte-reactivity */

	constructor(readonly ms = NEW_CARD_MS) {}

	has(key: string): boolean {
		return this.#marked.has(key);
	}

	/**
	 * Takes the current hand. `deal` identifies the deal it belongs to; when it changes, the hand
	 * starts over without marks.
	 */
	update(keys: readonly string[], deal?: unknown): void {
		// eslint-disable-next-line svelte/prefer-svelte-reactivity -- a snapshot, not rendered
		const next = new Set(keys);
		const prev = this.#hand;
		this.#hand = next;
		if (!prev || deal !== this.#deal) {
			this.#deal = deal;
			this.#reset();
			return;
		}
		for (const k of prev) if (!next.has(k)) this.#gone.add(k);
		const added = [...next].filter((k) => !prev.has(k));
		if (prev.size > 0 && added.length === next.size) {
			this.#reset();
			return;
		}
		const fresh = added.filter((k) => !this.#gone.has(k));
		if (fresh.length > 0) this.#gone.clear();
		// Only the latest cards to come in stay marked.
		for (const k of this.#marked) if (fresh.length > 0 || !next.has(k)) this.#unmark(k);
		for (const k of fresh) {
			this.#marked.add(k);
			this.#timers.set(
				k,
				setTimeout(() => this.#unmark(k), this.ms)
			);
		}
	}

	/** Stops the timers (the hand is gone). */
	dispose(): void {
		for (const t of this.#timers.values()) clearTimeout(t);
		this.#timers.clear();
	}

	#unmark(key: string): void {
		clearTimeout(this.#timers.get(key));
		this.#timers.delete(key);
		this.#marked.delete(key);
	}

	#reset(): void {
		this.dispose();
		this.#gone.clear();
		this.#marked.clear();
	}
}

/**
 * Tracks a hand component's cards; call it while the component initialises. `keys` returns the
 * keys of the cards in hand, `deal` (optional) what identifies the deal.
 */
export function trackNewCards(
	keys: () => readonly string[],
	deal: () => unknown = () => undefined
): NewCards {
	const cards = new NewCards();
	$effect.pre(() => {
		const k = keys();
		const d = deal();
		untrack(() => cards.update(k, d));
	});
	$effect(() => () => cards.dispose());
	return cards;
}
