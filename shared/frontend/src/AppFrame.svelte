<!--
	The frame of every game page: the "All games" link on the home screen, the light/dark toggle
	top right, the page and the footer with the imprint and privacy links. The game's +layout.svelte renders it inside its
	ModeWatcher and Toaster.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';
	import LayoutGrid from '@lucide/svelte/icons/layout-grid';
	import ThemeToggle from './ThemeToggle.svelte';
	import { Button } from './ui/button/index.js';
	import { imprintHref, privacyHref, startPageHref } from './startPage.js';

	let {
		url,
		isHome,
		dev,
		children
	}: {
		/** The current page URL (page.url): links go to the same host. */
		url: Pick<URL, 'protocol' | 'hostname'>;
		/** Whether this is the game's home screen (shows the "All games" link). */
		isHome: boolean;
		/** import.meta.env.DEV: links go to the dev servers' ports. */
		dev: boolean;
		children: Snippet;
	} = $props();
</script>

<main class="mx-auto flex min-h-svh w-full max-w-5xl flex-col px-3 py-3 sm:p-6">
	<div class="-mt-1 -mr-1 mb-1 flex items-center justify-end">
		{#if isHome}
			<!-- The start page is a separate app (same host in prod, own port in dev): a full page load. -->
			<Button
				variant="ghost"
				size="sm"
				class="mr-auto -ml-2 text-muted-foreground"
				href={startPageHref(url, dev)}
				rel="external"
			>
				<LayoutGrid class="size-4" />
				All games
			</Button>
		{/if}
		<ThemeToggle />
	</div>
	{@render children()}
	<footer class="mt-auto flex justify-center gap-4 pt-10 text-center text-xs text-muted-foreground">
		<!-- Pages of the start page app: full page loads. -->
		<a href={imprintHref(url, dev)} rel="external" class="underline-offset-4 hover:underline"
			>Impressum</a
		>
		<a href={privacyHref(url, dev)} rel="external" class="underline-offset-4 hover:underline"
			>Datenschutz</a
		>
	</footer>
</main>
