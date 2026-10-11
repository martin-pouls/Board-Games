<script lang="ts">
	import './layout.css';
	import { resolve } from '$app/paths';
	import ThemeToggle from '#lib/components/ThemeToggle.svelte';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<main class="mx-auto flex min-h-svh w-full max-w-3xl flex-col px-4 pt-3 pb-6 sm:pt-6">
	<div class="-mr-1 mb-4 flex justify-end sm:mb-6">
		<ThemeToggle />
	</div>
	{@render children()}
	<footer class="mt-auto flex justify-center gap-4 pt-12 text-center text-xs text-muted-foreground">
		<!-- Trailing slashes: the pages are <name>/index.html (see datenschutz/+page.ts). -->
		<a href="{resolve('/impressum')}/" class="underline-offset-4 hover:underline">Impressum</a>
		<a href="{resolve('/datenschutz')}/" class="underline-offset-4 hover:underline">Datenschutz</a>
	</footer>
</main>
