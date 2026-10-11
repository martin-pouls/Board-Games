<script lang="ts">
	import type { Suit } from '#lib/types.js';

	let { suit, class: className = '' }: { suit: Suit; class?: string } = $props();
</script>

<svg viewBox="0 0 24 24" class={className} fill="currentColor" aria-hidden="true">
	{#if suit === 'stars'}
		<polygon points="12,2 14.9,8.6 22,9.3 16.6,14 18.2,21 12,17.3 5.8,21 7.4,14 2,9.3 9.1,8.6" />
	{:else if suit === 'roses'}
		{#each [0, 72, 144, 216, 288] as a (a)}
			<circle
				cx={12 + 5 * Math.sin((a * Math.PI) / 180)}
				cy={11.5 - 5 * Math.cos((a * Math.PI) / 180)}
				r="4.3"
				opacity="0.8"
			/>
		{/each}
		<circle cx="12" cy="11.5" r="3.4" />
	{:else}
		<path
			d="M2.5 12.5c3.5.4 6-1.6 7.3-5.6 1.5 1.8 2.2 3.8 2.1 6.2 3-2.6 6.4-3.3 10.1-2.2-2.3.8-3.6 2.2-4.4 4.2-1.7 4.3-6.5 6.3-11 4.6l-3.4 2.3 1.2-3.9C3.2 16.7 2.4 14.8 2.5 12.5Z"
		/>
	{/if}
</svg>
