import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vitest/config';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { sharedAlias, sharedDedupe, sharedRoot, sharedSrc } from '../../shared/frontend/vite.js';

const backend = process.env.BACKEND_URL ?? 'http://localhost:8082';

// Sub-path the app is served under in production (e.g. BASE_PATH=/fox-duet behind Caddy);
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
	// Shared frontend code, compiled as source; the packages it imports resolve from this app
	// (one svelte, bits-ui, mode-watcher, ...). shared/frontend/vite.ts explains the setup.
	resolve: { alias: { [sharedAlias]: sharedSrc }, dedupe: sharedDedupe },
	define: { __BASE_PATH__: JSON.stringify(base) },
	// Dev only: scan every route for dependencies at startup. Otherwise Vite discovers them (bits-ui,
	// lucide icons, ...) on first navigation, re-bundles and reloads the page mid-click.
	optimizeDeps: { entries: ['src/routes/**/+*.svelte'] },
	server: {
		// Dev only: compile all components at startup, so the first visit to a page is not slow.
		warmup: {
			clientFiles: [
				'./src/routes/**/*.svelte',
				'./src/lib/**/*.{svelte,ts}',
				'../../shared/frontend/src/**/*.{svelte,ts}'
			]
		},
		fs: { allow: [sharedRoot] },
		// Fixed local port (The Game 5173, Regicide 5174, start page 5175, Fox Duet 5176): the start page links here in dev.
		port: 5176,
		strictPort: true,
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
