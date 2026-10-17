import { edition } from '#lib/edition.js';
import type { PlayerView } from '#lib/types.js';

/** The fixtures of this build's edition, loaded on demand (dev fixtures route). */
export async function fixturesForEdition(): Promise<{
	fixtures: Record<string, () => PlayerView>;
}> {
	if (edition === 'deepSea') {
		const { deepSeaFixtures, deepSeaLobbyFixtures } = await import('./fixturesDeepSea.js');
		return { fixtures: { ...deepSeaLobbyFixtures, ...deepSeaFixtures } };
	}
	return import('./fixtures.js');
}
