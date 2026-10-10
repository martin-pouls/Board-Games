<script lang="ts">
	import './layout.css';
	import ThemeToggle from '#lib/components/ThemeToggle.svelte';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<svelte:head>
	<title>Card Games</title>
	<meta
		name="description"
		content="Cooperative card games to play live with friends in the browser."
	/>
</svelte:head>

<main class="mx-auto flex min-h-svh w-full max-w-3xl flex-col px-4 pt-3 pb-10 sm:pt-6 sm:pb-16">
	<div class="-mr-1 mb-4 flex justify-end sm:mb-6">
		<ThemeToggle />
	</div>
	{@render children()}
</main>
