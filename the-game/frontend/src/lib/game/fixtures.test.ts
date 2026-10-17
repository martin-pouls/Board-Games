import { describe, expect, it } from 'vitest';
import type { Phase } from '#lib/types.js';
import { fixtures } from './fixtures.js';

// The dev route /dev/fixtures/<name> shows each fixture on the game screen (screenshots, visual
// review), so every phase needs at least one. A new phase fails to compile here until it is listed.
const phases: Record<Phase, true> = {
	lobby: true,
	choosingFirstPlayer: true,
	playing: true,
	won: true,
	lost: true,
	ended: true
};

describe('fixtures', () => {
	it('cover every phase', () => {
		const covered = new Set(Object.values(fixtures).map((make) => make().phase));
		expect([...covered].sort()).toEqual(Object.keys(phases).sort());
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
