<script lang="ts">
	import Crown from '@lucide/svelte/icons/crown';
	import Drama from '@lucide/svelte/icons/drama';
	import Swords from '@lucide/svelte/icons/swords';
	import { cn } from '#lib/utils.js';
	import { cardName, isRed, rankLabel, suitGlyph } from '#lib/game/format.js';
	import type { Card } from '#lib/types.js';

	let {
		card,
		size = 'md',
		selected = false,
		dimmed = false,
		disabled = false,
		onclick
	}: {
		card: Card;
		/** 'fluid' fills its container (hand grid). */
		size?: 'xs' | 'sm' | 'md' | 'lg' | 'fluid';
		selected?: boolean;
		dimmed?: boolean;
		disabled?: boolean;
		onclick?: () => void;
	} = $props();

	const suit = $derived(card.suit === '' ? null : card.suit);
	const jester = $derived(suit === null || card.rank === 0);
	const glyph = $derived(suit ? suitGlyph(suit) + '︎' : '');
	const label = $derived(rankLabel(card.rank));
	const royal = $derived(card.rank >= 11);
	const red = $derived(isRed(card.suit));
	const name = $derived(cardName(card));

	const dims: Record<string, string> = {
		xs: 'w-8 text-xs',
		sm: 'w-11 text-sm',
		md: 'w-14 text-base',
		lg: 'w-19 text-lg',
		fluid: 'w-full text-sm'
	};
	const centre: Record<string, string> = {
		xs: 'text-base',
		sm: 'text-xl',
		md: 'text-3xl',
		lg: 'text-5xl',
		fluid: 'text-2xl'
	};

	const classes = $derived(
		cn(
			'relative block aspect-[5/7] shrink-0 overflow-hidden rounded-lg border border-ink/70 ring-1 ring-gold/40 select-none ring-inset',
			dims[size],
			jester ? 'bg-jester/15 text-jester' : 'bg-parchment',
			!jester && (red ? 'text-suit-red' : 'text-suit-black'),
			selected && '-translate-y-3 ring-2 ring-gold',
			dimmed && 'opacity-50',
			onclick && !disabled && 'cursor-pointer transition-transform',
			disabled && 'cursor-not-allowed'
		)
	);
</script>

{#snippet face()}
	{#if jester}
		<span class="absolute inset-0 flex flex-col items-center justify-center gap-0.5">
			<Drama class={size === 'xs' ? 'size-3' : size === 'lg' ? 'size-8' : 'size-5'} />
			{#if size !== 'xs'}
				<span class="font-display text-[0.6em] font-bold tracking-wide">Jester</span>
			{/if}
		</span>
	{:else}
		<span
			class="absolute top-0.5 left-1 flex flex-col items-center leading-none"
			style="font-family: system-ui, sans-serif"
		>
			<span class="font-display font-bold">{label}</span>
			<span>{glyph}</span>
		</span>
		{#if size !== 'xs'}
			<span class="absolute inset-0 flex items-center justify-center">
				{#if royal}
					{@const Icon = card.rank === 11 ? Swords : Crown}
					<Icon class="absolute size-2/3 text-gold opacity-30" />
					<span class="font-display font-bold {centre[size]}">{label}</span>
				{:else}
					<span class={centre[size]} style="font-family: system-ui, sans-serif">{glyph}</span>
				{/if}
			</span>
		{/if}
	{/if}
{/snippet}

{#if onclick}
	<button
		type="button"
		class={classes}
		aria-label={name}
		aria-pressed={selected}
		{disabled}
		{onclick}
	>
		{@render face()}
	</button>
{:else}
	<div class={classes} role="img" aria-label={name}>
		{@render face()}
	</div>
{/if}
