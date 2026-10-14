import { describe, expect, it } from 'vitest';
import { HAND_ORDER_KEY, loadHandOrder, saveHandOrder } from './handOrder.js';

function memory() {
	const m = new Map<string, string>();
	return {
		getItem: (k: string) => m.get(k) ?? null,
		setItem: (k: string, v: string) => void m.set(k, v)
	};
}

const blocked = () => {
	throw new Error('SecurityError');
};

describe('hand order storage', () => {
	it('round-trips an order', () => {
		const s = memory();
		saveHandOrder('suitAceHigh', () => s);
		expect(s.getItem(HAND_ORDER_KEY)).toBe('suitAceHigh');
		expect(loadHandOrder(() => s)).toBe('suitAceHigh');
	});

	it('falls back to the default', () => {
		const s = memory();
		expect(loadHandOrder(() => s)).toBe('rank');
		s.setItem(HAND_ORDER_KEY, 'bogus');
		expect(loadHandOrder(() => s)).toBe('rank');
		expect(loadHandOrder(() => undefined)).toBe('rank');
	});

	it('survives blocked storage', () => {
		expect(loadHandOrder(blocked)).toBe('rank');
		expect(() => saveHandOrder('suit', blocked)).not.toThrow();
	});
});
