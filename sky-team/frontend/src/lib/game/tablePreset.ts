import type { SpaceId } from '#lib/types.js';

/**
 * The table's own UI state to start with, for the dev fixtures route only: a fixture can show a
 * die picked in the hand with a coffee modifier, a space picked, or a reroll selection.
 */
export interface TablePreset {
	/** The hand slot picked to place (or to give away, in the swap step). */
	slot?: number;
	/** The coffee modifier chosen (−3…+3). */
	coffee?: number;
	/** The space picked for the die; an intern space starts the training flow. */
	space?: SpaceId;
	/** The intern token's target, after the intern space. */
	tokenSpace?: SpaceId;
	/** Hand slots marked for the reroll. */
	rerollSlots?: number[];
}
