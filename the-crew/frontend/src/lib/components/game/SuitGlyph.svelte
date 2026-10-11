<script lang="ts">
	import type { Suit } from '#lib/types.js';

	// Each colour has its own shape, so the cards read without colour vision too: pink a ringed
	// planet, blue a wave, green a leaf, yellow a star; rockets a rocket.
	let { suit, class: className = '' }: { suit: Suit; class?: string } = $props();
</script>

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
		<path
			d="M12 1.5c3 2.3 4.6 5.8 4.6 9.7v5.3h-9.2v-5.3c0-3.9 1.6-7.4 4.6-9.7Zm0 6.3a1.9 1.9 0 1 0 0 3.8 1.9 1.9 0 0 0 0-3.8ZM7.4 13l-3 3.2v4.3l3-2.1Zm9.2 0 3 3.2v4.3l-3-2.1ZM10 17.5h4l-.8 4h-2.4Z"
		/>
	{/if}
</svg>
