<script lang="ts">
	import './layout.css';
	import favicon from '#lib/assets/favicon.svg';
	import { ModeWatcher } from 'mode-watcher';
	import ThemeToggle from '#lib/components/ThemeToggle.svelte';
	import { Toaster } from '#lib/components/ui/sonner/index.js';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>The Game</title>
</svelte:head>

<ModeWatcher themeColors={{ light: '#ffffff', dark: '#09090b' }} />
<Toaster />
<main class="mx-auto flex min-h-svh w-full max-w-5xl flex-col px-3 py-3 sm:p-6">
	<div class="-mt-1 -mr-1 mb-1 flex justify-end">
		<ThemeToggle />
	</div>
	{@render children()}
</main>
