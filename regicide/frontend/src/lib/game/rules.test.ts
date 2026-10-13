import { describe, expect, it } from 'vitest';
import {
	cardKey,
	cardValue,
	checkDiscard,
	checkPlay,
	compareCards,
	enemyStats,
	handKeys,
	handValue,
	isCompanion,
	isJester,
	isNumber,
	isRoyal,
	previewPlay,
	sortCards,
	SUITS
} from './rules.js';
import { card, cards, enemyOf, hand, makeView, player, playView } from './fixtures.js';
import type { Enemy, PlayerView } from '#lib/types.js';

describe('card helpers', () => {
	it('classifies cards', () => {
		expect(isJester(card('X'))).toBe(true);
		expect(isCompanion(card('AS'))).toBe(true);
		expect(isNumber(card('10H'))).toBe(true);
		expect(isNumber(card('AH'))).toBe(false);
		expect(isRoyal(card('JH'))).toBe(true);
		expect(isRoyal(card('10H'))).toBe(false);
	});

	it.each([
		['X', 0],
		['AH', 1],
		['2C', 2],
		['10D', 10],
		['JS', 10],
		['QS', 15],
		['KS', 20]
	])('value of %s is %i', (s, want) => {
		expect(cardValue(card(s))).toBe(want);
	});

	it('knows the enemy table', () => {
		expect(enemyStats(11)).toEqual({ attack: 10, health: 20 });
		expect(enemyStats(12)).toEqual({ attack: 15, health: 30 });
		expect(enemyStats(13)).toEqual({ attack: 20, health: 40 });
		expect(enemyStats(5)).toEqual({ attack: 0, health: 0 });
	});

	it('sums hand values', () => {
		expect(handValue(cards('X AH 5D KS'))).toBe(26);
		expect(handValue([])).toBe(0);
	});

	it('sorts by rank then suit order, as a copy', () => {
		const input = cards('KS 2C 2H X AD 10H JC 2S 2D');
		const sorted = sortCards(input);
		expect(sorted.map(cardKey)).toEqual([
			'jester-0',
			'diamonds-1',
			'hearts-2',
			'diamonds-2',
			'clubs-2',
			'spades-2',
			'hearts-10',
			'clubs-11',
			'spades-13'
		]);
		expect(input[0]).toEqual(card('KS'));
		expect(compareCards(card('2H'), card('2S'))).toBeLessThan(0);
		expect(compareCards(card('3H'), card('2S'))).toBeGreaterThan(0);
		expect(SUITS).toEqual(['hearts', 'diamonds', 'clubs', 'spades']);
	});

	it('builds keys', () => {
		expect(cardKey(card('X'))).toBe('jester-0');
		expect(cardKey(card('5H'))).toBe('hearts-5');
		expect(handKeys([card('X'), card('5H'), card('X')])).toEqual([
			'jester-0#0',
			'hearts-5#0',
			'jester-0#1'
		]);
	});
});

describe('checkPlay (rules §5)', () => {
	const ok = (s: string, kind: string, value: number) => ({ s, want: { ok: true, kind, value } });
	const bad = (s: string, problem: string) => ({ s, want: { ok: false, problem } });

	it.each([
		// empty
		bad('', 'empty'),
		// singles
		ok('7H', 'single', 7),
		ok('AS', 'single', 1),
		ok('JH', 'single', 10),
		ok('QD', 'single', 15),
		ok('KC', 'single', 20),
		ok('X', 'jester', 0),
		// pair combos: 2..5 ok, 6 too high
		ok('2H 2C', 'combo', 4),
		ok('3H 3S', 'combo', 6),
		ok('4D 4C', 'combo', 8),
		ok('5H 5S', 'combo', 10),
		bad('6H 6S', 'comboTooHigh'),
		bad('10H 10S', 'comboTooHigh'),
		// triples: 2, 3 ok; 4 too high
		ok('2H 2C 2S', 'combo', 6),
		ok('3H 3C 3S', 'combo', 9),
		bad('4H 4C 4S', 'comboTooHigh'),
		// quad: 2s only
		ok('2H 2D 2C 2S', 'combo', 8),
		bad('3H 3D 3C 3S', 'comboTooHigh'),
		bad('2H 2D 2C 2S 2H', 'tooManyCards'),
		// Animal Companion pairs
		ok('AH 7D', 'companion', 8),
		ok('AH AS', 'companion', 2),
		ok('AC KD', 'companion', 21),
		ok('AC QD', 'companion', 16),
		ok('AC 10D', 'companion', 11),
		bad('AH X', 'jesterNotAlone'),
		bad('AH 2D 2C', 'companionPair'),
		bad('AH AD AC', 'companionPair'),
		bad('AH AD AC AS', 'companionPair'),
		// royals, mixed ranks, jester with others
		bad('JH JS', 'royalCombo'),
		bad('KH KS', 'royalCombo'),
		bad('2H JS', 'royalCombo'),
		bad('2H 3D', 'mixedRanks'),
		bad('5H 5D 6C', 'mixedRanks'),
		bad('X 5H', 'jesterNotAlone'),
		bad('X X', 'jesterNotAlone'),
		bad('X AH', 'jesterNotAlone'),
		bad('X KH', 'jesterNotAlone')
	])('$s', ({ s, want }) => {
		expect(checkPlay(cards(s))).toEqual(want);
	});

	it('checks in the documented order (jester before ace before royal)', () => {
		expect(checkPlay(cards('X AH JH'))).toEqual({ ok: false, problem: 'jesterNotAlone' });
		expect(checkPlay(cards('AH JH KH'))).toEqual({ ok: false, problem: 'companionPair' });
		expect(checkPlay(cards('2H JH 3C'))).toEqual({ ok: false, problem: 'royalCombo' });
		expect(checkPlay(cards('2H 3H 3C 2C 2D'))).toEqual({ ok: false, problem: 'mixedRanks' });
	});
});

