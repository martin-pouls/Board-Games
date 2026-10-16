import type { Direction, PlayerView } from '#lib/types.js';

const cards = (n: number) => `${n} ${n === 1 ? 'card' : 'cards'}`;

/** Headline for the game-over screen (won or lost). */
export function resultHeadline(view: Pick<PlayerView, 'phase' | 'result'>): string {
	const result = view.result;
	if (view.phase === 'won') return 'You won! All 98 cards played.';
	if (!result) return 'Game over';
	// Extreme: the last card was a Skull that could not be covered.
	if (result.score === 0) return 'So close! All cards played, but a skull stayed uncovered.';
	if (result.score < 10) return `Excellent! Only ${cards(result.score)} left.`;
	return `Game over: ${cards(result.score)} left.`;
}

/**
 * Why the game was lost, naming the player ("You" for the viewer): "{who} gave up." or, when
 * stuck at the start of a turn, "{who} had no playable card." Null when won or unknown.
 */
export function lossLine(view: PlayerView): string | null {
	const r = view.result;
	if (view.phase !== 'lost' || !r?.reason) return null;
	const p = r.playerId ? view.players.find((x) => x.id === r.playerId) : undefined;
	if (r.reason === 'gaveUp') {
		if (!p) return 'The game was given up.';
		return `${p.isYou ? 'You' : p.name} gave up.`;
	}
	if (!p) return null;
	return `${p.isYou ? 'You' : p.name} had no playable card.`;
}

/** Short pile range label: "1 → 99" or "100 → 2". */
export function pileLabel(direction: Direction): string {
	return direction === 'up' ? '1 → 99' : '100 → 2';
}

export function plural(n: number, word: string): string {
	return `${n} ${word}${n === 1 ? '' : 's'}`;
}
