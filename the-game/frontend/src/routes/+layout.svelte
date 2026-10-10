<script lang="ts">
	import './layout.css';
	import { page } from '$app/state';
	import favicon from '#lib/assets/favicon.svg';
	import { ModeWatcher } from 'mode-watcher';
	import ThemeToggle from '#lib/components/ThemeToggle.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { privacyHref, startPageHref } from '#lib/startPage.js';
	import LayoutGrid from '@lucide/svelte/icons/layout-grid';
	import { Toaster } from '#lib/components/ui/sonner/index.js';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>The Game</title>
</svelte:head>

<ModeWatcher themeColors={{ light: '#f3f6fa', dark: '#0d1119' }} />
<Toaster />
<main class="mx-auto flex min-h-svh w-full max-w-5xl flex-col px-3 py-3 sm:p-6">
	<div class="-mt-1 -mr-1 mb-1 flex items-center justify-end">
		{#if page.route.id === '/'}
			<!-- The start page is a separate app (same host in prod, own port in dev): a full page load. -->
			<Button
				variant="ghost"
				size="sm"
				class="mr-auto -ml-2 text-muted-foreground"
				href={startPageHref(page.url, import.meta.env.DEV)}
				rel="external"
			>
				<LayoutGrid class="size-4" />
				All games
			</Button>
		{/if}
		<ThemeToggle />
	</div>
	{@render children()}
	<footer class="mt-auto pt-10 text-center text-xs text-muted-foreground">
		<!-- A page of the start page app: a full page load. -->
		<a
			href={privacyHref(page.url, import.meta.env.DEV)}
			rel="external"
			class="underline-offset-4 hover:underline">Datenschutz</a
		>
	</footer>
</main>
