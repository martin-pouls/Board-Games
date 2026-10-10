import { describe, expect, it, vi } from 'vitest';
import { ApiError } from '#lib/api.js';
import type { GameSummary, Outcome, Phase } from '#lib/types.js';
import {
	MAX_CONTINUE,
	fromSummary,
	isContinuable,
	loadRecentGames,
	parseCreatedAt,
	partitionGames,
	statusLabel,
	type RecentGame
} from './recent.js';

const game = (code: string, phase: Phase | null, createdAt: number, outcome?: Outcome) =>
	({ code, name: 'Anna', phase, outcome, createdAt }) satisfies RecentGame;

const summary = (over: Partial<GameSummary> = {}): GameSummary => ({
	gameId: 'K7MQ2',
	phase: 'lobby',
	hostName: 'Anna',
	playerCount: 1,
	maxPlayers: 5,
	joinable: true,
	...over
});

describe('isContinuable', () => {
	it.each([
		['lobby', true],
		['choosingFirstPlayer', true],
		['playing', true],
		['finished', false],
		[null, true]
	] as const)('%s -> %s', (phase, want) => {
		expect(isContinuable({ phase })).toBe(want);
	});
});

describe('partitionGames', () => {
	it('splits finished games into history and sorts both newest first', () => {
		const { continuable, history } = partitionGames([
			game('AAAAA', 'playing', 100),
			game('BBBBB', 'finished', 500, 'won'),
			game('CCCCC', 'lobby', 300),
			game('DDDDD', 'finished', 400, 'lost'),
			game('EEEEE', 'finished', 50, 'aborted')
		]);
		expect(continuable.map((g) => g.code)).toEqual(['CCCCC', 'AAAAA']);
		expect(history.map((g) => g.code)).toEqual(['BBBBB', 'DDDDD', 'EEEEE']);
	});

	it('keeps only the newest five continuable games; the rest are dropped, not moved to history', () => {
		const games = Array.from({ length: 8 }, (_, i) => game(`G${i}`.padEnd(5, '0'), 'playing', i));
		const { continuable, history } = partitionGames(games);
		expect(continuable).toHaveLength(MAX_CONTINUE);
		expect(continuable.map((g) => g.createdAt)).toEqual([7, 6, 5, 4, 3]);
		expect(history).toEqual([]);
	});

	it('sorts games without a date last and breaks ties by code', () => {
		const { continuable } = partitionGames([
			game('ZZZZZ', 'lobby', 0),
			game('BBBBB', 'lobby', 10),
			game('AAAAA', 'lobby', 0),
			game('CCCCC', 'lobby', 10)
		]);
		expect(continuable.map((g) => g.code)).toEqual(['BBBBB', 'CCCCC', 'AAAAA', 'ZZZZZ']);
	});

	it('does not mutate its input and handles empty lists', () => {
		const input = [game('AAAAA', 'lobby', 1), game('BBBBB', 'lobby', 2)];
		partitionGames(input);
		expect(input.map((g) => g.code)).toEqual(['AAAAA', 'BBBBB']);
		expect(partitionGames([])).toEqual({ continuable: [], history: [] });
	});
});

describe('parseCreatedAt / fromSummary', () => {
	it('parses RFC3339 and tolerates missing or broken values', () => {
		expect(parseCreatedAt('2026-10-10T12:30:05Z')).toBe(Date.UTC(2026, 9, 10, 12, 30, 5));
		expect(parseCreatedAt(undefined)).toBe(0);
		expect(parseCreatedAt('')).toBe(0);
		expect(parseCreatedAt('nope')).toBe(0);
	});

	it('maps a summary', () => {
		const g = fromSummary(
			'K7MQ2',
			'Anna',
			summary({ phase: 'finished', outcome: 'won', createdAt: '2026-10-10T12:30:05Z' })
		);
		expect(g).toEqual({
			code: 'K7MQ2',
			name: 'Anna',
			phase: 'finished',
			outcome: 'won',
			createdAt: Date.UTC(2026, 9, 10, 12, 30, 5)
		});
	});
});

describe('statusLabel', () => {
	it.each([
		[game('A', 'lobby', 0), 'Lobby'],
		[game('A', 'choosingFirstPlayer', 0), 'In progress'],
		[game('A', 'playing', 0), 'In progress'],
		[game('A', 'finished', 0, 'won'), 'Won'],
		[game('A', 'finished', 0, 'lost'), 'Lost'],
		[game('A', 'finished', 0, 'aborted'), 'Ended'],
		[game('A', 'finished', 0), 'Finished'],
		[game('A', null, 0), '']
	])('%o -> %s', (g, label) => {
		expect(statusLabel(g)).toBe(label);
	});
});

describe('loadRecentGames', () => {
	const stored = [
		{ code: 'AAAAA', name: 'Anna' },
		{ code: 'BBBBB', name: 'Ben' },
		{ code: 'CCCCC', name: 'Cy' },
		{ code: 'DDDDD', name: 'Di' }
	];

	it('drops games the server no longer has, keeps failed ones with an unknown phase', async () => {
		const fetchSummary = vi.fn(async (code: string): Promise<GameSummary> => {
			if (code === 'BBBBB') throw new ApiError(404, 'game_not_found', 'gone');
			if (code === 'CCCCC') throw new TypeError('network down');
			if (code === 'DDDDD') throw new ApiError(500, 'internal', 'boom');
			return summary({ gameId: code, phase: 'playing', createdAt: '2026-10-10T00:00:00Z' });
		});
		const gone: string[] = [];
		const games = await loadRecentGames(stored, fetchSummary, (c) => gone.push(c));

		expect(fetchSummary).toHaveBeenCalledTimes(4);
		expect(gone).toEqual(['BBBBB']);
		expect(games.map((g) => [g.code, g.phase])).toEqual([
			['AAAAA', 'playing'],
			['CCCCC', null],
			['DDDDD', null]
		]);
	});

	it('requests all summaries in parallel', async () => {
		let inFlight = 0;
		let peak = 0;
		const fetchSummary = async (): Promise<GameSummary> => {
			peak = Math.max(peak, ++inFlight);
			await new Promise((r) => setTimeout(r, 5));
			inFlight--;
			return summary();
		};
		await loadRecentGames(stored, fetchSummary, () => {});
		expect(peak).toBe(stored.length);
	});

	it('returns an empty list for no stored games', async () => {
		expect(await loadRecentGames([], vi.fn(), vi.fn())).toEqual([]);
	});
});
