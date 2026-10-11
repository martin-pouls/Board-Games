<script lang="ts">
	import { CARD_SIZE, cardLabel, COMPACT, SUIT_TEXT, type CardSize } from '#lib/game/cards.js';
	import { cn } from '#lib/utils.js';
	import type { Card } from '#lib/types.js';
	import SuitGlyph from './SuitGlyph.svelte';

	let {
		card,
		size = 'md',
		class: className = ''
	}: {
		card: Card;
		size?: CardSize;
		class?: string;
	} = $props();

	const rocket = $derived(card.suit === 'trump');
	const label = $derived(cardLabel(card));
</script>

<div
	class={cn(
		'relative flex shrink-0 flex-col shadow-sm select-none',
		rocket ? 'border border-rocket-ink/30 bg-rocket' : 'border border-ink/15 bg-paper',
		SUIT_TEXT[card.suit],
		CARD_SIZE[size],
		className
	)}
	role="img"
	aria-label={label}
	title={label}
>
	{#if COMPACT.has(size)}
		<span
			class={cn(
				'text-center font-display leading-tight font-bold',
				size === 'xs' ? 'text-sm' : 'text-lg'
			)}>{card.rank}</span
		>
		<SuitGlyph suit={card.suit} class={cn('mx-auto', size === 'xs' ? 'size-3' : 'size-4.5')} />
	{:else}
		<!-- Left strip: stays visible where cards overlap in the hand. -->
		<div class="flex w-fit flex-col items-center leading-none">
			<span class="font-display text-xl font-bold">{card.rank}</span>
			<SuitGlyph suit={card.suit} class="size-3.5" />
		</div>
		<SuitGlyph
			suit={card.suit}
			class={cn(
				'pointer-events-none absolute top-1/2 left-1/2 -translate-x-[30%] -translate-y-[40%]',
				size === 'md' ? 'size-8' : 'size-7',
				rocket ? 'opacity-80' : 'opacity-30'
			)}
		/>
		<span
			class="absolute right-1 bottom-0.5 rotate-180 font-display text-xs leading-none font-bold opacity-70"
			aria-hidden="true">{card.rank}</span
		>
	{/if}
</div>
