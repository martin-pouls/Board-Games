<script lang="ts">
	import NewCardMark from '@board-games/shared/NewCardMark.svelte';
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

	const keys = $derived(handKeys(cards));
	const fresh = trackNewCards(
		() => keys,
		() => deal
	);
</script>

<div class="flex flex-col">
	{#if cards.length === 0}
		<p class="flex h-20 items-center justify-center text-sm text-muted-foreground">
			No cards in hand.
		</p>
	{:else}
		<div
			class="grid justify-center gap-1 pt-3"
			style="grid-template-columns: repeat({cards.length}, minmax(0, 4.5rem))"
		>
			{#each cards as c, i (keys[i])}
				{@const isSelected = selected.includes(i)}
				<div class="relative rounded-lg">
					<PlayingCard
						card={c}
						size="fluid"
						selected={isSelected}
						dimmed={hint?.(i) === 'dim' && !isSelected}
						disabled={!interactive}
						onclick={() => ontoggle(i)}
					/>
					{#if fresh.has(keys[i])}
						<NewCardMark
							ring="var(--parchment)"
							class={isSelected ? '-translate-y-3 transition-transform' : 'transition-transform'}
						/>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>
