import { describe, expect, it } from 'vitest';
import { endedLine, plural, resultHeadline, stuckLine } from './format.js';
import { finishedLost, finishedLostSkull, finishedWon, playingMyTurn } from './fixtures.js';
import type { PlayerView, Result } from '#lib/types.js';

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

it('says the game ended early when aborted', () => {
	expect(resultHeadline({ ...result(3), outcome: 'aborted' })).toBe(
		'Game ended early: 3 cards left.'
	);
});

it('checks the outcome before the score: lost with 0 cards left is not a win', () => {
	expect(resultHeadline({ ...result(0), outcome: 'lost' })).toBe(
		'So close! All cards played, but a skull stayed uncovered.'
	);
	expect(resultHeadline(finishedLostSkull.result!)).toBe(
		'So close! All cards played, but a skull stayed uncovered.'
	);
	expect(resultHeadline(finishedWon.result!)).toBe('You won! All 98 cards played.');
});

describe('endedLine', () => {
	const aborted = (): PlayerView => {
		const v = structuredClone(finishedLost);
		v.result = { ...v.result!, outcome: 'aborted', stuckPlayerId: undefined };
		return v;
	};
	it('names the host', () => {
		expect(endedLine(aborted())).toBe('Anna ended the game.');
	});
	it('says "You" when the viewer is the host', () => {
		const v = aborted();
		v.hostId = v.you;
		expect(endedLine(v)).toBe('You ended the game.');
	});
	it('is null unless aborted', () => {
		expect(endedLine(finishedLost)).toBeNull();
		expect(endedLine(finishedWon)).toBeNull();
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
	it('names an uncovered skull', () => {
		expect(stuckLine(finishedLostSkull)).toBe("Dan couldn't cover the skull.");
		const v = structuredClone(finishedLostSkull);
		v.result!.stuckPlayerId = v.you;
		expect(stuckLine(v)).toBe("You couldn't cover the skull.");
	});
	it('names an unfinished 3!', () => {
		const v = structuredClone(finishedLostSkull);
		v.turn!.constraints = { ...v.turn!.constraints, skullPiles: [], exactCount: 3 };
		expect(stuckLine(v)).toBe("Dan couldn't complete 3!.");
		v.result!.stuckPlayerId = v.you;
		expect(stuckLine(v)).toBe("You couldn't complete 3!.");
	});
	it('falls back to "no playable card" when 3! is complete or absent', () => {
		const v = structuredClone(finishedLostSkull);
		v.turn!.constraints = { ...v.turn!.constraints, skullPiles: [], exactCount: 3 };
		v.turn!.cardsPlayed = 3;
		expect(stuckLine(v)).toBe('Dan had no playable card.');
		v.turn!.constraints.exactCount = null;
		v.turn!.cardsPlayed = 0;
		expect(stuckLine(v)).toBe('Dan had no playable card.');
	});
});

describe('plural', () => {
	it('handles one and many', () => {
		expect(plural(1, 'player')).toBe('1 player');
		expect(plural(3, 'player')).toBe('3 players');
	});
});
