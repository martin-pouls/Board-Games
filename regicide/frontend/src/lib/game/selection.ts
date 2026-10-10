/** Hand selection helpers for the table. Selections are sorted index lists into the hand. */
import type { Card, PlayerView } from '#lib/types.js';
import { cardKey } from './rules.js';

/**
 * Identity of the situation a selection belongs to. When it changes (new hand, new step or
 * turn, or the player may no longer act) the selection must be dropped.
 */
export function selectionKey(view: PlayerView, interactive: boolean): string {
	if (!interactive) return 'off';
	const hand = view.hand.map(cardKey).join(',');
	return `${view.turn?.number ?? 0}|${view.turn?.step ?? ''}|${hand}`;
}

/** Adds or removes `index`, keeping the list sorted. */
export function toggleIndex(selected: number[], index: number): number[] {
	return selected.includes(index)
		? selected.filter((i) => i !== index)
		: [...selected, index].sort((a, b) => a - b);
}

/** Runs one server action with the table locked meanwhile (busy flag + error toast). */
export type Runner = (action: () => Promise<void>) => Promise<void>;

/** The selected cards, ignoring indices outside the hand. */
export function pickCards(hand: Card[], selected: number[]): Card[] {
	return selected.filter((i) => i >= 0 && i < hand.length).map((i) => hand[i]);
}
