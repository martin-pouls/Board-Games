<!--
	The frame of every game page: PageFrame with the "All games" link on the home screen, the theme
	toggle wired to mode-watcher and the link to the start page's privacy page. The
	game's +layout.svelte renders it inside its ModeWatcher and Toaster. On the game page itself
	(/g/[code]) the layout passes topBar={false}: its headers show HeaderThemeToggle instead, and the
	footer is left out so the table fits a phone (the home screen links the privacy policy). The
	pages that do show the top bar (home, join) use a max-w-md column, so the bar is that wide too.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';
	import { toggleMode } from 'mode-watcher';
	import PageFrame from './PageFrame.svelte';
	import { privacyHref, startPageHref } from './startPage.js';

	let {
		url,
		isHome,
		dev,
		wide = false,
		topBar = true,
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
		/** false hides the top bar and the footer; the page's own header then shows HeaderThemeToggle. */
		topBar?: boolean;
		children: Snippet;
	} = $props();
</script>

<PageFrame
	privacyHref={privacyHref(url, dev)}
	allGamesHref={isHome ? startPageHref(url, dev) : undefined}
	onToggleTheme={toggleMode}
	{wide}
	{topBar}
	footer={topBar}
	topBarWidth="max-w-md"
>
	{@render children()}
</PageFrame>
