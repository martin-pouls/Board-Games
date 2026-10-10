import type { Direction, PlayerView, Result } from '#lib/types.js';

const cards = (n: number) => `${n} ${n === 1 ? 'card' : 'cards'}`;

/** Headline for the game-over screen. */
export function resultHeadline(result: Result): string {
	if (result.score === 0) return 'You won! All 98 cards played.';
	if (result.score < 10) return `Excellent! Only ${cards(result.score)} left.`;
	return `Game over: ${cards(result.score)} left.`;
}

/** "{name} had no playable card." or null when nobody got stuck. */
export function stuckLine(view: PlayerView): string | null {
	const id = view.result?.stuckPlayerId;
	if (!id) return null;
	const p = view.players.find((x) => x.id === id);
	if (!p) return null;
	return p.isYou ? 'You had no playable card.' : `${p.name} had no playable card.`;
}

/** Short pile range label: "1 → 99" or "100 → 2". */
export function pileLabel(direction: Direction): string {
	return direction === 'up' ? '1 → 99' : '100 → 2';
}

export function plural(n: number, word: string): string {
	return `${n} ${word}${n === 1 ? '' : 's'}`;
}
