<script lang="ts">
	import './layout.css';
	import { page } from '$app/state';
	import favicon from '#lib/assets/favicon.svg';
	import { ModeWatcher } from 'mode-watcher';
	import AppFrame from '@board-games/shared/AppFrame.svelte';
	import MistRidges from '#lib/components/MistRidges.svelte';
	import { Toaster } from '#lib/components/ui/sonner/index.js';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>Regicide</title>
</svelte:head>

<ModeWatcher themeColors={{ light: '#def1f1', dark: '#08181f' }} />
<Toaster />
<MistRidges />
<AppFrame
	url={page.url}
	isHome={page.route.id === '/'}
	topBar={page.route.id !== '/g/[code]' && page.route.id !== '/dev/fixtures/[name]'}
	dev={import.meta.env.DEV}
>
	{@render children()}
</AppFrame>
