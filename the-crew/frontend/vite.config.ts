import tailwindcss from '@tailwindcss/vite';
import type { Plugin } from 'vite';
import { defineConfig } from 'vitest/config';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { sharedAlias, sharedDedupe, sharedRoot, sharedSrc } from '../../shared/frontend/vite.js';

// The edition (Planet Nine or Mission Deep Sea) is fixed per build. Pick it with the env var EDITION
// (CI, Docker) or, locally and cross-platform, with the Vite mode: `vite dev --mode deepSea`
// (npm run dev:deep-sea). EDITION wins over the mode; the default is Planet Nine. src/app.html
// (theme before first paint) reads it through the placeholder below.
const editions = ['planetNine', 'deepSea'] as const;

function pickEdition(mode: string): (typeof editions)[number] {
	const wanted = process.env.EDITION || (mode === 'deepSea' ? 'deepSea' : 'planetNine');
	const edition = editions.find((e) => e === wanted);
	if (!edition) {
		throw new Error(`EDITION must be one of ${editions.join(', ')}, got "${wanted}"`);
	}
	return edition;
}

// SvelteKit 3 renders app.html without build-time env (%sveltekit.env.*% is empty in the SPA
// fallback), so this plugin swaps the placeholder in the generated template module for the edition.
const templatePlaceholder = '__EDITION_FOR_APP_HTML__';
function editionInTemplate(edition: string): Plugin {
	return {
		name: 'edition-in-app-html',
		transform(code) {
			if (code.includes(templatePlaceholder)) return code.replaceAll(templatePlaceholder, edition);
		}
	};
}

// Sub-path the app is served under in production (e.g. BASE_PATH=/the-crew behind Caddy);
// empty in dev. Routes use resolve(); API and WS URLs read __BASE_PATH__.
const base = (process.env.BASE_PATH ?? '') as '' | `/${string}`;

export default defineConfig(({ mode }) => {
	const edition = pickEdition(mode);
	const deepSea = edition === 'deepSea';
	// Planet Nine and Deep Sea run side by side locally, each with its own backend.
	const backend = process.env.BACKEND_URL ?? `http://localhost:${deepSea ? 8085 : 8084}`;

	return {
		plugins: [
			tailwindcss(),
			editionInTemplate(edition),
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
		define: { __BASE_PATH__: JSON.stringify(base), __EDITION__: JSON.stringify(edition) },
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
			// Fixed local port (The Game 5173, Regicide 5174, start page 5175, Fox Duet 5176, The Crew 5177,
			// Deep Sea 5179; the gate uses 5178): the start page links here in dev.
			port: deepSea ? 5179 : 5177,
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
	};
});
