<script lang="ts">
	import PawPrint from '@lucide/svelte/icons/paw-print';
	import { abilityName, cardLabel } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import type { Card } from '#lib/types.js';
	import SuitIcon from './SuitIcon.svelte';

	let {
		card,
		size = 'md',
		class: className = ''
	}: {
		card: Card;
		/** xs: results list; sm: hand; md: table. */
		size?: 'xs' | 'sm' | 'md';
		class?: string;
	} = $props();

	const COLOR = { doves: 'text-doves', roses: 'text-roses', stars: 'text-stars' } as const;
	const SIZE = {
		xs: 'h-14 w-10 rounded-md p-0.5',
		sm: 'h-[5.5rem] w-[3.9rem] rounded-lg p-1',
		md: 'h-24 w-[4.25rem] rounded-lg p-1'
	} as const;
	const ability = $derived(abilityName(card));
	const label = $derived(
		`${cardLabel(card)}, moves ${card.movement}${ability ? `, ${ability}` : ''}`
	);
</script>

<div
	class={cn(
		'relative flex shrink-0 flex-col border border-ink/15 bg-parchment shadow-sm select-none',
		COLOR[card.suit],
		SIZE[size],
		className
	)}
	role="img"
	aria-label={label}
	title={label}
>
	<!-- Left strip: stays visible where cards overlap in the hand. -->
	<div class="flex w-fit flex-col items-center leading-none">
		<span class={cn('font-display font-bold', size === 'xs' ? 'text-sm' : 'text-xl')}>
			{card.rank}
		</span>
		<SuitIcon suit={card.suit} class={size === 'xs' ? 'size-3' : 'size-4'} />
		{#if size !== 'xs'}
			<span class="mt-1 flex flex-col items-center gap-px text-ink/60">
				{#each Array.from({ length: card.movement }, (_, i) => i) as i (i)}
					<PawPrint class="size-2.5" />
				{/each}
			</span>
		{/if}
	</div>
	{#if size !== 'xs'}
		<SuitIcon
			suit={card.suit}
			class="pointer-events-none absolute top-1/2 left-1/2 size-8 -translate-x-[35%] -translate-y-1/2 opacity-25"
		/>
		{#if ability}
			<span
				class="absolute inset-x-0 bottom-1 truncate pl-3.5 text-center text-[0.6rem] leading-tight font-semibold text-ink/75"
			>
				{ability}
			</span>
		{/if}
	{/if}
</div>
