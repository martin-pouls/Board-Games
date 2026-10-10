import type { Direction, PlayerView, Result } from '#lib/types.js';

const cards = (n: number) => `${n} ${n === 1 ? 'card' : 'cards'}`;

/** Headline for the game-over screen. */
export function resultHeadline(result: Result): string {
	if (result.outcome === 'aborted') return `Game ended early: ${cards(result.score)} left.`;
	// Extreme: the last card was a Skull that could not be covered.
	if (result.outcome === 'lost' && result.score === 0)
		return 'So close! All cards played, but a skull stayed uncovered.';
	if (result.score === 0) return 'You won! All 98 cards played.';
	if (result.score < 10) return `Excellent! Only ${cards(result.score)} left.`;
	return `Game over: ${cards(result.score)} left.`;
}

/**
 * Why the game was lost, naming the stuck player ("You" for the viewer):
 * an uncovered skull, an unfinished 3!, or "had no playable card".
 * Null when nobody got stuck.
 */
export function stuckLine(view: PlayerView): string | null {
	const id = view.result?.stuckPlayerId;
	if (!id) return null;
	const p = view.players.find((x) => x.id === id);
	if (!p) return null;
	const who = p.isYou ? 'You' : p.name;
	const turn = view.turn;
	if (turn?.constraints.skullPiles.length) return `${who} couldn't cover the skull.`;
	const exact = turn?.constraints.exactCount ?? null;
	if (turn && exact !== null && turn.cardsPlayed < exact) return `${who} couldn't complete 3!.`;
	return `${who} had no playable card.`;
}

/** Short pile range label: "1 → 99" or "100 → 2". */
export function pileLabel(direction: Direction): string {
	return direction === 'up' ? '1 → 99' : '100 → 2';
}

export function plural(n: number, word: string): string {
	return `${n} ${word}${n === 1 ? '' : 's'}`;
}

/** "{host} ended the game." when the host ended it early, else null. */
export function endedLine(view: PlayerView): string | null {
	if (view.result?.outcome !== 'aborted') return null;
	const host = view.players.find((x) => x.id === view.hostId);
	if (!host) return 'The host ended the game.';
	return host.isYou ? 'You ended the game.' : `${host.name} ended the game.`;
}
