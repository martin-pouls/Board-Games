<!--
	A small card on a Deep Sea task face, drawn like the printed ones: filled with its colour, the
	value and glyph in white; "any colour" quartered in all four; a submarine dark. An orange "3×"
	badge sits left of it.
-->
<script lang="ts">
	import { cn } from '#lib/utils.js';
	import type { Mini } from '#lib/game/taskFace.js';
	import SuitGlyph from './SuitGlyph.svelte';

	let {
		card,
		size = 'task',
		small = false
	}: {
		card: Mini;
		size?: 'xs' | 'task';
		/** In a row of four or more. */ small?: boolean;
	} = $props();

	const BG: Record<Mini['suit'], string> = {
		pink: 'bg-pink text-white',
		yellow: 'bg-yellow text-white',
		blue: 'bg-blue text-white',
		green: 'bg-green text-white',
		trump: 'bg-rocket text-rocket-ink',
		any: 'task-any text-white'
	};
</script>

{#if card.count}
	<!-- The "3×" badge sits left of the card, overlapping its corner a little, as printed. -->
	<span
		class={cn(
			'relative z-10 -mr-1 self-start rounded-[0.15rem] bg-task-badge px-0.5 font-display leading-tight font-bold text-white shadow',
			size === 'xs' ? 'text-[0.5rem]' : 'text-[0.65rem]'
		)}>{card.count}×</span
	>
{/if}
<span
	class={cn(
		'relative flex shrink-0 flex-col items-center justify-center rounded-[0.2rem] border border-white/70 shadow-sm',
		size === 'xs'
			? 'h-[1.2rem] w-[0.85rem]'
			: small
				? 'h-[1.6rem] w-[1.12rem]'
				: 'h-[1.9rem] w-[1.35rem]',
		BG[card.suit]
	)}
>
	{#if card.rank}
		<span
			class={cn(
				'font-display leading-none font-bold [text-shadow:0_0_2px_rgb(0_0_0/45%)]',
				size === 'xs' ? 'text-[0.6rem]' : small ? 'text-[0.8rem]' : 'text-[0.95rem]'
			)}>{card.rank}</span
		>
	{/if}
	{#if card.suit !== 'any'}
		<SuitGlyph
			suit={card.suit}
			class={card.rank
				? size === 'xs'
					? 'size-1.5'
					: 'size-2.5'
				: size === 'xs'
					? 'size-2.5'
					: 'size-4'}
		/>
	{/if}
</span>
