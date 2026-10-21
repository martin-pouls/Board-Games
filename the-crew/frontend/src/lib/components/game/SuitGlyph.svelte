<script lang="ts">
	import { edition } from '#lib/edition.js';
	import type { Suit } from '#lib/types.js';

	// Each colour has its own symbol, so the cards read without colour vision too. Planet Nine: line
	// symbols (pink a bracket, blue two turning arrows, green a triangle, yellow a cross, each with a
	// dot) and a solid rocket with a dark window. Deep Sea keeps its solid ones: pink a ringed
	// planet, blue a wave, green a leaf, yellow a star, a submarine.
	let { suit, class: className = '' }: { suit: Suit; class?: string } = $props();
</script>

{#if edition !== 'deepSea'}
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
{:else}
	<svg viewBox="0 0 24 24" class={className} fill="currentColor" aria-hidden="true">
		{#if suit === 'pink'}
			<circle cx="12" cy="12" r="6" />
			<ellipse
				cx="12"
				cy="12"
				rx="11"
				ry="3.6"
				fill="none"
				stroke="currentColor"
				stroke-width="1.8"
				transform="rotate(-20 12 12)"
			/>
		{:else if suit === 'blue'}
			<path
				d="M2 9.5c2.2 0 2.2-2 4.5-2s2.3 2 4.5 2 2.2-2 4.5-2 2.3 2 4.5 2v3.2c-2.2 0-2.2-2-4.5-2s-2.3 2-4.5 2-2.2-2-4.5-2-2.3 2-4.5 2Zm0 6c2.2 0 2.2-2 4.5-2s2.3 2 4.5 2 2.2-2 4.5-2 2.3 2 4.5 2v3.2c-2.2 0-2.2-2-4.5-2s-2.3 2-4.5 2-2.2-2-4.5-2-2.3 2-4.5 2Z"
			/>
		{:else if suit === 'green'}
			<path d="M4 20C4 10 10 4 21 3c-.5 11-6.5 17-15.5 17.2L4 22Zm3.2-3.5 8.6-9.4-.9-.8-8.6 9.4Z" />
		{:else if suit === 'yellow'}
			<polygon
				points="12,2 14.6,8.4 21.5,9 16.2,13.4 17.9,20.2 12,16.5 6.1,20.2 7.8,13.4 2.5,9 9.4,8.4"
			/>
		{:else}
			<!-- Hull with three portholes (even-odd holes), conning tower, periscope and propeller. -->
			<g transform="translate(0 1.5)">
				<path
					fill-rule="evenodd"
					d="M6 10h9a4 4 0 0 1 0 8H6a4 4 0 0 1 0-8Zm.5 2.7a1.3 1.3 0 1 0 0 2.6 1.3 1.3 0 0 0 0-2.6Zm4 0a1.3 1.3 0 1 0 0 2.6 1.3 1.3 0 0 0 0-2.6Zm4 0a1.3 1.3 0 1 0 0 2.6 1.3 1.3 0 0 0 0-2.6Z"
				/>
				<path d="M8.5 6.5h5a1.2 1.2 0 0 1 1.2 1.2V10H7.3V7.7a1.2 1.2 0 0 1 1.2-1.2Z" />
				<path d="M10.8 3h3.6v1.5h-2.2v2h-1.4Z" />
				<path d="M19 13.2h1.6v1.6H19Zm1.4-2.7H22v7h-1.6Z" />
			</g>
		{/if}
	</svg>
{/if}
