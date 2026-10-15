<script lang="ts">
	import ArrowUp from '@lucide/svelte/icons/arrow-up';
	import ArrowDown from '@lucide/svelte/icons/arrow-down';
	import Layers from '@lucide/svelte/icons/layers';
	import Skull from '@lucide/svelte/icons/skull';
	import { scale } from 'svelte/transition';
	import { prefersReducedMotion } from 'svelte/motion';
	import { commandInfo } from '#lib/game/commands.js';
	import { cn } from '#lib/utils.js';
	import type { Pile } from '#lib/types.js';
	import CommandChip from './CommandChip.svelte';
	import CommandIcon from './CommandIcon.svelte';

	let {
		pile,
		highlight = false,
		dimmed = false,
		disabled = false,
		playsThisTurn = 0,
		bound = false,
		skull = false,
		onclick
	}: {
		pile: Pile;
		highlight?: boolean;
		dimmed?: boolean;
		disabled?: boolean;
		playsThisTurn?: number;
		/** Single pile: every further card this turn must go here. */
		bound?: boolean;
		/** A skull played this turn shows here and must be covered. */
		skull?: boolean;
		onclick?: () => void;
	} = $props();

	const up = $derived(pile.direction === 'up');
	const empty = $derived(pile.topCard === null);
	const command = $derived(pile.topCard?.command);
	const label = $derived(
		`${up ? 'Ascending' : 'Descending'} pile, ${empty ? 'empty' : `top card ${pile.topValue}`}` +
			(command
				? `, ${commandInfo(command).label}${pile.commandActive ? ' (active)' : ' (no effect any more)'}`
				: '') +
			(skull ? ', skull must be covered' : '') +
			(bound ? ', all cards this turn go here' : '') +
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
		bound &&
			'outline-2 outline-offset-2 outline-emerald-500 outline-dashed dark:outline-emerald-400',
		skull && 'border-red-500 ring-[3px] ring-red-500/50 dark:border-red-400 dark:ring-red-400/50',
		highlight &&
			'border-emerald-500 shadow-lg ring-[3px] ring-emerald-500/60 dark:border-emerald-400 dark:ring-emerald-400/60',
		dimmed && 'opacity-35 saturate-50',
		disabled && !dimmed && 'cursor-default'
	)}
>
	{#if skull}
		<span
			class="absolute -top-2 left-1/2 flex -translate-x-1/2 items-center gap-0.5 rounded-full bg-red-600 px-1.5 py-0.5 text-[0.65rem] leading-none font-bold text-white shadow-sm motion-safe:animate-pulse"
			aria-hidden="true"
		>
			<Skull class="size-3" />
			Cover!
		</span>
	{/if}

	<span class="flex items-center justify-between gap-1 text-xs font-semibold">
		<span class="flex min-w-0 items-center gap-0.5">
			{#if up}<ArrowUp class="size-4 shrink-0" />{:else}<ArrowDown class="size-4 shrink-0" />{/if}
			<span class="truncate tabular-nums">{up ? '1 → 99' : '100 → 2'}</span>
			{#if bound}
				<CommandIcon
					command="singlePile"
					class="ml-0.5 size-4 text-emerald-600 dark:text-emerald-400"
				/>
			{/if}
		</span>
		<span
			class="flex shrink-0 items-center gap-0.5 tabular-nums opacity-70"
			title="Cards on this pile"
		>
			<Layers class="size-3.5" aria-hidden="true" />{pile.cardCount}
		</span>
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

	{#if command}
		<span class="flex h-5 min-w-0 items-center justify-between gap-1 text-[0.7rem] font-medium">
			<CommandChip {command} size="xs" muted={!pile.commandActive} class="min-w-0" />
			{#if playsThisTurn > 0}
				<span class="shrink-0 whitespace-nowrap tabular-nums opacity-70"
					>+{playsThisTurn}<span class="@max-[10.5rem]:hidden">&nbsp;this turn</span></span
				>
			{/if}
		</span>
	{:else}
		<span class="h-4 text-center text-[0.7rem] font-medium opacity-70">
			{#if playsThisTurn > 0}+{playsThisTurn} this turn{/if}
		</span>
	{/if}
</button>
