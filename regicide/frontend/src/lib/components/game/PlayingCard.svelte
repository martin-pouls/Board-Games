<!--
	A card face. Number cards: rank and suit in the top left corner, one big suit in the middle.
	Jacks, Queens and Kings: the rank letter in the middle, where the suit sits on a number card,
	standing on three misty ridges. The Jester: its cap, in the corner and in the middle. The
	smallest size shows the corner only. Sizes inside the face are in card widths (cqw), so every
	size keeps the same proportions.
-->
<script lang="ts">
	import { cn } from '#lib/utils.js';
	import { cardName, isRed, rankLabel, suitGlyph } from '#lib/game/format.js';
	import type { Card } from '#lib/types.js';
	import JesterHat from './JesterHat.svelte';

	let {
		card,
		size = 'md',
		selected = false,
		dimmed = false,
		disabled = false,
		onclick
	}: {
		card: Card;
		/** xs 34 px (stacks against the enemy), sm 48 px (last turn), md 72 px (hand), lg 84 px (enemy). */
		size?: 'xs' | 'sm' | 'md' | 'lg';
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

	const widths: Record<string, string> = { xs: 'w-8.5', sm: 'w-12', md: 'w-18', lg: 'w-21' };

	// The ridges in a 100 × 140 card, lightest (farthest) first. The letter's foot rests on the
	// first one; the darker ones stay below the tails of J and Q.
	const RIDGES = [
		'M0 100 L14 92 L26 96 L42 87 L56 93 L70 88 L84 94 L100 90 V140 H0Z',
		'M0 117 L18 109 L32 113 L48 105 L62 111 L76 106 L90 111 L100 108 V140 H0Z',
		'M0 129 L12 123 L28 127 L44 119 L58 125 L76 120 L100 126 V140 H0Z'
	];
	const RIDGE_FILLS = ['fill-ridge-1', 'fill-ridge-2', 'fill-ridge-3'];

	const classes = $derived(
		cn(
			'face relative block aspect-[5/7] shrink-0 overflow-hidden bg-stock shadow-[0_1px_3px_var(--card-shade)] ring-1 ring-stock-edge select-none',
			widths[size],
			jester ? 'jester' : red ? 'text-suit-red' : 'text-suit-black',
			selected && '-translate-y-3 ring-[2.5px] ring-primary',
			dimmed && 'opacity-50',
			onclick && !disabled && 'cursor-pointer transition-transform',
			disabled && 'cursor-not-allowed'
		)
	);
</script>

{#snippet face()}
	{#if jester}
		<JesterHat class={size === 'xs' ? 'absolute top-1 left-[3px] size-4' : 'jester-index'} />
		{#if size !== 'xs'}
			<JesterHat class="jester-big" />
		{/if}
	{:else}
		<span class={cn('index', size === 'xs' && 'index-xs')}>
			<b class="font-display">{label}</b>
			<i class="suit-glyph">{glyph}</i>
		</span>
		{#if size !== 'xs'}
			{#if royal}
				<svg
					class="absolute inset-0 size-full"
					viewBox="0 0 100 140"
					preserveAspectRatio="none"
					aria-hidden="true"
				>
					{#each RIDGES as d, i (i)}
						<path {d} class={RIDGE_FILLS[i]} />
					{/each}
				</svg>
				<span class="middle royal font-display font-bold">{label}</span>
			{:else}
				<span class="middle suit-glyph">{glyph}</span>
			{/if}
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

<style>
	/* cqw inside the face; the face's own corners can't use it (they would follow the parent). */
	/* A stacking context of its own, so the middle and the index never paint over a card on top. */
	.face {
		isolation: isolate;
		container-type: inline-size;
		border-radius: 9% / 6.43%;
	}
	.index {
		position: absolute;
		top: 4%;
		left: 6%;
		z-index: 2;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 6cqw;
	}
	.index b {
		font-size: 30cqw;
		font-weight: 700;
		line-height: 0.82;
		letter-spacing: -0.02em;
	}
	.index i {
		font-size: 21cqw;
		line-height: 1;
	}
	.index-xs b {
		font-size: 12px;
	}
	.index-xs i {
		font-size: 10px;
	}
	/* The suit of a number card and the letter of a royal share one box, so they line up. */
	.middle {
		position: absolute;
		inset: 0;
		z-index: 1;
		display: grid;
		place-items: center;
		padding-top: 12%;
		font-size: 46cqw;
		line-height: 1;
	}
	.middle.royal {
		font-size: 50cqw;
	}
	.jester {
		--j1: var(--jester-card);
		--j2: color-mix(in oklab, var(--jester-card) 68%, black);
		--j2-opacity: 1;
	}
	.face :global(.jester-index) {
		position: absolute;
		top: 5%;
		left: 6%;
		z-index: 2;
		width: 27cqw;
		height: 27cqw;
	}
	.face :global(.jester-big) {
		position: absolute;
		/* Centred by offsets, not a translate: a transformed SVG inside a rotated card renders blurry. */
		left: 20cqw;
		top: calc(50% - 24cqw);
		width: 60cqw;
		height: 60cqw;
	}
</style>
