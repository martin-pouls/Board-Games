<script lang="ts">
	import { cardKey } from '#lib/game/rules.js';
	import type { Card } from '#lib/types.js';
	import PlayingCard from './PlayingCard.svelte';

	let {
		cards,
		selected,
		interactive,
		ontoggle,
		hint
	}: {
		cards: Card[];
		/** Indices into `cards`. */
		selected: number[];
		interactive: boolean;
		ontoggle: (index: number) => void;
		hint?: (index: number) => 'dim' | undefined;
	} = $props();
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
			{#each cards as c, i (`${cardKey(c)}#${i}`)}
				<PlayingCard
					card={c}
					size="fluid"
					selected={selected.includes(i)}
					dimmed={hint?.(i) === 'dim' && !selected.includes(i)}
					disabled={!interactive}
					onclick={() => ontoggle(i)}
				/>
			{/each}
		</div>
	{/if}
</div>