describe('checkDiscard (rules §4 App decision)', () => {
	it.each([
		['10H', 10, { ok: true, sum: 10 }],
		['KS', 15, { ok: true, sum: 20 }],
		['QS', 15, { ok: true, sum: 15 }],
		['JS', 15, { ok: false, sum: 10, problem: 'notEnough' }],
		['JS 6H', 15, { ok: true, sum: 16 }],
		['JS 5H', 15, { ok: true, sum: 15 }],
		['7H 8D', 15, { ok: true, sum: 15 }],
		['7H 9D', 15, { ok: true, sum: 16 }],
		['KS 2H', 15, { ok: false, sum: 22, problem: 'superfluous' }],
		['KS QS', 15, { ok: false, sum: 35, problem: 'superfluous' }],
		['5H 5D 5C', 12, { ok: true, sum: 15 }],
		['5H 5D 5C', 10, { ok: false, sum: 15, problem: 'superfluous' }],
		['AH', 3, { ok: false, sum: 1, problem: 'notEnough' }],
		['', 5, { ok: false, sum: 0, problem: 'empty' }],
		// Jesters (value 0) are never part of a legal set
		['X 10H', 10, { ok: false, sum: 10, problem: 'superfluous' }],
		['X 9H', 10, { ok: false, sum: 9, problem: 'notEnough' }],
		['X', 10, { ok: false, sum: 0, problem: 'notEnough' }],
		['X 6H 5D', 10, { ok: false, sum: 11, problem: 'superfluous' }],
		// the smallest card is the one that must not be superfluous
		['3H 8D', 10, { ok: true, sum: 11 }],
		['2H 3D 8C', 10, { ok: false, sum: 13, problem: 'superfluous' }]
	])('%s against %i', (s, required, want) => {
		expect(checkDiscard(cards(s), required)).toEqual(want);
	});
});

/** View where both players have `handCount` cards; I am p1. Enemy is given. */
function viewWith(enemy: Enemy | null, o: Partial<PlayerView> = {}, handCount = 3): PlayerView {
	return makeView({
		players: [
			player('p1', 'Anna', 0, { isYou: true, isActive: true, handCount }),
			player('p2', 'Ben', 1, { handCount })
		],
		enemy,
		tavernCount: 30,
		discardCount: 10,
		...o
	});
}

