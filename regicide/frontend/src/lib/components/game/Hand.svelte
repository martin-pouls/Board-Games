<script lang="ts">
	import { flip } from 'svelte/animate';
	import NewCardMark from '@board-games/shared/NewCardMark.svelte';
	import { getCardMotion } from '@board-games/shared/cardMotion.js';
	import { trackNewCards } from '@board-games/shared/newCards.svelte.js';
	import { handKeys } from '#lib/game/rules.js';
	import type { Card } from '#lib/types.js';
	import PlayingCard from './PlayingCard.svelte';

	let {
		cards,
		selected,
		interactive,
		deal,
		ontoggle,
		hint
	}: {
		cards: Card[];
		/** Indices into `cards`. */
		selected: number[];
		interactive: boolean;
		/** Identifies the deal; when it changes, no card counts as just drawn. */
		deal?: string;
		ontoggle: (index: number) => void;
		hint?: (index: number) => 'dim' | undefined;
	} = $props();

	// Played and discarded cards fly to the last turn's piles (same keys there), drawn cards
	// come in from the Tavern deck above, one after another.
	const motion = getCardMotion();
	const { send, receive } = motion;

	const keys = $derived(handKeys(cards));
	// Each card carries its own key: a leaving card must not read its key from the new hand.
	const items = $derived(cards.map((card, i) => ({ card, key: keys[i] })));
	const fresh = trackNewCards(
		() => keys,
		() => deal
	);

	// The keys shown before the latest update (bookkeeping, not rendered), so the cards that just
	// came in can be staggered in hand order.
	let shown: string[] = [];
	const arrived = $derived(keys.filter((k) => !shown.includes(k)));
	$effect(() => {
		shown = keys;
	});
	const arrivalDelay = (key: string) => motion.dealDelay(Math.max(0, arrived.indexOf(key)));

	// A hand swapped for a new one (the solo Jester power) is a new deal: the old row goes at once
	// and the new one comes in as a whole. Flying a full hand out while another flies in squeezes
	// both into one row and looks messy.
	let previous: string[] = [];
	let deals = 0;
	const dealNo = $derived.by(() => {
		if (previous.length > 0 && keys.length > 0 && !keys.some((k) => previous.includes(k))) deals++;
		previous = keys;
		return deals;
	});
</script>

<div class="flex flex-col">
	<!--
		Always rendered, so the last cards to leave the hand still fly. One row of cards that keep
		their size, like The Game's: side by side while they fit, then each card after the first is
		pulled left over the one before, just enough for the row to fit the width. Leaving cards are
		taken out of the flow by animate:flip.
	-->
	{#key dealNo}
		<div
			class="hand flex justify-center pt-3"
			style:--n={items.length}
			in:receive={{ key: `hand-${dealNo}`, offstage: 'top' }}
		>
			{#each items as { card, key }, i (key)}
				{@const isSelected = selected.includes(i)}
				<div
					class="relative shrink-0 rounded-[6.5px]"
					animate:flip={motion.flip}
					in:receive={{ key, offstage: 'top', delay: arrivalDelay(key) }}
					out:send={{ key, offstage: 'top' }}
				>
					<PlayingCard
						{card}
						selected={isSelected}
						dimmed={hint?.(i) === 'dim' && !isSelected}
						disabled={!interactive}
						onclick={() => ontoggle(i)}
					/>
					{#if fresh.has(key)}
						<NewCardMark
							dot="corner-left"
							ring="var(--stock)"
							class={isSelected ? '-translate-y-3 transition-transform' : 'transition-transform'}
						/>
					{/if}
				</div>
			{/each}
		</div>
	{/key}
	{#if cards.length === 0}
		<p class="flex h-[100.8px] items-center justify-center text-sm text-muted-foreground">
			No cards in hand.
		</p>
	{/if}
</div>

<style>
	.hand {
		container-type: inline-size;
	}
	.hand > :not(:first-child) {
		margin-left: calc(min(76px, (100cqw - 72px) / max(var(--n) - 1, 1)) - 72px);
	}
</style>
