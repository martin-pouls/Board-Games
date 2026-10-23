<!--
	The frame of every page (games, start page, gate): the top bar and a centred column for the page.
	The games use it through AppFrame.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';
	import TopBar from './TopBar.svelte';

	let {
		allGamesHref,
		onToggleTheme,
		wide = false,
		topBar = true,
		topBarWidth,
		start,
		children
	}: {
		/** Shows the "All games" link in the top bar. */
		allGamesHref?: string;
		/** Click handler of the theme toggle; without it, the page's theme script handles the click. */
		onToggleTheme?: () => void;
		/** A wider column on desktop, for screens that need the room (e.g. a big board). */
		wide?: boolean;
		/** false hides the top bar, for pages whose own header has the theme toggle (the games' table). */
		topBar?: boolean;
		/** A max-width class for the top bar, when the page's column is narrower than the frame's:
		 *  the bar then lines up with it. */
		topBarWidth?: string;
		/** More items on the left of the top bar. */
		start?: Snippet;
		children: Snippet;
	} = $props();
</script>

<main
	class={[
		'mx-auto flex min-h-svh w-full flex-col px-3 py-3 sm:p-6',
		wide ? 'max-w-6xl' : 'max-w-3xl'
	]}
>
	{#if topBar}
		<div class={['mx-auto w-full', topBarWidth]}>
			<TopBar {allGamesHref} {onToggleTheme} {start} />
		</div>
	{/if}
	{@render children()}
</main>
