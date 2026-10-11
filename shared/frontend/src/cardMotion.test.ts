import { describe, expect, it } from 'vitest';
import { createCardMotion, flightDuration, offstageOffset } from './cardMotion.js';

describe('offstageOffset', () => {
	it.each([
		['top', { x: 0, y: -100 }],
		['bottom', { x: 0, y: 100 }],
		['left', { x: -100, y: 0 }],
		['right', { x: 100, y: 0 }],
		['none', { x: 0, y: 0 }]
	] as const)('%s', (offstage, want) => {
		expect(offstageOffset(offstage, 100)).toEqual(want);
	});
});

describe('flightDuration', () => {
	it('grows with the distance within bounds', () => {
		expect(flightDuration(0, 450)).toBe(180);
		expect(flightDuration(100, 450)).toBe(280);
		expect(flightDuration(10_000, 450)).toBe(450);
	});
});

describe('createCardMotion', () => {
	it('staggers a deal', () => {
		const m = createCardMotion({ stagger: 50 });
		expect([0, 1, 4].map((i) => m.dealDelay(i))).toEqual([0, 50, 200]);
		expect(m.flip.duration).toBe(250);
	});
});
