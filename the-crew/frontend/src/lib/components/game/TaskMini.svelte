<!--
	A small card on a Deep Sea task (its face, its chip), drawn like the printed ones: filled with the
	body colour of its window, the value and symbol in white; "any colour" quartered in all four; a
	submarine deep blue with the yellow bathyscaphe. An orange "3×" badge sits left of it.
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
		/** chip: in a task chip (15 × 21). */
		size?: 'xs' | 'task' | 'chip';
		/** In a row of four or more. */ small?: boolean;
	} = $props();

	const trump = $derived(card.suit === 'trump');
</script>

{#if card.count}
	<!-- The "3×" badge sits left of the card, overlapping its corner a little, as printed. -->
	<span
		class={cn(
			'relative z-10 rounded-[0.15rem] bg-task-badge font-display leading-tight font-bold text-white shadow',
			size === 'chip'
				? '-mr-px px-0.5 text-[0.66rem] leading-[1.2]'
				: cn('-mr-1 self-start px-0.5', size === 'xs' ? 'text-[0.5rem]' : 'text-[0.65rem]')
		)}>{card.count}×</span
	>
{/if}
<span
	class={cn(
		'relative flex shrink-0 flex-col items-center justify-center shadow-sm ring-1 ring-white/75',
		size === 'chip'
			? 'h-[1.3125rem] w-[0.9375rem] rounded-[0.1875rem]'
			: size === 'xs'
				? 'h-[1.2rem] w-[0.85rem] rounded-[0.2rem]'
				: small
					? 'h-[1.6rem] w-[1.12rem] rounded-[0.2rem]'
					: 'h-[1.9rem] w-[1.35rem] rounded-[0.2rem]',
		card.suit === 'any' && 'task-any',
		trump ? 'text-(--nn-trump)' : 'text-white'
	)}
	style:background-color={card.suit === 'any'
		? undefined
		: trump
			? 'var(--nd-trump)'
			: `var(--ng-${card.suit})`}
>
	{#if card.rank}
		<span
			class={cn(
				'font-display leading-none font-extrabold',
				!trump && '[text-shadow:0_0_2px_rgb(0_0_0/55%),0_0_1px_rgb(0_0_0/60%)]',
				size === 'chip'
					? 'text-[0.72rem]'
					: size === 'xs'
						? 'text-[0.6rem]'
						: small
							? 'text-[0.8rem]'
							: 'text-[0.95rem]'
			)}>{card.rank}</span
		>
	{/if}
	{#if card.suit !== 'any'}
		<SuitGlyph
			suit={card.suit}
			class={cn(
				'drop-shadow-[0_0_1px_rgb(0_0_0/40%)]',
				card.rank
					? size === 'xs' || size === 'chip'
						? 'size-1.5'
						: 'size-2.5'
					: size === 'xs' || size === 'chip'
						? trump
							? 'size-3'
							: 'size-2.5'
						: 'size-4'
			)}
		/>
	{/if}
</span>
