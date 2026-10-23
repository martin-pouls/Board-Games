import { describe, expect, it } from 'vitest';
import type { Phase } from '#lib/types.js';
import { fixtures } from './fixtures.js';

// The dev route /dev/fixtures/<name> shows each fixture on the game screen (screenshots, visual
// review), so every phase with a screen needs at least one. A new phase fails to compile here
// until it is listed; `false` marks a phase whose screen is not built yet (its work package adds
// the fixtures and flips the flag).
const phases: Record<Phase, boolean> = {
	lobby: true,
	playing: false,
	won: false,
	lost: false,
	ended: true
};

describe('fixtures', () => {
	it('cover every phase that has a screen', () => {
		const covered = new Set(Object.values(fixtures).map((make) => make().phase));
		const expected = Object.entries(phases)
			.filter(([, built]) => built)
			.map(([phase]) => phase);
		expect([...covered].sort()).toEqual(expected.sort());
	});

	it('return a fresh view on every call', () => {
		for (const [name, make] of Object.entries(fixtures)) {
			const first = make();
			expect(make(), name).toEqual(first);
			expect(make(), name).not.toBe(first);
		}
	});

	it('have exactly one viewer, who is a player', () => {
		for (const [name, make] of Object.entries(fixtures)) {
			const v = make();
			expect(
				v.players.filter((p) => p.isYou).map((p) => p.id),
				name
			).toEqual([v.you]);
		}
	});
});
