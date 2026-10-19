<!--
	The frame of every page (games, start page, gate): the top bar, a centred column for the page and
	the footer with the imprint and privacy links (left out on the games' table). The games use it
	through AppFrame.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';
	import TopBar from './TopBar.svelte';

	let {
		imprintHref,
		privacyHref,
		allGamesHref,
		onToggleTheme,
		wide = false,
		topBar = true,
		topBarWidth,
		footer = true,
		start,
		children
	}: {
		imprintHref: string;
		privacyHref: string;
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
		/** false hides the imprint and privacy footer (the games' table, which must fit a phone). */
		footer?: boolean;
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
	{#if footer}
		<footer
			class="mt-auto flex justify-center gap-4 pt-10 text-center text-xs text-muted-foreground"
		>
			<!-- Pages of the start page app: full page loads. -->
			<a href={imprintHref} rel="external" class="underline-offset-4 hover:underline">Impressum</a>
			<a href={privacyHref} rel="external" class="underline-offset-4 hover:underline">Datenschutz</a
			>
		</footer>
	{/if}
</main>
