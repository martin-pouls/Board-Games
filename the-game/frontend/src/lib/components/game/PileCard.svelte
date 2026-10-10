<script lang="ts">
	import ArrowUp from '@lucide/svelte/icons/arrow-up';
	import ArrowDown from '@lucide/svelte/icons/arrow-down';
	import { scale } from 'svelte/transition';
	import { prefersReducedMotion } from 'svelte/motion';
	import { cn } from '#lib/utils.js';
	import type { Pile } from '#lib/types.js';

	let {
		pile,
		highlight = false,
		dimmed = false,
		disabled = false,
		playsThisTurn = 0,
		onclick
	}: {
		pile: Pile;
		highlight?: boolean;
		dimmed?: boolean;
		disabled?: boolean;
		playsThisTurn?: number;
		onclick?: () => void;
	} = $props();

	const up = $derived(pile.direction === 'up');
	const empty = $derived(pile.topCard === null);
	const label = $derived(
		`${up ? 'Ascending' : 'Descending'} pile, ${empty ? 'empty' : `top card ${pile.topValue}`}` +
			(highlight ? ', tap to play here' : '')
	);
</script>

<button
	type="button"
	{disabled}
	{onclick}
	aria-label={label}
	class={cn(
		'[container-type:size] relative flex h-full min-h-24 w-full [touch-action:manipulation] flex-col rounded-2xl border-2 p-2 text-left shadow-sm transition-all duration-150 select-none [-webkit-tap-highlight-color:transparent]',
		up
			? 'border-sky-300 bg-sky-50 text-sky-900 dark:border-sky-800 dark:bg-sky-950/50 dark:text-sky-100'
			: 'border-amber-300 bg-amber-50 text-amber-900 dark:border-amber-800 dark:bg-amber-950/50 dark:text-amber-100',
		highlight &&
			'border-emerald-500 shadow-lg ring-[3px] ring-emerald-500/60 dark:border-emerald-400 dark:ring-emerald-400/60',
		dimmed && 'opacity-35 saturate-50',
		disabled && !dimmed && 'cursor-default'
	)}
>
	<span class="flex items-center justify-between text-xs font-semibold">
		<span class="flex items-center gap-0.5">
			{#if up}<ArrowUp class="size-4" />{:else}<ArrowDown class="size-4" />{/if}
			<span class="tabular-nums">{up ? '1 → 99' : '100 → 2'}</span>
		</span>
		<span class="tabular-nums opacity-70" title="Cards on this pile">{pile.cardCount}</span>
	</span>

	<span class="flex flex-1 items-center justify-center">
		{#key pile.topValue}
			<span
				in:scale={{ start: 0.6, duration: prefersReducedMotion.current ? 0 : 220 }}
				class={cn(
					'leading-none font-bold tracking-tight tabular-nums',
					up ? 'text-sky-700 dark:text-sky-300' : 'text-amber-700 dark:text-amber-300',
					empty && 'opacity-40'
				)}
				style="font-size: clamp(1.75rem, min(34cqw, 48cqh), 5.5rem)"
			>
				{pile.topValue}
			</span>
		{/key}
	</span>

	<span class="h-4 text-center text-[0.7rem] font-medium opacity-70">
		{#if playsThisTurn > 0}+{playsThisTurn} this turn{/if}
	</span>
</button>
