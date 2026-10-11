import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vitest/config';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { sharedAlias, sharedDedupe, sharedRoot, sharedSrc } from '../../shared/frontend/vite.js';

// Sub-path the login page and its assets are served under in production (BASE_PATH=/gate behind
// Caddy, which strips it); empty for the gate's own `go run` and in dev.
const base = (process.env.BASE_PATH ?? '') as '' | `/${string}`;

// The build goes straight into the Go package, which embeds it: the page becomes the gate's
// html/template, its assets are served under /_app/ (see backend/internal/gate/ui.go).
const out = '../backend/internal/gate/ui/build';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			// Fully prerendered: one HTML page, its CSS and fonts. No client-side JavaScript.
			adapter: adapter({ pages: out, assets: out }),
			// Absolute asset URLs: the page is served at <base>/login, not where it was prerendered.
			paths: { base, relative: false }
		})
	],
	// Shared frontend code (the page frame), compiled as source; the packages it imports resolve
	// from this app. shared/frontend/vite.ts explains the setup.
	resolve: { alias: { [sharedAlias]: sharedSrc }, dedupe: sharedDedupe },
	server: {
		fs: { allow: [sharedRoot] },
		// Fixed local port (The Game 5173, Regicide 5174, start page 5175, Fox Duet 5176, The Crew 5177).
		port: 5178,
		strictPort: true
	},
	test: {
		expect: { requireAssertions: true },
		passWithNoTests: true,
		include: ['src/**/*.{test,spec}.{js,ts}']
	}
});
