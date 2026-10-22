<!--
	A playing card: white, with the value and its symbol once in the top-left corner and the art in a
	window beside them. Planet Nine: its night print (the neon symbol on the colour's dark glow;
	rockets: a white rocket among stars). Mission Deep Sea: the bright duotone of its printed art,
	lit from the top left with the symbol in white (submarines: the yellow bathyscaphe in deep blue,
	the value alone in the corner). The small sizes (tasks, a communicated card) have no window.
-->
<script lang="ts">
	import { edition } from '#lib/edition.js';
	import { CARD_SIZE, cardLabel, type CardSize } from '#lib/game/cards.js';
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

	const deep = edition === 'deepSea';
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

<div
	class={cn('face relative shrink-0 overflow-hidden select-none', CARD_SIZE[size], className)}
	class:rocket
	class:deep
	style:--sc={rocket ? 'var(--tok)' : `var(--${card.suit})`}
	style:--nh={`var(--nh-${card.suit})`}
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
	{#if !(deep && rocket && windowed)}
		<span class="six" style:top={`${corner.top}px`}>
			<SuitGlyph suit={card.suit} class="block" />
		</span>
	{/if}
</div>

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
	/* Deep Sea's bright window: the colour's highlight, body and shadow, a sheen across it. */
	.deep .art {
		background:
			linear-gradient(118deg, transparent 36%, oklch(1 0 0 / 20%) 45%, transparent 53%),
			radial-gradient(120% 95% at 30% 18%, var(--nh), var(--ng) 52%, var(--nd));
	}
	.deep .art :global(.art-glyph) {
		opacity: 0.92;
		filter: drop-shadow(0 1px 1.5px oklch(0 0 0 / 35%));
	}
	.deep.rocket .art {
		background:
			linear-gradient(118deg, transparent 30%, oklch(1 0 0 / 12%) 42%, transparent 54%),
			radial-gradient(120% 95% at 30% 18%, var(--nh), var(--ng) 50%, var(--nd));
	}
	.deep.rocket .art :global(.art-glyph) {
		width: 72%;
		height: 72%;
		color: var(--nn);
		opacity: 1;
		filter: drop-shadow(0 0 3px color-mix(in oklab, var(--nn) 45%, transparent));
	}
</style>
