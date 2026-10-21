<script lang="ts">
	import { flip } from 'svelte/animate';
	import NewCardMark from '@board-games/shared/NewCardMark.svelte';
	import { getCardMotion } from '@board-games/shared/cardMotion.js';
	import { trackNewCards } from '@board-games/shared/newCards.svelte.js';
	import { cardKey, sameCard } from '#lib/game/cards.js';
	import { cn } from '#lib/utils.js';
	import type { Card } from '#lib/types.js';
	import PlayingCard from './PlayingCard.svelte';

	let {
		cards,
		selected = null,
		marked = [],
		enabled = () => false,
		dealing = false,
		deal,
		label = 'Your hand',
		onpick
	}: {
		cards: Card[];
		selected?: Card | null;
		/** Cards set aside (the chosen pass card), shown with a ring. */
		marked?: Card[];
		/** Whether a card can be picked now; the others get a veil while any card is enabled. */
		enabled?: (c: Card) => boolean;
		/** A fresh deal: cards come in one after another. */
		dealing?: boolean;
		/** Identifies the deal; when it changes, no card counts as just received. */
		deal?: string;
		label?: string;
		onpick?: (c: Card) => void;
	} = $props();

	const motion = getCardMotion();
	const { send, receive } = motion;
	const CARD_WIDTH = 58;
	let width = $state(0);
	const fresh = trackNewCards(
		() => cards.map(cardKey),
		() => deal
	);

	const anyEnabled = $derived(cards.some((c) => enabled(c)));
	// Overlap the cards so the whole hand fits (14 cards at 343 px); never spread them more than a
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
			{@const isMarked = marked.some((m) => sameCard(card, m))}
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
				<!-- Only the picked card lifts. A card that can't be picked gets a veil in the table
				     colour, not transparency, so the cards under it never show through. -->
				<button
					type="button"
					class={cn(
						'relative block rounded-md transition-[translate] duration-150 focus-visible:ring-3 focus-visible:ring-ring focus-visible:outline-none',
						isSelected && '-translate-y-3 shadow-[0_3px_8px_var(--card-shadow)] ring-2 ring-signal',
						isMarked && !isSelected && 'ring-2 ring-signal/60'
					)}
					disabled={!on}
					aria-pressed={isSelected}
					onclick={() => onpick?.(card)}
				>
					<PlayingCard {card} size="hand" />
					{#if anyEnabled && !on && !isSelected}
						<span class="absolute inset-0 z-[5] rounded-md bg-background/58"></span>
					{/if}
					{#if fresh.has(cardKey(card))}
						<NewCardMark dot="corner-left" ring="var(--paper)" />
					{/if}
				</button>
			</li>
		{:else}
			<li class="flex h-[5.25rem] items-center text-sm text-muted-foreground">No cards in hand</li>
		{/each}
	</ul>
</div>
