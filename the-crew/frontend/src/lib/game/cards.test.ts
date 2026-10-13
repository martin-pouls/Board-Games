import { describe, expect, it } from 'vitest';
import { byTokenOrder } from './cards.js';
import type { Token } from '#lib/types.js';

const tasks = (...tokens: Token[]) => tokens.map((token, id) => ({ id, token }));
const ids = (ts: { id: number }[]) => ts.map((t) => t.id);

describe('byTokenOrder', () => {
	it('puts numbered tokens in order', () => {
		expect(ids(byTokenOrder(tasks('3', '1', '2')))).toEqual([1, 2, 0]);
	});
	it('puts arrows in order, after numbers', () => {
		expect(ids(byTokenOrder(tasks('arrow3', 'arrow1', '1', 'arrow2')))).toEqual([2, 1, 3, 0]);
	});
	it('puts tasks without a token before omega, keeping their order', () => {
		expect(ids(byTokenOrder(tasks('omega', '', '2', '', '1')))).toEqual([4, 2, 1, 3, 0]);
	});
	it('follows a swap of two tokens', () => {
		expect(ids(byTokenOrder(tasks('1', '3', '2')))).toEqual([0, 2, 1]);
	});
	it('does not change the input', () => {
		const ts = tasks('2', '1');
		byTokenOrder(ts);
		expect(ids(ts)).toEqual([0, 1]);
	});
});
