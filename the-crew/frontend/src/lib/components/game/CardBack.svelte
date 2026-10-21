<script lang="ts">
	import { edition } from '#lib/edition.js';
	import { CARD_SIZE, type CardSize } from '#lib/game/cards.js';
	import { cn } from '#lib/utils.js';

	let {
		size = 'trick',
		label = 'Hidden card',
		class: className = ''
	}: { size?: CardSize; label?: string; class?: string } = $props();
</script>

<div
	class={cn(
		'relative grid shrink-0 place-items-center overflow-hidden border border-white/15 bg-rocket text-rocket-ink shadow-sm select-none',
		CARD_SIZE[size],
		className
	)}
	role="img"
	aria-label={label}
	title={label}
>
	{#if edition === 'deepSea'}
		<!-- Sonar ripples: the back of every card. -->
		<svg viewBox="0 0 24 24" class="size-3/5 opacity-50" fill="none" aria-hidden="true">
			<circle cx="12" cy="12" r="2.4" fill="currentColor" />
			<circle cx="12" cy="12" r="6" stroke="currentColor" stroke-width="1.4" />
			<circle
				cx="12"
				cy="12"
				r="10"
				stroke="currentColor"
				stroke-width="1.4"
				stroke-dasharray="4 3"
			/>
		</svg>
	{:else}
		<!-- A small planet with a ring: the back of every card. -->
		<svg viewBox="0 0 24 24" class="size-3/5 opacity-50" fill="none" aria-hidden="true">
			<circle cx="12" cy="12" r="5" fill="currentColor" />
			<ellipse
				cx="12"
				cy="12"
				rx="10.5"
				ry="3.2"
				stroke="currentColor"
				stroke-width="1.4"
				transform="rotate(-20 12 12)"
			/>
		</svg>
	{/if}
</div>
