import { describe, expect, it } from 'vitest';
import {
	cardKey,
	cardLabel,
	displayIndex,
	endLabel,
	lossText,
	moveLabel,
	resultHeadline,
	sameCard,
	towardMe,
	turnStatus
} from './format.js';
import { card, playingView } from './fixtures.js';
import type { MoveOption } from '#lib/types.js';

const opt = (o: Partial<MoveOption>): MoveOption => ({
	delta: 0,
	to: 5,
	beyond: '',
	gem: false,
	lost: false,
	...o
});

describe('board orientation', () => {
	it('keeps the left end on the left for seat 0 and mirrors for seat 1', () => {
		expect(displayIndex(0, 11, 'left')).toBe(0);
		expect(displayIndex(0, 11, 'right')).toBe(10);
		expect(displayIndex(4, 10, 'right')).toBe(5);
	});
	it('counts moves toward the viewer', () => {
		expect(towardMe(-3, 'left')).toBe(3);
		expect(towardMe(-3, 'right')).toBe(-3);
		expect(towardMe(2, 'right')).toBe(2);
	});
});

describe('cards', () => {
	it('labels and keys', () => {
		expect(cardLabel(card('roses', 7))).toBe('7 Rose');
		expect(cardKey(card('stars', 10))).toBe('stars-10');
		expect(sameCard(card('doves', 1), { suit: 'doves', rank: 1 })).toBe(true);
		expect(sameCard(card('doves', 1), null)).toBe(false);
	});
});

describe('moveLabel', () => {
	const view = playingView(); // you (Alice) on the left, Bob on the right
	it.each([
		[opt({ delta: -3, gem: true }), '3 spaces toward you', 'collect a gem'],
		[opt({ delta: 1 }), '1 space toward Bob', 'no gem'],
		[opt({ delta: 0 }), 'Stay', 'no gem'],
		[
			opt({ delta: 4, beyond: 'right' }),
			'4 spaces toward Bob',
			'off the path: forest token, back to start'
		],
		[
			opt({ delta: 4, beyond: 'right', lost: true }),
			'4 spaces toward Bob',
			'off the path: lost in the forest!'
		]
	])('%o', (o, title, result) => {
		expect(moveLabel(view, o)).toEqual({ title, result });
	});
});

describe('texts', () => {
	it('names the ends from the viewer side', () => {
		const view = playingView();
		expect(endLabel(view, 'left')).toBe('your end');
		expect(endLabel(view, 'right')).toBe("Bob's end");
	});

	it('describes the pending step', () => {
		const view = playingView();
		expect(turnStatus(view)).toEqual({
			mine: true,
			title: 'Your turn',
			hint: 'Lead any card',
			waiting: ''
		});
		view.pending = { ...view.pending!, step: 'follow', playerIds: ['bob'] };
		expect(turnStatus(view)).toEqual({
			mine: false,
			title: "Bob's turn",
			hint: '',
			waiting: 'Waiting for Bob to follow.'
		});
		view.pending = { ...view.pending!, step: 'follow', playerIds: ['alice'] };
		view.trick = { number: 1, leaderId: 'bob', lead: card('stars', 4), follow: null };
		expect(turnStatus(view)?.hint).toBe('Follow Stars if you can');
		view.trick.lead = card('stars', 9);
		expect(turnStatus(view)?.hint).toBe('Royal Heir led: play any card');
		view.trick.follow = card('stars', 1);
		view.pending = { ...view.pending!, step: 'move' };
		expect(turnStatus(view)?.hint).toBe('Musician: pick the way');
		view.pending = {
			...view.pending!,
			step: 'gift',
			playerIds: ['bob'],
			yourGift: card('roses', 2)
		};
		expect(turnStatus(view)?.waiting).toBe("Waiting for Bob's gift.");
	});

	it('explains results', () => {
		const view = playingView();
		view.result = {
			outcome: 'lost',
			reason: 'gaveUp',
			playerId: 'bob',
			round: 2,
			gemsLeft: 4,
			score: null
		};
		expect(lossText(view)).toBe('Bob gave up.');
		expect(resultHeadline(view.result)).toBe('Defeat: gave up');
		view.result = { outcome: 'lost', reason: 'outOfTime', round: 3, gemsLeft: 1, score: null };
		expect(lossText(view)).toBe('Round 3 ended with 1 gem still in the forest.');
		expect(
			resultHeadline({
				outcome: 'won',
				round: 2,
				gemsLeft: 0,
				score: { level: 10, cards: 4, earlyWin: 10, forestTokens: 6, total: 30 }
			})
		).toBe('Victory! 30 points');
	});
});
