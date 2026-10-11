<!--
	The frame of every game page: PageFrame with the "All games" link on the home screen, the theme
	toggle wired to mode-watcher and the links to the start page's imprint and privacy pages. The
	game's +layout.svelte renders it inside its ModeWatcher and Toaster.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';
	import { toggleMode } from 'mode-watcher';
	import PageFrame from './PageFrame.svelte';
	import { imprintHref, privacyHref, startPageHref } from './startPage.js';

	let {
		url,
		isHome,
		dev,
		wide = false,
		children
	}: {
		/** The current page URL (page.url): links go to the same host. */
		url: Pick<URL, 'protocol' | 'hostname'>;
		/** Whether this is the game's home screen (shows the "All games" link). */
		isHome: boolean;
		/** import.meta.env.DEV: links go to the dev servers' ports. */
		dev: boolean;
		/** A wider column on desktop (see PageFrame). */
		wide?: boolean;
		children: Snippet;
	} = $props();
</script>

<PageFrame
	imprintHref={imprintHref(url, dev)}
	privacyHref={privacyHref(url, dev)}
	allGamesHref={isHome ? startPageHref(url, dev) : undefined}
	onToggleTheme={toggleMode}
	{wide}
>
	{@render children()}
</PageFrame>
