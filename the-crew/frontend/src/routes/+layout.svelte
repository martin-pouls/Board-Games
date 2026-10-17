<script lang="ts">
	import './layout.css';
	import { page } from '$app/state';
	import favicon from '#lib/assets/favicon.svg';
	import faviconDeepSea from '#lib/assets/favicon-deep-sea.svg';
	import { edition, terms } from '#lib/edition.js';
	import { ModeWatcher } from 'mode-watcher';
	import AppFrame from '@board-games/shared/AppFrame.svelte';
	import { Toaster } from '#lib/components/ui/sonner/index.js';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<svelte:head>
	<link rel="icon" href={edition === 'deepSea' ? faviconDeepSea : favicon} />
	<title>{terms.title}</title>
</svelte:head>

<ModeWatcher themeColors={terms.themeColor} />
<Toaster />
<AppFrame
	url={page.url}
	isHome={page.route.id === '/'}
	topBar={page.route.id !== '/g/[code]'}
	dev={import.meta.env.DEV}
>
	{@render children()}
</AppFrame>
