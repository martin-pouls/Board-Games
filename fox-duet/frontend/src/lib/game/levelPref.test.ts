import { describe, expect, it } from 'vitest';
import { parseLevel } from './levelPref.js';

describe('parseLevel', () => {
	it('keeps the three levels', () => {
		expect([1, 2, 3].map(parseLevel)).toEqual([1, 2, 3]);
	});

	it('rejects anything else', () => {
		for (const v of [0, 4, '2', null, 1.5]) expect(parseLevel(v)).toBeNull();
	});
});
