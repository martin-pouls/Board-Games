import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vitest/config';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';

// Sub-path the start page is served under in production (BASE_PATH=/board-games behind Caddy);
// empty in dev. The games live next to it at /the-game, /regicide and /fox-duet; in dev the tiles link to
// the games' own dev servers instead (see src/lib/games.ts).
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
	// Fixed local port (The Game 5173, Regicide 5174, start page 5175, Fox Duet 5176).
	server: { port: 5175, strictPort: true },
	test: {
		expect: { requireAssertions: true },
		passWithNoTests: true,
		include: ['src/**/*.{test,spec}.{js,ts}']
	}
});
