<!--
	The running text of a legal page, the same in every app and set in the app's own theme: the
	start page's pages and every app's LegalDialog render it.
-->
<script lang="ts">
	import Imprint from './Imprint.svelte';
	import Privacy from './Privacy.svelte';
	import type { LegalPage } from './legal.js';

	let { page, class: className }: { page: LegalPage; class?: string } = $props();
</script>

<div lang="de" class={['legal', className]}>
	{#if page === 'impressum'}
		<Imprint />
	{:else}
		<Privacy />
	{/if}
</div>

<style>
	.legal {
		display: flex;
		flex-direction: column;
		gap: 0.75em;
		line-height: 1.65;
	}
	.legal :global(h2) {
		margin-top: 1em;
		font-size: 1.25em;
		font-weight: 600;
		line-height: 1.3;
	}
	.legal :global(h3) {
		margin-top: 0.5em;
		font-weight: 600;
		line-height: 1.4;
	}
	.legal :global(address) {
		font-style: normal;
	}
	.legal :global(ul) {
		display: flex;
		flex-direction: column;
		gap: 0.25em;
		padding-left: 1.25em;
		list-style: disc;
	}
</style>
