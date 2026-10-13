<script lang="ts">
	import { flip } from 'svelte/animate';
	import NewCardMark from '@board-games/shared/NewCardMark.svelte';
	import { getCardMotion } from '@board-games/shared/cardMotion.js';
	import { trackNewCards } from '@board-games/shared/newCards.svelte.js';
	import { cardKey, sameCard } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import type { Card } from '#lib/types.js';
	import PlayingCard from './PlayingCard.svelte';

	let {
		cards,
		selected = null,
		marked = null,
		enabled = () => false,
		dealing = false,
		deal,
		onpick
	}: {
		cards: Card[];
		selected?: Card | null;
		/** A card set aside (the chosen gift), shown lifted and outlined. */
		marked?: Card | null;
		/** Whether a card can be picked now; others are dimmed while any card is enabled. */
		enabled?: (c: Card) => boolean;
		/** A fresh deal: cards come in one after another. */
		dealing?: boolean;
		/** Identifies the deal; when it changes, no card counts as just received. */
		deal?: string;
		onpick?: (c: Card) => void;
	} = $props();

	const motion = getCardMotion();
	const { send, receive } = motion;
	const CARD_WIDTH = 62;
	let width = $state(0);
	const fresh = trackNewCards(
		() => cards.map(cardKey),
		() => deal
	);

	const anyEnabled = $derived(cards.some((c) => enabled(c)));
	// Overlap the cards so the whole hand fits; never spread them more than a small gap.
	const step = $derived(
		cards.length > 1
			? Math.min(CARD_WIDTH + 6, Math.max(18, (width - CARD_WIDTH) / (cards.length - 1)))
			: CARD_WIDTH
	);
</script>

<div bind:clientWidth={width} class="w-full">
	<ul class="flex h-28 items-end justify-center pt-3" aria-label="Your hand">
		{#each cards as card, i (cardKey(card))}
			{@const on = enabled(card)}
			{@const isSelected = sameCard(card, selected)}
			{@const isMarked = sameCard(card, marked)}
			<li
				class="relative shrink-0 transition-transform duration-150"
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
						'relative block rounded-lg transition-[translate,opacity,filter] duration-150 focus-visible:ring-3 focus-visible:ring-ring focus-visible:outline-none',
						(isSelected || isMarked) && '-translate-y-3',
						isSelected && 'ring-3 ring-fox',
						isMarked && 'ring-3 ring-moss',
						on && !isSelected && 'hover:-translate-y-1.5',
						anyEnabled && !on && !isMarked && 'opacity-45 saturate-50'
					)}
					disabled={!on}
					aria-pressed={isSelected}
					onclick={() => onpick?.(card)}
				>
					<PlayingCard {card} size="sm" />
					{#if fresh.has(cardKey(card))}
						<NewCardMark dot="corner-left" ring="var(--parchment)" />
					{/if}
				</button>
			</li>
		{:else}
			<li class="flex h-[5.5rem] items-center text-sm text-muted-foreground">No cards in hand</li>
		{/each}
	</ul>
</div>
