<!--
	The bar at the top of every page (games, start page, gate): optional links on the left, the
	light/dark toggle on the right. Each part on the left is switched on by passing it.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';
	import LayoutGrid from '@lucide/svelte/icons/layout-grid';
	import HeaderButton, { headerRow } from './HeaderButton.svelte';
	import ThemeToggle from './ThemeToggle.svelte';
	import { cn } from './utils.js';

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

<div class={cn(headerRow, 'mb-1')}>
	{#if allGamesHref}
		<HeaderButton
			icon={LayoutGrid}
			label="All games"
			showLabel
			href={allGamesHref}
			rel="external"
		/>
	{/if}
	{@render start?.()}
	<div class="ml-auto">
		<ThemeToggle onclick={onToggleTheme} />
	</div>
</div>
