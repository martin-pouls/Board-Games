// Card helpers shared by the components: labels, keys, colours.
import type { Card, Suit, Token } from '#lib/types.js';

export const SUIT_NAMES: Record<Suit, string> = {
	pink: 'pink',
	blue: 'blue',
	green: 'green',
	yellow: 'yellow',
	trump: 'rocket'
};

/** "pink 7", "rocket 4". */
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

export type CardSize = 'xs' | 'task' | 'stack' | 'sm' | 'md';
/** xs: inline chips; task: task cards; stack: JARVIS's stacks; sm: the hand; md: the trick. */
export const CARD_SIZE: Record<CardSize, string> = {
	xs: 'h-9 w-6.5 rounded-[0.3rem] p-px',
	task: 'h-14 w-10 rounded-md p-0.5',
	stack: 'h-[3.6rem] w-[2.6rem] rounded-md p-0.5',
	sm: 'h-[4.75rem] w-[3.25rem] rounded-lg p-1',
	md: 'h-[5.25rem] w-[3.6rem] rounded-lg p-1'
};
/** Sizes that show the rank and glyph stacked in the middle instead of in the corner. */
export const COMPACT: ReadonlySet<CardSize> = new Set(['xs', 'task', 'stack']);
