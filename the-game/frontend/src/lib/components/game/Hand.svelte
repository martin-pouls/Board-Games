<script lang="ts">
	import { commandInfo } from '#lib/game/commands.js';
	import { cn } from '#lib/utils.js';
	import type { HandCard } from '#lib/types.js';
	import CommandIcon from './CommandIcon.svelte';

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
			{@const info = c.command ? commandInfo(c.command) : undefined}
			<li class="min-w-0">
				<button
					type="button"
					disabled={!interactive}
					aria-pressed={interactive ? isSelected : undefined}
					aria-label={`Card ${c.value}${info ? `, ${info.label}` : ''}${noMove ? ', no legal pile' : ''}`}
					onclick={() => onselect?.(c)}
					class={cn(
						'flex aspect-[2/3] w-full flex-col items-center justify-center gap-0.5 rounded-lg border-2 bg-card text-lg font-bold text-card-foreground tabular-nums shadow-sm transition-all duration-150 select-none sm:text-xl',
						'[touch-action:manipulation] [-webkit-tap-highlight-color:transparent] disabled:cursor-default',
						info?.kind === 'lightning' && 'border-red-400 dark:border-red-500',
						info?.kind === 'infinity' && 'border-violet-400 dark:border-violet-500',
						!info && 'border-border',
						isSelected && '-translate-y-2 border-primary shadow-md ring-2 ring-primary',
						noMove && !isSelected && 'opacity-45'
					)}
				>
					<span class="leading-none">{c.value}</span>
					{#if c.command}
						<CommandIcon
							command={c.command}
							class={cn(
								'size-3.5 sm:size-4',
								info?.kind === 'lightning'
									? 'text-red-600 dark:text-red-400'
									: 'text-violet-600 dark:text-violet-400'
							)}
						/>
					{/if}
				</button>
			</li>
		{/each}
	</ul>
{/if}
