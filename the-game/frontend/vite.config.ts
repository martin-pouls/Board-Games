import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vitest/config';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';

const backend = process.env.BACKEND_URL ?? 'http://localhost:8080';

// Sub-path the app is served under in production (e.g. BASE_PATH=/the-game behind Caddy);
// empty in dev. Routes use resolve(); API and WS URLs read __BASE_PATH__.
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
			// SPA mode: unknown routes fall back to index.html (served by the Go backend in prod).
			adapter: adapter({ fallback: 'index.html' }),
			paths: { base }
		})
	],
	define: { __BASE_PATH__: JSON.stringify(base) },
	server: {
		proxy: {
			'/api': backend,
			'/ws': { target: backend, ws: true }
		}
	},
	test: {
		expect: { requireAssertions: true },
		passWithNoTests: true,
		projects: [
			{
				extends: './vite.config.ts',
				test: {
					name: 'server',
					environment: 'node',
					include: ['src/**/*.{test,spec}.{js,ts}'],
					exclude: ['src/**/*.svelte.{test,spec}.{js,ts}']
				}
			}
		]
	}
});
