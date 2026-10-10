import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vitest/config';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';

// Sub-path the start page is served under in production (BASE_PATH=/card-games behind Caddy);
// empty in dev. The games live next to it at /the-game and /regicide (see src/lib/games.ts).
const base = (process.env.BASE_PATH ?? '') as '' | `/${string}`;

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			// Fully prerendered: plain static files, no SPA fallback needed.
			adapter: adapter(),
			paths: { base }
		})
	],
	test: {
		expect: { requireAssertions: true },
		passWithNoTests: true,
		include: ['src/**/*.{test,spec}.{js,ts}']
	}
});
