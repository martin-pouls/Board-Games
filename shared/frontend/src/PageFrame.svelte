<!--
	The frame of every page (games, start page, gate): the top bar, a centred column for the page and
	the footer with the privacy link, which opens LegalDialog (left out on the games'
	table). The games use it through AppFrame.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';
	import TopBar from './TopBar.svelte';
	import LegalDialog from './legal/LegalDialog.svelte';
	import { legalDialogId, type LegalPage } from './legal/legal.js';

	let {
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
		/** false hides the privacy footer (the games' table, which must fit a phone). */
		footer?: boolean;
		/** More items on the left of the top bar. */
		start?: Snippet;
		children: Snippet;
	} = $props();

	/** Opens the page's dialog instead of following the link (a new tab still follows it). */
	function openLegal(e: MouseEvent, page: LegalPage) {
		if (e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
		const dialog = document.getElementById(legalDialogId(page));
		if (!(dialog instanceof HTMLDialogElement)) return;
		e.preventDefault();
		dialog.showModal();
	}
</script>

<!--
	The privacy policy opens as a dialog over the page, in its theme. The link stays a link to the
	start page's page (full page load), the way there without JS. In the games the click handler opens
	the dialog; the start page and the gate ship no JS, so their app.html script handles data-legal.
-->
{#snippet legalLink(page: LegalPage, href: string, label: string)}
	<a
		{href}
		rel="external"
		data-legal={page}
		onclick={(e) => openLegal(e, page)}
		class="underline-offset-4 hover:underline">{label}</a
	>
{/snippet}

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
			{@render legalLink('datenschutz', privacyHref, 'Datenschutz')}
		</footer>
		<LegalDialog page="datenschutz" />
	{/if}
</main>
