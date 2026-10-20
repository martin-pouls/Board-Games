<script lang="ts">
	import type { Suit } from '#lib/types.js';

	let { suit, class: className = '' }: { suit: Suit; class?: string } = $props();

	const ROSE_PETALS = [0, 72, 144, 216, 288].map((a) => ({
		cx: 12 + 5.2 * Math.sin((a * Math.PI) / 180),
		cy: 12 - 5.2 * Math.cos((a * Math.PI) / 180)
	}));
</script>

<!-- The suit symbols as printed: a dove, a five-petal rose, a four-pointed star. -->
<svg viewBox="0 0 24 24" class={className} fill="currentColor" aria-hidden="true">
	{#if suit === 'stars'}
		<path d="M12 1.5 14.3 9.7 22.5 12 14.3 14.3 12 22.5 9.7 14.3 1.5 12 9.7 9.7Z" />
	{:else if suit === 'roses'}
		{#each ROSE_PETALS as p, i (i)}
			<circle cx={p.cx} cy={p.cy} r="4.4" opacity="0.75" />
		{/each}
		<circle cx="12" cy="12" r="3.6" />
	{:else}
		<path
			d="M2.5 12.5c3.5.4 6-1.6 7.3-5.6 1.5 1.8 2.2 3.8 2.1 6.2 3-2.6 6.4-3.3 10.1-2.2-2.3.8-3.6 2.2-4.4 4.2-1.7 4.3-6.5 6.3-11 4.6l-3.4 2.3 1.2-3.9C3.2 16.7 2.4 14.8 2.5 12.5Z"
		/>
	{/if}
</svg>
