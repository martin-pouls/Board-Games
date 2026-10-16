import type { Pile } from '#lib/types.js';

// Card motion keys (see the shared cardMotion helper). A card keeps one key in every place it
// can appear (your hand, a pile's top), so a card moving between them flies. Card values are
// unique in a deck (2–99), command cards included.

/** The motion key of the card with this value. */
export function cardKey(value: number): string {
	return `card-${value}`;
}

/** The motion key of a pile's top: its top card, or the pile's own start card (1 or 100). */
export function pileTopKey(pile: Pick<Pile, 'index' | 'topCard'>): string {
	return pile.topCard ? cardKey(pile.topCard.value) : `start-${pile.index}`;
}
