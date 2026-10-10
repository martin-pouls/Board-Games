<script lang="ts">
	import './layout.css';
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

<main class="mx-auto flex min-h-svh w-full max-w-3xl flex-col px-4 py-10 sm:py-16">
	{@render children()}
</main>
