<script lang="ts">
	import { flip } from 'svelte/animate';
	import NewCardMark from '@board-games/shared/NewCardMark.svelte';
	import { getCardMotion } from '@board-games/shared/cardMotion.js';
	import { trackNewCards } from '@board-games/shared/newCards.svelte.js';
	import { commandInfo } from '#lib/game/commands.js';
	import { cardKey } from '#lib/game/motion.js';
	import { cn } from '#lib/utils.js';
	import type { HandCard } from '#lib/types.js';
	import GameCard from './GameCard.svelte';

	let {
		cards,
		handSize,
		selected = null,
		interactive = false,
		deal,
		onselect
	}: {
		cards: HandCard[];
		/** The game's hand size: it sizes the cards, so they keep their size all game. */
		handSize: number;
		selected?: number | null;
		interactive?: boolean;
		/** Identifies the deal; when it changes, no card counts as just drawn. */
		deal?: string;
		onselect?: (card: HandCard) => void;
	} = $props();

	// Played cards fly to their pile (same keys there) and come back on undo; drawn cards come in
	// from the draw pile above, one after another.
	const motion = getCardMotion();
	const { send, receive } = motion;

	// Each card carries its own key: a leaving card must not read its key from the new hand.
	const items = $derived(
		[...cards].sort((a, b) => a.value - b.value).map((c) => ({ c, key: cardKey(c.value) }))
	);
	const keys = $derived(items.map((i) => i.key));
	const fresh = trackNewCards(
		() => keys,
		() => deal
	);

	// The keys shown before the latest update (bookkeeping, not rendered), so the cards that just
	// came in can be staggered in hand order.
	let shown: string[] = [];
	const arrived = $derived(keys.filter((k) => !shown.includes(k)));
	$effect(() => {
		shown = keys;
	});
	const arrivalDelay = (key: string) => motion.dealDelay(Math.max(0, arrived.indexOf(key)));

	/*
	 * One row. A full hand fits the width with the cards overlapping so that the corner value and
	 * the left half of the numeral stay visible (VISIBLE of a card); fewer cards spread out up to
	 * GAP apart. The card width depends only on the hand size, so the hand never changes height.
	 */
	const VISIBLE = 0.8;
	const GAP = 6;
	const MAX_W = 84;
	/** Room above the cards for the lifted selected card and the new-card dot. */
	const TOP = 12;
	let width = $state(0);
	const w = $derived(Math.min(MAX_W, Math.floor(width / (VISIBLE * (handSize - 1) + 1))));
	const step = $derived(items.length > 1 ? Math.min(w + GAP, (width - w) / (items.length - 1)) : 0);
	const offset = $derived((width - (w + step * (items.length - 1))) / 2);
</script>

<!--
	Always rendered, so the last card to leave the hand still flies. Cards are placed absolutely
	(leaving cards stay where they were while the rest close up via animate:flip).
-->
<ul
	bind:clientWidth={width}
	class="relative w-full shrink-0"
	style:height="{w * 1.4 + TOP}px"
	style:--w="{w}px"
	aria-label="Your hand"
>
	{#each items as { c, key }, i (key)}
		{@const isSelected = selected === c.value}
		{@const noMove = interactive && c.legalPiles.length === 0}
		{@const info = c.command ? commandInfo(c.command) : undefined}
		<li
			class="absolute"
			style:top="{TOP}px"
			style:left="{Math.round(offset + i * step)}px"
			animate:flip={motion.flip}
			in:receive={{ key, offstage: 'top', delay: arrivalDelay(key) }}
			out:send={{ key, offstage: 'top' }}
		>
			<button
				type="button"
				disabled={!interactive}
				aria-pressed={interactive ? isSelected : undefined}
				aria-label={`Card ${c.value}${info ? `, ${info.label}` : ''}${noMove ? ', no legal pile' : ''}`}
				onclick={() => onselect?.(c)}
				class={cn(
					'block rounded-[calc(var(--w)*0.1)] transition-[translate,opacity] duration-150 select-none',
					'[touch-action:manipulation] [-webkit-tap-highlight-color:transparent] disabled:cursor-default',
					isSelected &&
						'-translate-y-[calc(var(--w)*0.1)] outline-3 outline-offset-2 outline-primary',
					noMove && !isSelected && 'opacity-45'
				)}
			>
				<GameCard
					value={c.value}
					command={c.command}
					overlap={i > 0 && step < w}
					covered={i < items.length - 1 && step < w}
				>
					{#if fresh.has(key)}
						<NewCardMark dot="corner-left" ring="var(--card-face)" />
					{/if}
				</GameCard>
			</button>
		</li>
	{/each}
	{#if items.length === 0}
		<li class="absolute inset-0 grid place-items-center text-sm text-muted-foreground">
			No cards in hand.
		</li>
	{/if}
</ul>