describe('previewPlay', () => {
	it('returns null without an enemy', () => {
		expect(previewPlay(viewWith(null), cards('5H'))).toBeNull();
	});

	it('reports illegal plays', () => {
		const p = previewPlay(viewWith(enemyOf('JH')), cards('2H 3D'))!;
		expect(p.check).toEqual({ ok: false, problem: 'mixedRanks' });
		expect(p.defeats).toBe(false);
	});

	it('example: 3♦ + 3♠ + 3♣ draws 9, shields 9, deals 18', () => {
		const p = previewPlay(viewWith(enemyOf('JH')), cards('3D 3S 3C'))!;
		expect(p.check).toEqual({ ok: true, kind: 'combo', value: 9 });
		expect(p.draw).toBe(9);
		expect(p.shieldAfter).toBe(9);
		expect(p.damage).toBe(18);
		expect(p.damageAfter).toBe(18);
		expect(p.defeats).toBe(false);
		expect(p.attackAfter).toBe(1);
		expect(p.effects.map((e) => [e.suit, e.blocked, e.amount])).toEqual([
			['diamonds', false, 9],
			['clubs', false, 9],
			['spades', false, 9]
		]);
	});

	it('example: 8♦ + A♣ draws 9 and deals 18', () => {
		const p = previewPlay(viewWith(enemyOf('JH')), cards('8D AC'))!;
		expect(p.check).toEqual({ ok: true, kind: 'companion', value: 9 });
		expect(p.draw).toBe(9);
		expect(p.damage).toBe(18);
	});

	it('same suit twice counts once at the total value', () => {
		const p = previewPlay(viewWith(enemyOf('JD')), cards('AC 8C'))!;
		expect(p.value).toBe(9);
		expect(p.damage).toBe(18);
		expect(p.effects).toHaveLength(1);
		const h = previewPlay(viewWith(enemyOf('JD')), cards('5H AH'))!;
		expect(h.heal).toBe(6);
		expect(h.effects).toHaveLength(1);
	});

	it('hearts heal up to the discard size', () => {
		expect(previewPlay(viewWith(enemyOf('JD')), cards('5H'))!.heal).toBe(5);
		const small = viewWith(enemyOf('JD'), { discardCount: 2 });
		const p = previewPlay(small, cards('5H'))!;
		expect(p.heal).toBe(2);
		expect(p.effects[0]).toEqual({ suit: 'hearts', blocked: false, amount: 2 });
		expect(p.damage).toBe(5);
	});

	it('heal-then-draw: healed cards can be drawn when the Tavern is short', () => {
		const v = viewWith(enemyOf('JC'), { tavernCount: 0, discardCount: 6 });
		const p = previewPlay(v, cards('5H 5D'))!;
		expect(p.heal).toBe(6);
		expect(p.draw).toBe(6); // min(10, 0 + 6, free slots)
	});

	it('draw is capped by the Tavern', () => {
		const v = viewWith(enemyOf('JC'), { tavernCount: 3, discardCount: 0 });
		expect(previewPlay(v, cards('8D'))!.draw).toBe(3);
	});

	it('draw is capped by free slots, counting the played cards as removed', () => {
		// max 7; me 5 cards (play 2 -> 3, 4 free), Ben 6 cards (1 free): 5 free in total
		const v = viewWith(enemyOf('JC'), {
			players: [
				player('p1', 'Anna', 0, { isYou: true, isActive: true, handCount: 5 }),
				player('p2', 'Ben', 1, { handCount: 6 })
			]
		});
		const p = previewPlay(v, cards('5D 5H'))!;
		expect(p.draw).toBe(5);
		const full = viewWith(enemyOf('JC'), {
			players: [
				player('p1', 'Anna', 0, { isYou: true, isActive: true, handCount: 7 }),
				player('p2', 'Ben', 1, { handCount: 7 })
			]
		});
		expect(previewPlay(full, cards('4D'))!.draw).toBe(1);
	});

	describe('immunity', () => {
		it('J♦ is immune to diamonds: no draw, value still damages', () => {
			const p = previewPlay(viewWith(enemyOf('JD')), cards('5D'))!;
			expect(p.draw).toBe(0);
			expect(p.damage).toBe(5);
			expect(p.effects).toEqual([{ suit: 'diamonds', blocked: true, amount: 0 }]);
		});

		it('J♥ is immune to hearts: no heal', () => {
			const p = previewPlay(viewWith(enemyOf('JH')), cards('5H'))!;
			expect(p.heal).toBe(0);
			expect(p.effects[0].blocked).toBe(true);
			expect(p.damage).toBe(5);
		});

		it('J♣ is immune to clubs: not doubled', () => {
			const p = previewPlay(viewWith(enemyOf('JC')), cards('5C'))!;
			expect(p.damage).toBe(5);
			expect(p.effects[0]).toEqual({ suit: 'clubs', blocked: true, amount: 0 });
		});

		it('J♠ is immune to spades: no shield', () => {
			const p = previewPlay(viewWith(enemyOf('JS')), cards('5S'))!;
			expect(p.shieldAfter).toBe(0);
			expect(p.attackAfter).toBe(10);
			expect(p.effects[0].blocked).toBe(true);
		});

		it('only the matching suit is blocked in a mixed play', () => {
			const p = previewPlay(viewWith(enemyOf('JC')), cards('3C 3S'))!;
			expect(p.effects.map((e) => [e.suit, e.blocked])).toEqual([
				['clubs', true],
				['spades', false]
			]);
			expect(p.damage).toBe(6);
			expect(p.shieldAfter).toBe(6);
		});

		it('after a Jester the enemy suit works again', () => {
			const e = enemyOf('JS', { plays: [{ by: 'p2', cards: 'X' }] });
			const p = previewPlay(viewWith(e), cards('5S'))!;
			expect(p.shieldAfter).toBe(5);
			expect(p.attackAfter).toBe(5);
			const c = enemyOf('JC', { plays: [{ by: 'p2', cards: 'X' }] });
			expect(previewPlay(viewWith(c), cards('5C'))!.damage).toBe(10);
		});

		it('a Jester play cancels immunity and does no damage', () => {
			const p = previewPlay(viewWith(enemyOf('JS')), cards('X'))!;
			expect(p.check).toEqual({ ok: true, kind: 'jester', value: 0 });
			expect(p.cancelsImmunity).toBe(true);
			expect(p.damage).toBe(0);
			expect(p.effects).toEqual([]);
			expect(p.defeats).toBe(false);
		});

		it('a Jester against a ♠ enemy makes earlier spades count (retroactive shield)', () => {
			const e = enemyOf('JS', {
				plays: [
					{ by: 'p1', cards: '4S', blocked: ['spades'] },
					{ by: 'p2', cards: '3H 3S', blocked: ['spades'] }
				]
			});
			expect(e.shield).toBe(0);
			const p = previewPlay(viewWith(e), cards('X'))!;
			expect(p.shieldAfter).toBe(10);
			expect(p.attackAfter).toBe(0);
		});

		it('a Jester does not change a non-spade enemy shield, nor re-double clubs', () => {
			const e = enemyOf('JC', { plays: [{ by: 'p1', cards: '4C', blocked: ['clubs'] }] });
			const p = previewPlay(viewWith(e), cards('X'))!;
			expect(p.shieldAfter).toBe(0);
			expect(p.damageAfter).toBe(4); // the earlier 4♣ stays undoubled
		});

		it('a second Jester no longer cancels anything', () => {
			const e = enemyOf('JS', { plays: [{ by: 'p2', cards: 'X' }] });
			expect(previewPlay(viewWith(e), cards('X'))!.cancelsImmunity).toBe(false);
		});

		it('shield from earlier plays carries into the preview', () => {
			const p = previewPlay(playView, cards('9S'))!; // J♠ cancelled, shield 4 already
			expect(p.shieldAfter).toBe(13);
			expect(p.attackAfter).toBe(0);
			expect(p.damageAfter).toBe(13);
		});
	});

	describe('defeat', () => {
		it('exact damage defeats and flags exact', () => {
			const p = previewPlay(viewWith(enemyOf('JH')), cards('10C'))!; // 20 = health
			expect(p.defeats).toBe(true);
			expect(p.exact).toBe(true);
		});

		it('overkill defeats without exact', () => {
			const p = previewPlay(viewWith(enemyOf('JH', { damage: 15 })), cards('KC'))!;
			expect(p.damageAfter).toBe(55);
			expect(p.defeats).toBe(true);
			expect(p.exact).toBe(false);
		});

		it('accumulated damage counts', () => {
			const e = enemyOf('QS', { damage: 25 });
			expect(previewPlay(viewWith(e), cards('4H'))!.defeats).toBe(false);
			expect(previewPlay(viewWith(e), cards('5H'))!.exact).toBe(true);
		});

		it('doubled clubs can defeat', () => {
			const e = enemyOf('JH', { damage: 8 });
			const p = previewPlay(viewWith(e), cards('6C'))!;
			expect(p.damageAfter).toBe(20);
			expect(p.exact).toBe(true);
		});

		it('computes the remaining attack from the shield', () => {
			const p = previewPlay(viewWith(enemyOf('KH')), cards('9S'))!;
			expect(p.attackAfter).toBe(11);
			expect(p.defeats).toBe(false);
		});

		it('never goes below zero attack', () => {
			const p = previewPlay(viewWith(enemyOf('JH')), cards('KS'))!;
			expect(p.attackAfter).toBe(0);
		});
	});
});

describe('fixtures sanity', () => {
	it('sorts hands', () => {
		expect(hand('KD 2H AS').map(cardKey)).toEqual(['spades-1', 'hearts-2', 'diamonds-13']);
	});

	it('rejects bad cards', () => {
		expect(() => card('ZZ')).toThrow();
	});
});
