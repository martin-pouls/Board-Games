<script lang="ts">
	import { CARD_SIZE, type CardSize } from '#lib/game/cards.js';
	import { cn } from '#lib/utils.js';

	let {
		size = 'md',
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
</div>
