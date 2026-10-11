import { svelte } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	plugins: [svelte({ compilerOptions: { runes: true } })],
	test: {
		expect: { requireAssertions: true },
		passWithNoTests: true,
		environment: 'node',
		include: ['src/**/*.{test,spec}.{js,ts}']
	}
});
