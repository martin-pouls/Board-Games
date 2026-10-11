<script lang="ts">
	import { flip } from 'svelte/animate';
	import { getCardMotion } from '@board-games/shared/cardMotion.js';
	import { cardKey, sameCard } from '#lib/game/cards.js';
	import { cn } from '#lib/utils.js';
	import type { Card } from '#lib/types.js';
	import PlayingCard from './PlayingCard.svelte';

	let {
		cards,
		selected = null,
		marked = null,
		enabled = () => false,
		dealing = false,
		label = 'Your hand',
		onpick
	}: {
		cards: Card[];
		selected?: Card | null;
		/** A card set aside (the chosen pass card, the communicated card), shown with a ring. */
		marked?: Card | null;
		/** Whether a card can be picked now; others are dimmed while any card is enabled. */
		enabled?: (c: Card) => boolean;
		/** A fresh deal: cards come in one after another. */
		dealing?: boolean;
		label?: string;
		onpick?: (c: Card) => void;
	} = $props();

	const motion = getCardMotion();
	const { send, receive } = motion;
	const CARD_WIDTH = 52;
	let width = $state(0);

	const anyEnabled = $derived(cards.some((c) => enabled(c)));
	// Overlap the cards so the whole hand fits (14 cards at 360 px); never spread them more than a
	// small gap.
	const step = $derived(
		cards.length > 1
			? Math.min(CARD_WIDTH + 4, Math.max(16, (width - CARD_WIDTH) / (cards.length - 1)))
			: CARD_WIDTH
	);
</script>

<div bind:clientWidth={width} class="w-full">
	<ul class="flex h-24 items-end justify-center pt-3" aria-label={label}>
		{#each cards as card, i (cardKey(card))}
			{@const on = enabled(card)}
			{@const isSelected = sameCard(card, selected)}
			{@const isMarked = sameCard(card, marked)}
			<li
				class="relative shrink-0"
				style:margin-left={i === 0 ? '0' : `${step - CARD_WIDTH}px`}
				style:z-index={i}
				animate:flip={motion.flip}
				in:receive={{
					key: cardKey(card),
					offstage: 'top',
					delay: dealing ? motion.dealDelay(i) : 0
				}}
				out:send={{ key: cardKey(card), offstage: 'top' }}
			>
				<button
					type="button"
					class={cn(
						'block rounded-lg transition-[translate,opacity,filter] duration-150 focus-visible:ring-3 focus-visible:ring-ring focus-visible:outline-none',
						isSelected && '-translate-y-3 ring-3 ring-signal',
						isMarked && !isSelected && 'ring-2 ring-signal/60',
						on && !isSelected && 'hover:-translate-y-1.5',
						anyEnabled && !on && !isSelected && 'opacity-40 saturate-50'
					)}
					disabled={!on}
					aria-pressed={isSelected}
					onclick={() => onpick?.(card)}
				>
					<PlayingCard {card} size="sm" />
				</button>
			</li>
		{:else}
			<li class="flex h-[4.75rem] items-center text-sm text-muted-foreground">No cards in hand</li>
		{/each}
	</ul>
</div>
