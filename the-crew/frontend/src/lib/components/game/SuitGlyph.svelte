<script lang="ts">
	import { edition } from '#lib/edition.js';
	import type { Suit } from '#lib/types.js';

	// Each colour has its own symbol, so the cards read without colour vision too: line symbols
	// (pink a bracket, blue two turning arrows, green a triangle, yellow a cross, each with a dot),
	// the same in both editions. The trump: Planet Nine a solid rocket with a dark window, Deep Sea
	// the bathyscaphe (float tank, hull on skids, the glass dome in front, its lamp), its dome in the
	// dark of the submarine window.
	let { suit, class: className = '' }: { suit: Suit; class?: string } = $props();
</script>

<svg
	viewBox="0 0 24 24"
	class={className}
	fill="none"
	stroke="currentColor"
	stroke-width="2.4"
	stroke-linecap="round"
	stroke-linejoin="round"
	aria-hidden="true"
>
	{#if suit === 'pink'}
		<path d="M10 5H5v14h14v-5M14 5h5v5" />
		<circle cx="12" cy="5" r="1.6" fill="currentColor" stroke="none" />
	{:else if suit === 'blue'}
		<path d="M18.5 9A7 7 0 0 0 6 7.5M5.5 15A7 7 0 0 0 18 16.5" />
		<path d="M6 3.5v4h4M18 20.5v-4h-4" />
	{:else if suit === 'green'}
		<path d="M4 5h16l-8 14z" />
		<circle cx="19.5" cy="18.5" r="1.6" fill="currentColor" stroke="none" />
	{:else if suit === 'yellow'}
		<path d="M5 5l14 14M19 5L5 19" />
		<circle cx="5" cy="5" r="1.8" fill="currentColor" stroke="none" />
		<circle cx="19" cy="19" r="1.8" fill="currentColor" stroke="none" />
	{:else if edition === 'deepSea'}
		<g stroke="none" fill="currentColor">
			<rect x="4.5" y="4.6" width="9" height="3.6" rx="1.8" />
			<path d="M9 8v2" stroke="currentColor" stroke-width="1.8" />
			<rect x="1.5" y="9.5" width="13" height="8" rx="4" />
			<path d="M4.5 19.6h9" stroke="currentColor" stroke-width="1.4" />
			<path d="M6 17v2.6M12 17v2.6" stroke="currentColor" stroke-width="1.2" />
			<circle cx="15" cy="13.5" r="5.2" />
			<circle cx="15.6" cy="13.3" r="3.6" fill="var(--nd-trump)" />
			<circle cx="16.7" cy="12.2" r="0.95" fill-opacity="0.6" />
			<path d="M20 11.7 23.6 9.6v8l-3.6-2.1z" fill-opacity="0.42" />
		</g>
	{:else}
		<path
			d="M12 2c3 2.5 4.5 6 4.5 10v5h-9v-5C7.5 8 9 4.5 12 2z"
			fill="currentColor"
			stroke="none"
		/>
		<path d="M7.5 13 4.5 19l3-1M16.5 13l3 6-3-1" fill="currentColor" stroke="none" />
		<circle cx="12" cy="9.5" r="1.8" fill="var(--nd-trump)" stroke="none" />
	{/if}
</svg>
