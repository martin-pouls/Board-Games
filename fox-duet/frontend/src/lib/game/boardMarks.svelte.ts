import { untrack } from 'svelte';
import type { Board } from '#lib/types.js';

/** What changed on the board since the last card was led, per space index. */
export interface Marks {
	/** Gems that arrived at a location (round-end gems, gems moved by a forest token). */
	added: Record<number, number>;
	/** Spaces a forest token just covered. */
	covered: number[];
}

const none = (): Marks => ({ added: {}, covered: [] });

/**
 * Marks board changes the players didn't make by moving the traveller, so the board explains
 * itself without a narrating line: gems added at round end or moved by a forest token, and new
 * forest tokens. Marks add up over several views and end when the next card is led. Nothing is
 * marked on the first board after a page load or for a new game.
 */
export class BoardMarks {
	current = $state<Marks>(none());
	#prev: Board | null = null;
	#game: unknown = undefined;

	update(board: Board | null, led: boolean, game?: unknown): void {
		const prev = this.#prev;
		this.#prev = board;
		if (!board || !prev || game !== this.#game || prev.spaces.length !== board.spaces.length) {
			this.#game = game;
			this.current = none();
			return;
		}
		if (led) {
			if (this.current.covered.length > 0 || Object.keys(this.current.added).length > 0) {
				this.current = none();
			}
			return;
		}
		const added = { ...this.current.added };
		const covered = [...this.current.covered];
		let changed = false;
		board.spaces.forEach((s, i) => {
			const more = s.gems - prev.spaces[i].gems;
			if (more > 0) {
				added[i] = (added[i] ?? 0) + more;
				changed = true;
			}
			if (s.covered && !prev.spaces[i].covered) {
				covered.push(i);
				changed = true;
			}
		});
		if (changed) this.current = { added, covered };
	}
}

/** Tracks the board of a view; call it while the component initialises. */
export function trackBoardMarks(
	board: () => Board | null,
	led: () => boolean,
	game: () => unknown
): BoardMarks {
	const marks = new BoardMarks();
	$effect.pre(() => {
		const b = board();
		const l = led();
		const g = game();
		untrack(() => marks.update(b, l, g));
	});
	return marks;
}
