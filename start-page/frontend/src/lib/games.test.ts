import { describe, expect, it } from 'vitest';
import { gameHref, games } from './games.js';

describe('games', () => {
	it('links each game to its own sub-path', () => {
		for (const game of games) expect(game.path).toBe(`/${game.id}/`);
	});

	it('has unique ids and dev ports', () => {
		expect(new Set(games.map((g) => g.id)).size).toBe(games.length);
		expect(new Set(games.map((g) => g.devPort)).size).toBe(games.length);
	});
});

describe('gameHref', () => {
	const game = games.find((g) => g.id === 'regicide')!;

	it('uses the path on the same host in production', () => {
		expect(gameHref(game, new URL('https://yoshi-und-kruemli.de/board-games/'), false)).toBe(
			'/regicide/'
		);
	});

	it("uses the game's dev port on the same hostname in dev", () => {
		expect(gameHref(game, new URL('http://localhost:5175/'), true)).toBe('http://localhost:5174/');
		expect(gameHref(game, new URL('http://192.168.1.20:5175/'), true)).toBe(
			'http://192.168.1.20:5174/'
		);
	});
});
