<!--
	The bar at the top of every page (games, start page, gate): optional links on the left, the
	light/dark toggle on the right. Each part on the left is switched on by passing it.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';
	import LayoutGrid from '@lucide/svelte/icons/layout-grid';
	import ThemeToggle from './ThemeToggle.svelte';
	import { Button } from './ui/button/index.js';

	let {
		allGamesHref,
		onToggleTheme,
		start
	}: {
		/** Shows the "All games" link to the start page (a separate app: a full page load). */
		allGamesHref?: string;
		/** Click handler of the theme toggle; without it, the page's theme script handles the click. */
		onToggleTheme?: () => void;
		/** More items on the left, after "All games". */
		start?: Snippet;
	} = $props();
</script>

<div class="-mt-1 -mr-1 mb-1 flex min-h-8 items-center gap-1">
	{#if allGamesHref}
		<Button
			variant="ghost"
			size="sm"
			class="-ml-2 text-muted-foreground"
			href={allGamesHref}
			rel="external"
		>
			<LayoutGrid class="size-4" />
			All games
		</Button>
	{/if}
	{@render start?.()}
	<div class="ml-auto">
		<ThemeToggle onclick={onToggleTheme} />
	</div>
</div>
