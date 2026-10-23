import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

// Dev only: in a production build the branch, and with it the fixtures, is dropped.
export const load: PageLoad = async () => {
	if (import.meta.env.DEV) {
		const { fixtures } = await import('#lib/game/fixtures.js');
		return { names: Object.keys(fixtures) };
	}
	error(404, 'Not found');
};
