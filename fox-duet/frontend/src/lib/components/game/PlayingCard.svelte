<script lang="ts">
	import { abilityName, cardLabel } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import type { Card } from '#lib/types.js';
	import AbilityIcon from './AbilityIcon.svelte';
	import PawIcon from './PawIcon.svelte';
	import SuitIcon from './SuitIcon.svelte';

	let {
		card,
		size = 'md',
		dim = false,
		class: className = ''
	}: {
		card: Card;
		/** xs: results list (40 px); decree (54 px); sm: hand (62 px); md: trick (84 px, 90 px from a 390 px wide screen). */
		size?: 'xs' | 'decree' | 'sm' | 'md';
		/** Can't be picked now: the ink fades, the paper stays opaque (cards overlap in the hand). */
		dim?: boolean;
		class?: string;
	} = $props();

	const ability = $derived(card.ability || null);
	const label = $derived(
		`${cardLabel(card)}, moves ${card.movement}${ability ? `, ${abilityName(card)}` : ''}`
	);
	const paws = $derived(Array.from({ length: card.movement }, (_, i) => i));
</script>

<!--
	The index sits in the left 36 %, so it stays readable where cards overlap in the hand: a
	suit-coloured tab with rank and suit, then the paws (movement) and, on small cards, the
	ability. Odd ranks have an oval with the ability picture that runs behind the tab; table cards
	name the ability on a ribbon.
-->
<div
	class={cn('card', size, dim && 'dim', className)}
	style:--suit="var(--{card.suit})"
	role="img"
	aria-label={label}
	title={label}
>
	{#if ability}<i class="oval"></i>{/if}
	<i class="tab"></i>
	<div class="idx">
		<span class="rank">{card.rank}</span>
		<SuitIcon suit={card.suit} class="suit" />
	</div>
	{#if size !== 'xs'}
		<div class="below">
			{#each paws as i (i)}<PawIcon class="paw" />{/each}
			{#if ability && size !== 'md'}<AbilityIcon {ability} class="mini" />{/if}
		</div>
	{/if}
	{#if ability}
		<AbilityIcon {ability} class="art" />
	{:else}
		<SuitIcon suit={card.suit} class="art pip" />
	{/if}
	{#if ability && size === 'md'}
		<svg class="ribbon" viewBox="0 0 100 20" preserveAspectRatio="none" aria-hidden="true">
			<path d="M0 2H100L94 10 100 18H0L6 10Z" />
		</svg>
		<span class="name">{abilityName(card)}</span>
	{/if}
</div>

<style>
	.card {
		--w: 84px;
		position: relative;
		flex: none;
		width: var(--w);
		height: calc(var(--w) * 1.4);
		overflow: hidden;
		border-radius: calc(var(--w) * 0.09);
		background: var(--parchment);
		color: var(--suit);
		box-shadow:
			0 0 0 1px color-mix(in oklch, var(--ink) 14%, transparent),
			0 1px 3px color-mix(in oklch, var(--ink) 22%, transparent);
		user-select: none;
	}
	.dim {
		filter: saturate(0.35) brightness(0.9);
	}
	.dim > :global(*) {
		opacity: 0.45;
	}
	.xs {
		--w: 40px;
	}
	.decree {
		--w: 54px;
	}
	.sm {
		--w: 62px;
	}
	@media (min-width: 390px) {
		.md {
			--w: 90px;
		}
	}
	.oval {
		position: absolute;
		inset: 13% 4% 22% 20%;
		border-radius: 50%;
		background: color-mix(in oklch, var(--suit) 13%, var(--parchment));
	}
	/* The tab runs over the oval; a rim of paper separates them. */
	.tab {
		position: absolute;
		top: 0;
		left: 0;
		z-index: 1;
		width: 36%;
		height: calc(var(--w) * 0.64);
		border-bottom-right-radius: calc(var(--w) * 0.14);
		background: var(--suit);
		box-shadow: 0 0 0 calc(var(--w) * 0.03) var(--parchment);
	}
	.idx,
	.below {
		position: absolute;
		left: 0;
		z-index: 2;
		display: flex;
		width: 36%;
		flex-direction: column;
		align-items: center;
	}
	.idx {
		top: calc(var(--w) * 0.05);
		gap: calc(var(--w) * 0.02);
		color: var(--parchment);
		line-height: 0.9;
	}
	.rank {
		font-family: var(--font-display);
		font-size: calc(var(--w) * 0.344);
		font-weight: 700;
		letter-spacing: -0.02em;
	}
	.idx :global(.suit) {
		width: calc(var(--w) * 0.2);
		height: calc(var(--w) * 0.2);
	}
	/* Below the tab and its rim, with a little air. */
	.below {
		top: calc(var(--w) * 0.72);
		gap: calc(var(--w) * 0.015);
	}
	.below :global(.paw) {
		width: calc(var(--w) * 0.15);
		height: calc(var(--w) * 0.15);
	}
	.below :global(.mini) {
		width: calc(var(--w) * 0.19);
		height: calc(var(--w) * 0.19);
		margin-top: calc(var(--w) * 0.015);
	}
	.card :global(.art) {
		position: absolute;
		top: 49%;
		left: 63%;
		width: calc(var(--w) * 0.44);
		height: calc(var(--w) * 0.44);
		translate: -50% -50%;
	}
	.card :global(.art.pip) {
		width: calc(var(--w) * 0.5);
		height: calc(var(--w) * 0.5);
		opacity: 0.2;
	}
	.ribbon {
		position: absolute;
		right: 4%;
		bottom: calc(var(--w) * 0.12);
		left: 4%;
		width: 92%;
		height: 18px;
		fill: var(--ribbon);
	}
	.name {
		position: absolute;
		right: 0;
		bottom: calc(var(--w) * 0.12);
		left: 0;
		color: var(--ink);
		font-size: 0.75rem;
		font-weight: 800;
		line-height: 18px;
		text-align: center;
		white-space: nowrap;
	}
</style>
