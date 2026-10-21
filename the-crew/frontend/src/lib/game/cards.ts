// Card helpers shared by the components: labels, keys, colours.
import { terms } from '#lib/edition.js';
import type { Card, Suit, Token } from '#lib/types.js';

export const SUIT_NAMES: Record<Suit, string> = {
	pink: 'pink',
	blue: 'blue',
	green: 'green',
	yellow: 'yellow',
	trump: terms.trump
};

/** "pink 7", "rocket 4" (Deep Sea: "submarine 4"). */
export function cardLabel(c: Card): string {
	return `${SUIT_NAMES[c.suit]} ${c.rank}`;
}

/** A stable key for keyed each blocks and card motion. */
export function cardKey(c: Card): string {
	return `${c.suit}-${c.rank}`;
}

export function sameCard(a: Card | null | undefined, b: Card | null | undefined): boolean {
	return !!a && !!b && a.suit === b.suit && a.rank === b.rank;
}

/** The Tailwind text colour class of a suit (theme tokens). */
export const SUIT_TEXT: Record<Suit, string> = {
	pink: 'text-pink',
	blue: 'text-blue',
	green: 'text-green',
	yellow: 'text-yellow',
	trump: 'text-rocket-ink'
};

/** The short label of a token: 1–5, Ω, › … ››››. */
export function tokenLabel(t: Token): string {
	if (t === 'omega') return 'Ω';
	if (t.startsWith('arrow')) return '›'.repeat(Number(t.slice(5)));
	return t;
}

/** The place of a token in the order the tasks are to be fulfilled: 1–5, then › … ››››, then
 * tasks without a token, then Ω. */
function tokenRank(t: Token): number {
	if (t === '') return 100;
	if (t === 'omega') return 200;
	if (t.startsWith('arrow')) return 10 + Number(t.slice(5));
	return Number(t);
}

/** The tasks in the order their tokens ask for (stable: tasks with equal tokens keep their order). */
export function byTokenOrder<T extends { token: Token }>(tasks: readonly T[]): T[] {
	return [...tasks].sort((a, b) => tokenRank(a.token) - tokenRank(b.token));
}

/** A sentence about a token, for titles and help. */
export function tokenMeaning(t: Token): string {
	if (t === '') return '';
	if (t === 'omega') return 'must be the last task fulfilled';
	if (t.startsWith('arrow')) {
		const n = Number(t.slice(5));
		return n === 1
			? 'before the tasks with more arrows'
			: `after the tasks with fewer arrows (${'›'.repeat(n - 1)})`;
	}
	const ord = ['first', 'second', 'third', 'fourth', 'fifth'][Number(t) - 1];
	return `must be the ${ord} task fulfilled`;
}

export type CardSize = 'xs' | 'comm' | 'task' | 'stack' | 'trick' | 'hand';
/**
 * Card sizes in px: xs 26×36 (tasks in a row, inline chips), comm 30×42 (a communicated card),
 * task 40×56 (tasks in dialogs), stack 42×60 (JARVIS's stacks), trick 52×74, hand 58×84.
 */
export const CARD_SIZE: Record<CardSize, string> = {
	xs: 'h-9 w-6.5 rounded-[0.25rem]',
	comm: 'h-[2.625rem] w-[1.875rem] rounded-[0.3125rem]',
	task: 'h-14 w-10 rounded-md',
	stack: 'h-[3.75rem] w-[2.625rem] rounded-md',
	trick: 'h-[4.625rem] w-[3.25rem] rounded-md',
	hand: 'h-[5.25rem] w-[3.625rem] rounded-md'
};
/** Sizes without the art window (Planet Nine) or with the rank and glyph stacked in the middle
 * (Deep Sea). */
export const COMPACT: ReadonlySet<CardSize> = new Set(['xs', 'comm', 'task', 'stack']);
