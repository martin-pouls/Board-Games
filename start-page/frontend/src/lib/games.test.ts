import { describe, expect, it } from 'vitest';
import { games } from './games.js';

describe('games', () => {
	it('links each game to its own sub-path', () => {
		for (const game of games) expect(game.href).toBe(`/${game.id}/`);
	});

	it('has unique ids', () => {
		expect(new Set(games.map((g) => g.id)).size).toBe(games.length);
	});
});
