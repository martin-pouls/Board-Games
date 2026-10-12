/** Display helpers: card labels and rule explanations. */
import type { Card, Grade, PlayedSet, PlayerView, Result, Suit } from '#lib/types.js';
import { cardValue, type DiscardCheck, type PlayProblem } from './rules.js';

const GLYPHS: Record<Suit, string> = { hearts: '♥', diamonds: '♦', clubs: '♣', spades: '♠' };
const SUIT_NAMES: Record<Suit, string> = {
	hearts: 'Hearts',
	diamonds: 'Diamonds',
	clubs: 'Clubs',
	spades: 'Spades'
};

export const suitGlyph = (s: Suit): string => GLYPHS[s];
export const isRed = (s: Suit | ''): boolean => s === 'hearts' || s === 'diamonds';

export function rankLabel(r: number): string {
	if (r === 0) return 'Jester';
	if (r === 1) return 'A';
	if (r === 11) return 'J';
	if (r === 12) return 'Q';
	if (r === 13) return 'K';
	return String(r);
}

/** "5♥", "A♠", "Jester". */
export function cardLabel(c: Card): string {
	return c.suit === '' ? 'Jester' : `${rankLabel(c.rank)}${GLYPHS[c.suit]}`;
}

const RANK_NAMES: Record<number, string> = { 1: 'Ace', 11: 'Jack', 12: 'Queen', 13: 'King' };

/** "5 of Hearts", "Queen of Spades", "Jester". */
export function cardName(c: Card): string {
	if (c.suit === '') return 'Jester';
	return `${RANK_NAMES[c.rank] ?? c.rank} of ${SUIT_NAMES[c.suit]}`;
}

const POWER_NAMES: Record<Suit, string> = {
	hearts: 'Heal',
	diamonds: 'Draw',
	clubs: 'Double damage',
	spades: 'Shield'
};
export const suitPowerName = (s: Suit): string => POWER_NAMES[s];

/** The plays against an enemy as one stack per player, in the order they first played. */
export function playStacks(plays: PlayedSet[]): { playerId: string; cards: Card[] }[] {
	const byPlayer = new Map<string, Card[]>();
	for (const p of plays)
		byPlayer.set(p.playerId, [...(byPlayer.get(p.playerId) ?? []), ...p.cards]);
	return [...byPlayer].map(([playerId, cards]) => ({ playerId, cards }));
}

export function plural(n: number, word: string): string {
	return `${n} ${word}${n === 1 ? '' : 's'}`;
}

const PROBLEM_TEXTS: Record<PlayProblem, string> = {
	empty: 'Select a card to play.',
	jesterNotAlone: 'The Jester must be played on its own.',
	companionPair: 'An Animal Companion can be paired with exactly one other card.',
	royalCombo: 'Jacks, Queens and Kings cannot be part of a combo.',
	mixedRanks: 'A combo needs cards of the same rank.',
	tooManyCards: 'A combo has at most 4 cards.',
	comboTooHigh: 'A combo can total at most 10.'
};
export const playProblemText = (p: PlayProblem): string => PROBLEM_TEXTS[p];

/** Explains the state of a discard selection against the damage to cover. */
export function discardProblemText(check: DiscardCheck, required: number, cards: Card[]): string {
	if (check.ok) return `${check.sum} covers ${required} damage.`;
	switch (check.problem) {
		case 'empty':
			return `Select cards worth at least ${required}.`;
		case 'notEnough':
			return `Need ${required - check.sum} more (${check.sum} of ${required}).`;
		case 'superfluous': {
			const smallest = cards.reduce((a, b) => (cardValue(b) < cardValue(a) ? b : a), cards[0]);
			return `Too many cards: ${cardLabel(smallest)} isn't needed.`;
		}
	}
}

/** "You" for the viewer, else the player's name. */
export function who(view: PlayerView, id: string): string {
	if (id === view.you) return 'You';
	return view.players.find((p) => p.id === id)?.name ?? 'Someone';
}

const GRADE_LABELS: Record<Grade, string> = {
	gold: 'Gold victory',
	silver: 'Silver victory',
	bronze: 'Bronze victory'
};
export const gradeLabel = (g: Grade): string => GRADE_LABELS[g];

export function resultHeadline(result: Result): string {
	if (result.outcome === 'won') {
		return result.grade
			? `${gradeLabel(result.grade)}! The realm is saved.`
			: 'Victory! The realm is saved.';
	}
	return `Defeat: ${result.enemiesDefeated} of 12 enemies slain.`;
}

/** Why the game was lost, for the game-over screen. */
export function lossText(view: PlayerView): string {
	const r = view.result;
	if (!r || r.outcome !== 'lost') return '';
	const name = r.playerId ? who(view, r.playerId) : 'Someone';
	switch (r.reason) {
		case 'cannotCoverDamage': {
			const dmg = view.enemy?.effectiveAttack;
			return dmg ? `${name} couldn't take ${dmg} damage.` : `${name} couldn't take the damage.`;
		}
		case 'cannotAct':
			return `${name} had no cards and couldn't yield.`;
		case 'gaveUp':
			return `${name} gave up.`;
		default:
			return 'The game was lost.';
	}
}
