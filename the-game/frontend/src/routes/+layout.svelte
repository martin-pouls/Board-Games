<script lang="ts">
	import './layout.css';
	import favicon from '#lib/assets/favicon.svg';
	import { Toaster } from '#lib/components/ui/sonner/index.js';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>The Game</title>
</svelte:head>

<Toaster />
<main class="mx-auto flex min-h-svh w-full max-w-5xl flex-col px-3 py-3 sm:p-6">
	{@render children()}
</main>
