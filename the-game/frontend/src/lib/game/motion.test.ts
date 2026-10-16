import { describe, expect, it } from 'vitest';
import { cardKey, pileTopKey } from './motion.js';

describe('motion keys', () => {
	it('gives a hand card and the same card on a pile the same key', () => {
		expect(pileTopKey({ index: 2, topCard: { value: 42, kind: 'number' } })).toBe(cardKey(42));
	});

	it('gives each empty pile its own start key', () => {
		expect(pileTopKey({ index: 0, topCard: null })).toBe('start-0');
		expect(pileTopKey({ index: 3, topCard: null })).not.toBe(
			pileTopKey({ index: 0, topCard: null })
		);
	});

	it('never mixes card keys with start keys', () => {
		expect(cardKey(1)).not.toBe(pileTopKey({ index: 1, topCard: null }));
	});
});
