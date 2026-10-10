import { describe, expect, it } from 'vitest';
import { plural, resultHeadline, stuckLine } from './format.js';
import { finishedLost, finishedWon, playingMyTurn } from './fixtures.js';
import type { Result } from '#lib/types.js';

const result = (score: number): Result => ({
	outcome: score === 0 ? 'won' : 'lost',
	score,
	excellent: score < 10,
	cardsLeftInHands: score,
	cardsLeftInDrawPile: 0
});

describe('resultHeadline', () => {
	it.each([
		[0, 'You won! All 98 cards played.'],
		[1, 'Excellent! Only 1 card left.'],
		[9, 'Excellent! Only 9 cards left.'],
		[10, 'Game over: 10 cards left.'],
		[57, 'Game over: 57 cards left.']
	])('score %i', (score, want) => {
		expect(resultHeadline(result(score))).toBe(want);
	});
});

describe('stuckLine', () => {
	it('names the stuck player', () => {
		expect(stuckLine(finishedLost)).toBe('Ben had no playable card.');
	});
	it('says "You" when the viewer got stuck', () => {
		const v = structuredClone(finishedLost);
		v.result!.stuckPlayerId = v.you;
		expect(stuckLine(v)).toBe('You had no playable card.');
	});
	it('is null without a stuck player', () => {
		expect(stuckLine(finishedWon)).toBeNull();
		expect(stuckLine(playingMyTurn)).toBeNull();
	});
	it('is null for an unknown player id', () => {
		const v = structuredClone(finishedLost);
		v.result!.stuckPlayerId = 'p_nobody';
		expect(stuckLine(v)).toBeNull();
	});
});

describe('plural', () => {
	it('handles one and many', () => {
		expect(plural(1, 'player')).toBe('1 player');
		expect(plural(3, 'player')).toBe('3 players');
	});
});
