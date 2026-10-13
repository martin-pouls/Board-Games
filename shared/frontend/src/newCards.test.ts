import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { NewCards } from './newCards.svelte.js';

const marked = (n: NewCards, keys: string[]) => keys.filter((k) => n.has(k));

describe('NewCards', () => {
	let n: NewCards;
	beforeEach(() => {
		vi.useFakeTimers();
		n = new NewCards(1000);
	});
	afterEach(() => {
		n.dispose();
		vi.useRealTimers();
	});

	it('marks nothing on the first hand', () => {
		n.update(['a', 'b', 'c']);
		expect(marked(n, ['a', 'b', 'c'])).toEqual([]);
	});

	it('marks drawn cards until the time is up', () => {
		n.update(['a', 'b']);
		n.update(['a', 'b', 'c', 'd']);
		expect(marked(n, ['a', 'b', 'c', 'd'])).toEqual(['c', 'd']);
		vi.advanceTimersByTime(999);
		expect(n.has('c')).toBe(true);
		vi.advanceTimersByTime(1);
		expect(marked(n, ['a', 'b', 'c', 'd'])).toEqual([]);
	});

	it('keeps only the latest draw marked, timed from that draw', () => {
		n.update(['a']);
		n.update(['a', 'b']);
		vi.advanceTimersByTime(600);
		n.update(['a', 'b', 'c']);
		expect(marked(n, ['a', 'b', 'c'])).toEqual(['c']);
		vi.advanceTimersByTime(999);
		expect(n.has('c')).toBe(true);
		vi.advanceTimersByTime(1);
		expect(n.has('c')).toBe(false);
	});

	it('keeps the marks while cards only leave or come back', () => {
		n.update(['a', 'b']);
		n.update(['a', 'b', 'c', 'd']);
		n.update(['b', 'c', 'd']);
		n.update(['a', 'b', 'c', 'd']);
		expect(marked(n, ['a', 'b', 'c', 'd'])).toEqual(['c', 'd']);
	});

	it('marks cards drawn after playing in the same update', () => {
		n.update(['a', 'b', 'c']);
		n.update(['b', 'c', 'd']);
		expect(marked(n, ['b', 'c', 'd'])).toEqual(['d']);
	});

	it('drops the mark of a card that leaves the hand', () => {
		n.update(['a']);
		n.update(['a', 'b']);
		n.update(['a']);
		expect(n.has('b')).toBe(false);
	});

	it('does not mark a card that comes back (an undone play)', () => {
		n.update(['a', 'b', 'c']);
		n.update(['a', 'c']);
		n.update(['a']);
		n.update(['a', 'c']);
		n.update(['a', 'b', 'c']);
		expect(marked(n, ['a', 'b', 'c'])).toEqual([]);
	});

	it('marks a card again once others were drawn in between', () => {
		n.update(['a', 'b']);
		n.update(['a']);
		n.update(['a', 'c']);
		n.update(['a', 'b', 'c']);
		expect(n.has('b')).toBe(true);
	});

	it('marks nothing when the whole hand is replaced', () => {
		n.update(['a', 'b']);
		n.update(['c', 'd', 'e']);
		expect(marked(n, ['c', 'd', 'e'])).toEqual([]);
	});

	it('marks cards drawn into an empty hand', () => {
		n.update(['a']);
		n.update([]);
		n.update(['b', 'c']);
		expect(marked(n, ['b', 'c'])).toEqual(['b', 'c']);
	});

	it('starts over without marks when the deal changes', () => {
		n.update(['a'], 1);
		n.update(['a', 'b'], 1);
		n.update(['a', 'b', 'c'], 2);
		expect(marked(n, ['a', 'b', 'c'])).toEqual([]);
		n.update(['a', 'b', 'c', 'd'], 2);
		expect(n.has('d')).toBe(true);
	});

	it('counts duplicate keys once', () => {
		n.update(['a', 'a']);
		n.update(['a', 'a', 'b']);
		expect(marked(n, ['a', 'b'])).toEqual(['b']);
	});
});
