import { describe, expect, it } from 'vitest';
import { canPlayerYield } from './rules.js';
import { discardView, hand, makeView, playView } from './fixtures.js';
import { pickCards, selectionKey, toggleIndex } from './selection.js';

describe('selectionKey', () => {
	it('is stable for the same situation', () => {
		expect(selectionKey(playView, true)).toBe(selectionKey({ ...playView, revision: 99 }, true));
	});
	it('changes with the hand, the step and the turn', () => {
		const base = selectionKey(playView, true);
		expect(selectionKey({ ...playView, hand: hand('2H 3H') }, true)).not.toBe(base);
		expect(selectionKey(discardView, true)).not.toBe(base);
		expect(selectionKey({ ...playView, turn: { ...playView.turn!, number: 5 } }, true)).not.toBe(
			base
		);
	});
	it('collapses to one key while not interactive', () => {
		expect(selectionKey(playView, false)).toBe(selectionKey(discardView, false));
		expect(selectionKey(playView, false)).not.toBe(selectionKey(playView, true));
	});
	it('copes with a missing turn', () => {
		expect(() => selectionKey(makeView(), true)).not.toThrow();
	});
});

describe('toggleIndex / pickCards', () => {
	it('adds sorted and removes', () => {
		expect(toggleIndex([1, 4], 2)).toEqual([1, 2, 4]);
		expect(toggleIndex([1, 2, 4], 2)).toEqual([1, 4]);
	});
	it('picks cards by index and skips stale ones', () => {
		const h = hand('2H 5D 9S');
		expect(pickCards(h, [0, 2])).toEqual([h[0], h[2]]);
		expect(pickCards(h, [1, 7])).toEqual([h[1]]);
	});
});

describe('canPlayerYield', () => {
	const ps = (...flags: boolean[]) => flags.map((y, i) => ({ id: `p${i}`, yieldedLastTurn: y }));
	it('is always true solo', () => {
		expect(canPlayerYield(ps(true), 'p0')).toBe(true);
	});
	it('needs some other player who did not yield', () => {
		expect(canPlayerYield(ps(false, true), 'p0')).toBe(false);
		expect(canPlayerYield(ps(false, false), 'p0')).toBe(true);
		expect(canPlayerYield(ps(true, true, false), 'p0')).toBe(true);
		expect(canPlayerYield(ps(true, true, true), 'p0')).toBe(false);
	});
});
