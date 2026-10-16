import { describe, expect, it } from 'vitest';
import { lossLine, plural, resultHeadline } from './format.js';
import { givenUp, lostGame, lostSkull, playingMyTurn, wonGame } from './fixtures.js';
import type { PlayerView } from '#lib/types.js';

const lost = (score: number): Pick<PlayerView, 'phase' | 'result'> => ({
	phase: 'lost',
	result: {
		score,
		excellent: score < 10,
		cardsLeftInHands: score,
		cardsLeftInDrawPile: 0,
		reason: 'stuck',
		playerId: 'p_ben'
	}
});

describe('resultHeadline', () => {
	it('celebrates a win', () => {
		expect(resultHeadline(wonGame)).toBe('You won! All 98 cards played.');
	});
	it.each([
		[1, 'Excellent! Only 1 card left.'],
		[9, 'Excellent! Only 9 cards left.'],
		[10, 'Game over: 10 cards left.'],
		[57, 'Game over: 57 cards left.']
	])('lost with score %i', (score, want) => {
		expect(resultHeadline(lost(score))).toBe(want);
	});
	it('checks the phase before the score: lost with 0 cards left is not a win', () => {
		expect(resultHeadline(lost(0))).toBe(
			'So close! All cards played, but a skull stayed uncovered.'
		);
		expect(resultHeadline(lostSkull)).toBe(
			'So close! All cards played, but a skull stayed uncovered.'
		);
	});
	it('scores a given-up game like any loss', () => {
		expect(resultHeadline(givenUp)).toBe('Game over: 17 cards left.');
		const v = { ...givenUp, result: { ...givenUp.result!, score: 4, excellent: true } };
		expect(resultHeadline(v)).toBe('Excellent! Only 4 cards left.');
	});
});

describe('lossLine', () => {
	it('names the stuck player', () => {
		expect(lossLine(lostGame)).toBe('Ben had no playable card.');
	});
	it('says "You" when the viewer got stuck', () => {
		const v = structuredClone(lostGame);
		v.result!.playerId = v.you;
		expect(lossLine(v)).toBe('You had no playable card.');
	});
	it('names who gave up, whoever was on turn', () => {
		expect(lossLine(givenUp)).toBe('Anna gave up.');
		expect(lossLine(lostSkull)).toBe('Dan gave up.');
		const v = structuredClone(givenUp);
		v.result!.playerId = v.you;
		expect(lossLine(v)).toBe('You gave up.');
	});
	it('still explains a give-up without a known player', () => {
		const v = structuredClone(givenUp);
		v.result!.playerId = undefined;
		expect(lossLine(v)).toBe('The game was given up.');
	});
	it('is null when won, still running, or the stuck player is unknown', () => {
		expect(lossLine(wonGame)).toBeNull();
		expect(lossLine(playingMyTurn)).toBeNull();
		const v = structuredClone(lostGame);
		v.result!.playerId = 'p_nobody';
		expect(lossLine(v)).toBeNull();
	});
});

describe('plural', () => {
	it('handles one and many', () => {
		expect(plural(1, 'player')).toBe('1 player');
		expect(plural(3, 'player')).toBe('3 players');
	});
});
