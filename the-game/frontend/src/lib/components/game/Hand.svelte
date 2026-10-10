<script lang="ts">
	import { cn } from '#lib/utils.js';
	import type { HandCard } from '#lib/types.js';

	let {
		cards,
		selected = null,
		interactive = false,
		onselect
	}: {
		cards: HandCard[];
		selected?: number | null;
		interactive?: boolean;
		onselect?: (card: HandCard) => void;
	} = $props();

	const sorted = $derived([...cards].sort((a, b) => a.value - b.value));
</script>

{#if sorted.length === 0}
	<p class="py-4 text-center text-sm text-muted-foreground">No cards in hand.</p>
{:else}
	<ul
		class="grid justify-center gap-1 pt-2"
		style="grid-template-columns: repeat({sorted.length}, minmax(0, 3.5rem))"
		aria-label="Your hand"
	>
		{#each sorted as c (c.value)}
			{@const isSelected = selected === c.value}
			{@const noMove = interactive && c.legalPiles.length === 0}
			<li class="min-w-0">
				<button
					type="button"
					disabled={!interactive}
					aria-pressed={interactive ? isSelected : undefined}
					aria-label={`Card ${c.value}${noMove ? ', no legal pile' : ''}`}
					onclick={() => onselect?.(c)}
					class={cn(
						'flex aspect-[2/3] w-full items-center justify-center rounded-lg border-2 bg-card text-lg font-bold text-card-foreground tabular-nums shadow-sm transition-all duration-150 select-none sm:text-xl',
						'[touch-action:manipulation] [-webkit-tap-highlight-color:transparent] disabled:cursor-default',
						isSelected
							? '-translate-y-2 border-primary shadow-md ring-2 ring-primary'
							: 'border-border',
						noMove && !isSelected && 'opacity-45'
					)}
				>
					{c.value}
				</button>
			</li>
		{/each}
	</ul>
{/if}
