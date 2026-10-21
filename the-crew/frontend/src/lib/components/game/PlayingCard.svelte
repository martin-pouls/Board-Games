<!--
	A playing card. Planet Nine: a white card with the value and its symbol once, in the top-left
	corner, and Planet Nine's night print in a window beside them (the neon symbol on the colour's
	dark glow; rockets: a white rocket among stars). The small sizes (tasks, a communicated card)
	have no window. Mission Deep Sea keeps its own face until its theme is redone: the value and
	glyph in the corner, a large faint glyph and the value upside down at the bottom.
-->
<script lang="ts">
	import { edition } from '#lib/edition.js';
	import { CARD_SIZE, cardLabel, COMPACT, SUIT_TEXT, type CardSize } from '#lib/game/cards.js';
	import { cn } from '#lib/utils.js';
	import type { Card } from '#lib/types.js';
	import SuitGlyph from './SuitGlyph.svelte';

	let {
		card,
		size = 'trick',
		class: className = ''
	}: {
		card: Card;
		size?: CardSize;
		class?: string;
	} = $props();

	const rocket = $derived(card.suit === 'trump');
	const label = $derived(cardLabel(card));

	/** The corner of each size: value size, corner width, symbol top and size (px). */
	const CORNER: Record<CardSize, { vs: number; ix: number; top: number; sym: number }> = {
		xs: { vs: 14, ix: 26, top: 18, sym: 10 },
		comm: { vs: 16, ix: 30, top: 21, sym: 11 },
		task: { vs: 20, ix: 40, top: 26, sym: 14 },
		stack: { vs: 16, ix: 14, top: 20, sym: 10 },
		trick: { vs: 18, ix: 16, top: 22, sym: 11 },
		hand: { vs: 19, ix: 17, top: 23, sym: 12 }
	};
	const corner = $derived(CORNER[size]);
	const windowed = $derived(size === 'stack' || size === 'trick' || size === 'hand');
</script>

{#if edition !== 'deepSea'}
	<div
		class={cn('face relative shrink-0 overflow-hidden select-none', CARD_SIZE[size], className)}
		class:rocket
		style:--sc={rocket ? 'var(--tok)' : `var(--${card.suit})`}
		style:--nd={`var(--nd-${card.suit})`}
		style:--ng={`var(--ng-${card.suit})`}
		style:--nn={`var(--nn-${card.suit})`}
		style:--vs={`${corner.vs}px`}
		style:--ix={`${corner.ix}px`}
		style:--sym={`${corner.sym}px`}
		role="img"
		aria-label={label}
		title={label}
	>
		{#if windowed}
			<span class="art"><SuitGlyph suit={card.suit} class="art-glyph" /></span>
		{/if}
		<span class="val font-display">{card.rank}</span>
		<span class="six" style:top={`${corner.top}px`}>
			<SuitGlyph suit={card.suit} class="block" />
		</span>
	</div>
{:else}
	<div
		class={cn(
			'relative flex shrink-0 flex-col shadow-sm select-none',
			rocket ? 'border border-rocket-ink/30 bg-rocket' : 'border border-ink/15 bg-paper',
			SUIT_TEXT[card.suit],
			CARD_SIZE[size],
			size === 'xs' ? 'p-px' : COMPACT.has(size) ? 'p-0.5' : 'p-1',
			className
		)}
		role="img"
		aria-label={label}
		title={label}
	>
		{#if COMPACT.has(size)}
			<span
				class={cn(
					'text-center font-display leading-tight font-bold',
					size === 'xs' ? 'text-sm' : 'text-lg'
				)}>{card.rank}</span
			>
			<SuitGlyph suit={card.suit} class={cn('mx-auto', size === 'xs' ? 'size-3' : 'size-4.5')} />
		{:else}
			<!-- Left strip: stays visible where cards overlap in the hand. -->
			<div class="flex w-fit flex-col items-center leading-none">
				<span class="font-display text-xl font-bold">{card.rank}</span>
				<SuitGlyph suit={card.suit} class="size-3.5" />
			</div>
			<SuitGlyph
				suit={card.suit}
				class={cn(
					'pointer-events-none absolute top-1/2 left-1/2 -translate-x-[30%] -translate-y-[40%]',
					size === 'hand' ? 'size-8' : 'size-7',
					rocket ? 'opacity-80' : 'opacity-30'
				)}
			/>
			<span
				class="absolute right-1 bottom-0.5 rotate-180 font-display text-xs leading-none font-bold opacity-70"
				aria-hidden="true">{card.rank}</span
			>
		{/if}
	</div>
{/if}

<style>
	.face {
		background: var(--paper);
		box-shadow:
			0 0 0 1px var(--card-edge),
			0 1px 2px var(--card-shadow);
		color: var(--sc);
	}
	.val {
		position: absolute;
		left: 0;
		top: 2px;
		z-index: 1;
		width: var(--ix);
		text-align: center;
		font-size: var(--vs);
		font-weight: 750;
		line-height: 1;
	}
	.six {
		position: absolute;
		left: 0;
		z-index: 1;
		display: grid;
		place-items: center;
		width: var(--ix);
	}
	.six :global(svg) {
		width: var(--sym);
		height: var(--sym);
	}
	/* The night print: the colour's glow fading into its dark, the symbol in neon. */
	.art {
		position: absolute;
		inset: 3px 3px 3px var(--ix);
		display: grid;
		place-items: center;
		overflow: hidden;
		border-radius: 3px;
		background: radial-gradient(110% 75% at 60% 38%, var(--ng), var(--nd) 72%);
		color: var(--nn);
	}
	.art :global(.art-glyph) {
		width: 60%;
		height: 60%;
		opacity: 0.75;
		filter: drop-shadow(0 0 2px var(--nn));
	}
	.rocket .art {
		background:
			radial-gradient(1px 1px at 30% 70%, var(--window-star), transparent),
			radial-gradient(1px 1px at 75% 20%, var(--window-star), transparent),
			radial-gradient(1px 1px at 55% 88%, var(--window-star), transparent),
			radial-gradient(110% 75% at 60% 38%, var(--ng), var(--nd) 72%);
	}
	.rocket .art :global(.art-glyph) {
		color: var(--rocket-glyph);
		opacity: 0.85;
		filter: none;
	}
</style>
