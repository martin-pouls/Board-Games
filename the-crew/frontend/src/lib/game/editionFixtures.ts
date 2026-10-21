import { edition } from '#lib/edition.js';
import type { PlayerView } from '#lib/types.js';
import type { TablePreset } from './tablePreset.js';

/**
 * The fixtures of this build's edition, loaded on demand (dev fixtures route), and the table UI
 * state some of them start with (a card picked, the communicate flow), derived from the view.
 */
export async function fixturesForEdition(): Promise<{
	fixtures: Record<string, () => PlayerView>;
	presets: Record<string, (v: PlayerView) => TablePreset>;
}> {
	if (edition === 'deepSea') {
		const { deepSeaFixtures, deepSeaLobbyFixtures, deepSeaUiFixtures, deepSeaPresets } =
			await import('./fixturesDeepSea.js');
		return {
			fixtures: { ...deepSeaLobbyFixtures, ...deepSeaFixtures, ...deepSeaUiFixtures },
			presets: deepSeaPresets
		};
	}
	const { fixtures, presets } = await import('./fixtures.js');
	return { fixtures, presets };
}
