import { error } from '@sveltejs/kit';
import type { PlayerView } from '#lib/types.js';
import type { PageLoad } from './$types';
import { fixturesForEdition } from '#lib/game/editionFixtures.js';

// Dev only: in a production build the branch, and with it the fixtures, is dropped.
export const load: PageLoad = async ({ params }) => {
	if (import.meta.env.DEV) {
		const { fixtures, presets } = await fixturesForEdition();
		const all: Record<string, () => PlayerView> = fixtures;
		if (Object.hasOwn(all, params.name)) {
			const view = all[params.name]();
			const preset = Object.hasOwn(presets, params.name) ? presets[params.name](view) : null;
			return { name: params.name, view, preset };
		}
	}
	error(404, 'No such fixture');
};
