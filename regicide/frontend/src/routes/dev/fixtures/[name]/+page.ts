import { error } from '@sveltejs/kit';
import type { PlayerView } from '#lib/types.js';
import type { PageLoad } from './$types';

// Dev only: in a production build the branch, and with it the fixtures, is dropped.
export const load: PageLoad = async ({ params }) => {
	if (import.meta.env.DEV) {
		const { fixtures } = await import('#lib/game/fixtures.js');
		const all: Record<string, () => PlayerView> = fixtures;
		if (Object.hasOwn(all, params.name)) {
			return { name: params.name, view: all[params.name]() };
		}
	}
	error(404, 'No such fixture');
};
