<script lang="ts">
	import './layout.css';
	import { page } from '$app/state';
	import favicon from '#lib/assets/favicon.svg';
	import { ModeWatcher } from 'mode-watcher';
	import AppFrame from '@board-games/shared/AppFrame.svelte';
	import { Toaster } from '#lib/components/ui/sonner/index.js';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>Fox Duet</title>
</svelte:head>

<ModeWatcher themeColors={{ light: '#f3eedd', dark: '#0f1d17' }} />
<Toaster />
<AppFrame url={page.url} isHome={page.route.id === '/'} dev={import.meta.env.DEV}>
	{@render children()}
</AppFrame>
