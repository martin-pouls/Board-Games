// The home screen's list of games remembered in this browser: which ones can be
// continued, which are history, and how to load their summaries.
import { ApiError } from '#lib/api.js';
import type { GameSummary, Phase } from '#lib/types.js';

/** At most this many games are offered under "Continue". */
export const MAX_CONTINUE = 5;

export interface RecentGame {
	code: string;
	/** The nickname used in that game. */
	name: string;
	/** null when the summary could not be loaded (offline, server error). */
	phase: Phase | null;
	/** Creation time in ms since the epoch; 0 when unknown. */
	createdAt: number;
}

/** Lobby and running games can be continued; finished ones (won, lost) cannot. */
export function isContinuable(g: Pick<RecentGame, 'phase'>): boolean {
	// An unknown phase (network error) stays continuable: the game page reports the problem.
	return g.phase !== 'won' && g.phase !== 'lost';
}

/** Newest first; games without a date, then equal dates, fall back to the code for a stable order. */
export function byNewest(a: RecentGame, b: RecentGame): number {
	return b.createdAt - a.createdAt || a.code.localeCompare(b.code);
}

/** Continuable games (newest first, at most `max`) and the finished rest (newest first). */
export function partitionGames(
	games: RecentGame[],
	max = MAX_CONTINUE
): { continuable: RecentGame[]; history: RecentGame[] } {
	const sorted = [...games].sort(byNewest);
	return {
		continuable: sorted.filter(isContinuable).slice(0, max),
		history: sorted.filter((g) => !isContinuable(g))
	};
}

/** Parses an RFC3339 timestamp to ms; 0 for missing or invalid values. */
export function parseCreatedAt(value: string | undefined): number {
	if (!value) return 0;
	const t = Date.parse(value);
	return Number.isNaN(t) ? 0 : t;
}

export function fromSummary(code: string, name: string, s: GameSummary): RecentGame {
	return {
		code,
		name,
		phase: s.phase,
		createdAt: parseCreatedAt(s.createdAt)
	};
}

/** Short status label for a row. */
export function statusLabel(g: RecentGame): string {
	switch (g.phase) {
		case 'lobby':
			return 'Lobby';
		case 'playing':
			return 'In progress';
		case 'won':
			return 'Won';
		case 'lost':
			return 'Lost';
		default:
			return '';
	}
}

/**
 * Loads the summary of every stored game in parallel. Games the server no longer knows
 * (deleted by the janitor) are dropped and reported through `onGone`; any other failure
 * keeps the game with an unknown phase, so a flaky network never empties the list.
 */
export async function loadRecentGames(
	stored: { code: string; name: string }[],
	fetchSummary: (code: string) => Promise<GameSummary>,
	onGone: (code: string) => void
): Promise<RecentGame[]> {
	const results = await Promise.all(
		stored.map(async ({ code, name }): Promise<RecentGame | null> => {
			try {
				return fromSummary(code, name, await fetchSummary(code));
			} catch (e) {
				if (e instanceof ApiError && (e.code === 'game_not_found' || e.status === 404)) {
					onGone(code);
					return null;
				}
				return { code, name, phase: null, createdAt: 0 };
			}
		})
	);
	return results.filter((g): g is RecentGame => g !== null);
}
