<script lang="ts">
	import { cardLabel, SUIT_TEXT } from '#lib/game/cards.js';
	import { cn } from '#lib/utils.js';
	import type { Card } from '#lib/types.js';
	import SuitGlyph from './SuitGlyph.svelte';

	let {
		card,
		size = 'md',
		class: className = ''
	}: {
		card: Card;
		/** xs: chips and lists; sm: hand and JARVIS; md: the trick. */
		size?: 'xs' | 'sm' | 'md';
		class?: string;
	} = $props();

	const SIZE = {
		xs: 'h-9 w-6.5 rounded-[0.3rem] p-px',
		sm: 'h-[4.75rem] w-[3.25rem] rounded-lg p-1',
		md: 'h-[5.25rem] w-[3.6rem] rounded-lg p-1'
	} as const;
	const rocket = $derived(card.suit === 'trump');
	const label = $derived(cardLabel(card));
</script>

<div
	class={cn(
		'relative flex shrink-0 flex-col shadow-sm select-none',
		rocket ? 'border border-rocket-ink/30 bg-rocket' : 'border border-ink/15 bg-paper',
		SUIT_TEXT[card.suit],
		SIZE[size],
		className
	)}
	role="img"
	aria-label={label}
	title={label}
>
	{#if size === 'xs'}
		<span class="text-center font-display text-sm leading-tight font-bold">{card.rank}</span>
		<SuitGlyph suit={card.suit} class="mx-auto size-3" />
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
