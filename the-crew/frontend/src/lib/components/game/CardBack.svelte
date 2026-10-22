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
		<!-- The sunken temple, lit in the dark: the back of every card. -->
		<svg viewBox="0 0 24 24" class="size-3/5 opacity-55" fill="currentColor" aria-hidden="true">
			<path d="M5.5 13a6.5 6.5 0 0 1 13 0z" />
			<rect x="11.4" y="4.2" width="1.2" height="2" rx="0.4" />
			<rect x="4" y="13" width="16" height="1.6" rx="0.4" />
			<path d="M5.6 15.4h1.4v4.6H5.6zm3.4 0h1.4v4.6H9zm3.4 0h1.4v4.6h-1.4zm3.4 0h1.4v4.6h-1.4z" />
			<rect x="3.5" y="20" width="17" height="1.4" rx="0.4" />
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
