import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { fixturesForEdition } from '#lib/game/editionFixtures.js';

// Dev only: in a production build the branch, and with it the fixtures, is dropped.
export const load: PageLoad = async () => {
	if (import.meta.env.DEV) {
		const { fixtures } = await fixturesForEdition();
		return { names: Object.keys(fixtures) };
	}
	error(404, 'Not found');
};
