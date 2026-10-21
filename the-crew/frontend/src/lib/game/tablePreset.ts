import type { Card } from '#lib/types.js';

/**
 * The table's own UI state to start with, for the dev fixtures route only: a fixture can show a
 * card picked in the hand, or the communicate flow with or without a card picked.
 */
export interface TablePreset {
	/** The card picked to play (or, in the pass step, to pass). */
	selected?: Card;
	/** Communicate was tapped: the hand offers the cards that can be shown. */
	communicate?: boolean;
	/** The card picked to show (with `communicate`). */
	shown?: Card;
	/** Task ids tapped (draft, token swap). */
	tasks?: number[];
}
