/**
 * Client-side mirror of the Go domain rules (backend/internal/core/domain), for UX only:
 * the server is authoritative and re-validates everything. docs/rules.md is the source of truth.
 */
import type { Card, PlayerView, Suit } from '#lib/types.js';

export const SUITS: Suit[] = ['hearts', 'diamonds', 'clubs', 'spades'];

export const isJester = (c: Card): boolean => c.rank === 0;
export const isCompanion = (c: Card): boolean => c.rank === 1;
export const isNumber = (c: Card): boolean => c.rank >= 2 && c.rank <= 10;
export const isRoyal = (c: Card): boolean => c.rank >= 11 && c.rank <= 13;

/** Value in hand, used for both attacking and discarding. */
export function cardValue(c: Card): number {
	if (isRoyal(c)) return 10 + (c.rank - 11) * 5;
	return c.rank;
}

const ENEMY_STATS: Record<number, { attack: number; health: number }> = {
	11: { attack: 10, health: 20 },
	12: { attack: 15, health: 30 },
	13: { attack: 20, health: 40 }
};

/** Printed attack and health of an enemy rank (zeros for non-enemies). */
export function enemyStats(rank: number): { attack: number; health: number } {
	return ENEMY_STATS[rank] ?? { attack: 0, health: 0 };
}

export const handValue = (cs: Card[]): number => cs.reduce((sum, c) => sum + cardValue(c), 0);

/** Rank ascending (Jester, A, 2..10, J, Q, K), ties by suit order. */
export function compareCards(a: Card, b: Card): number {
	if (a.rank !== b.rank) return a.rank - b.rank;
	return SUITS.indexOf(a.suit as Suit) - SUITS.indexOf(b.suit as Suit);
}

/** Returns a sorted copy. */
export const sortCards = (cs: Card[]): Card[] => [...cs].sort(compareCards);

export const cardKey = (c: Card): string => `${c.suit || 'jester'}-${c.rank}`;

export type PlayKind = 'single' | 'combo' | 'companion' | 'jester';
export type PlayProblem =
	| 'empty'
	| 'jesterNotAlone'
	| 'companionPair'
	| 'royalCombo'
	| 'mixedRanks'
	| 'tooManyCards'
	| 'comboTooHigh';
export type PlayCheck =
	{ ok: true; kind: PlayKind; value: number } | { ok: false; problem: PlayProblem };

/** Shape check of a play (rules §5); same order and reasons as Go's ClassifyPlay. */
export function checkPlay(cards: Card[]): PlayCheck {
	const fail = (problem: PlayProblem): PlayCheck => ({ ok: false, problem });
	const value = handValue(cards);
	if (cards.length === 0) return fail('empty');
	if (cards.some(isJester)) {
		return cards.length === 1 ? { ok: true, kind: 'jester', value } : fail('jesterNotAlone');
	}
	if (cards.length === 1) return { ok: true, kind: 'single', value };
	if (cards.some(isCompanion)) {
		return cards.length === 2 ? { ok: true, kind: 'companion', value } : fail('companionPair');
	}
	if (cards.some(isRoyal)) return fail('royalCombo');
	if (cards.some((c) => c.rank !== cards[0].rank)) return fail('mixedRanks');
	if (cards.length > 4) return fail('tooManyCards');
	if (value > 10) return fail('comboTooHigh');
	return { ok: true, kind: 'combo', value };
}

/** amount: heal/draw/shield count; for clubs the extra damage. */
export interface SuitEffect {
	suit: Suit;
	blocked: boolean;
	amount: number;
}

export interface PlayPreview {
	check: PlayCheck;
	value: number;
	/** Suits in the play, SUITS order. */
	effects: SuitEffect[];
	/** Value x2 if clubs are active; 0 for a Jester. */
	damage: number;
	heal: number;
	draw: number;
	shieldAfter: number;
	damageAfter: number;
	defeats: boolean;
	exact: boolean;
	/** Step 4 damage if the enemy survives. */
	attackAfter: number;
	/** True when this play is a Jester and the enemy was still immune. */
	cancelsImmunity: boolean;
}

/** Previews what playing `cards` would do. Null without an enemy. */
export function previewPlay(view: PlayerView, cards: Card[]): PlayPreview | null {
	const enemy = view.enemy;
	if (!enemy) return null;
	const check = checkPlay(cards);
	const value = handValue(cards);
	const isJesterPlay = check.ok && check.kind === 'jester';
	const immune: Suit | null = enemy.immunityCancelled ? null : enemy.immunity;

	const suits =
		check.ok && !isJesterPlay ? SUITS.filter((s) => cards.some((c) => c.suit === s)) : [];
	const active = (s: Suit) => suits.includes(s) && immune !== s;

	const heal = active('hearts') ? Math.min(value, view.discardCount) : 0;

	let draw = 0;
	if (active('diamonds')) {
		let free = 0;
		for (const p of view.players) {
			const count = p.isYou ? p.handCount - cards.length : p.handCount;
			free += Math.max(0, view.rules.maxHandSize - count);
		}
		draw = Math.max(0, Math.min(value, view.tavernCount + heal, free));
	}

	const damage = active('clubs') ? value * 2 : isJesterPlay ? 0 : value;

	// A Jester against a spade enemy makes earlier spades count retroactively (rules §6).
	let shield = enemy.shield;
	if (isJesterPlay && immune === 'spades') {
		shield = enemy.plays
			.filter((p) => p.cards.some((c) => c.suit === 'spades'))
			.reduce((sum, p) => sum + p.value, 0);
	}
	const shieldAfter = shield + (active('spades') ? value : 0);

	const damageAfter = enemy.damage + damage;
	const defeats = check.ok && !isJesterPlay && damageAfter >= enemy.health;

	return {
		check,
		value,
		effects: suits.map((suit) => ({
			suit,
			blocked: !active(suit),
			amount: suit === 'hearts' ? heal : suit === 'diamonds' ? draw : active(suit) ? value : 0
		})),
		damage,
		heal,
		draw,
		shieldAfter,
		damageAfter,
		defeats,
		exact: defeats && damageAfter === enemy.health,
		attackAfter: Math.max(0, enemy.attack - shieldAfter),
		cancelsImmunity: isJesterPlay && !enemy.immunityCancelled
	};
}

export type DiscardCheck =
	| { ok: true; sum: number }
	| { ok: false; sum: number; problem: 'empty' | 'notEnough' | 'superfluous' };

/**
 * Step 4 discard set legality (rules §4 App decision): sum >= required, and removing the
 * smallest card must drop below required (no superfluous card). `required` is > 0.
 */
export function checkDiscard(cards: Card[], required: number): DiscardCheck {
	const sum = handValue(cards);
	if (cards.length === 0) return { ok: false, sum, problem: 'empty' };
	if (sum < required) return { ok: false, sum, problem: 'notEnough' };
	const min = Math.min(...cards.map(cardValue));
	if (sum - min >= required) return { ok: false, sum, problem: 'superfluous' };
	return { ok: true, sum };
}

/**
 * Whether `id` could yield right now: solo always, else some other player did not yield on
 * their last turn (rules §8). UX hint only; the server decides.
 */
export function canPlayerYield(
	players: { id: string; yieldedLastTurn: boolean }[],
	id: string
): boolean {
	if (players.length <= 1) return true;
	return players.some((p) => p.id !== id && !p.yieldedLastTurn);
}
