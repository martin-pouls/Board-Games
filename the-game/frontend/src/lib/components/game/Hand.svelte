<script lang="ts">
	import { flip } from 'svelte/animate';
	import NewCardMark from '@board-games/shared/NewCardMark.svelte';
	import { getCardMotion } from '@board-games/shared/cardMotion.js';
	import { trackNewCards } from '@board-games/shared/newCards.svelte.js';
	import { commandInfo } from '#lib/game/commands.js';
	import { cardKey } from '#lib/game/motion.js';
	import { cn } from '#lib/utils.js';
	import type { HandCard } from '#lib/types.js';
	import CommandIcon from './CommandIcon.svelte';

	let {
		cards,
		selected = null,
		interactive = false,
		deal,
		onselect
	}: {
		cards: HandCard[];
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
</script>

<!--
	Always rendered, so the last card to leave the hand still flies. One implicit row (no column
	count in the style): leaving cards are taken out of the flow by animate:flip, and a column
	count updated after that would briefly wrap the hand into a tall column.
-->
<ul
	class="grid auto-cols-[minmax(0,3.5rem)] grid-flow-col justify-center gap-1 pt-2"
	aria-label="Your hand"
>
	{#each items as { c, key } (key)}
		{@const isSelected = selected === c.value}
		{@const noMove = interactive && c.legalPiles.length === 0}
		{@const info = c.command ? commandInfo(c.command) : undefined}
		<li
			class="min-w-0"
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
					'relative flex aspect-[2/3] w-full flex-col items-center justify-center gap-0.5 rounded-lg border-2 bg-card text-lg font-bold text-card-foreground tabular-nums shadow-sm transition-all duration-150 select-none sm:text-xl',
					'[touch-action:manipulation] [-webkit-tap-highlight-color:transparent] disabled:cursor-default',
					info?.kind === 'lightning' && 'border-red-400 dark:border-red-500',
					info?.kind === 'infinity' && 'border-violet-400 dark:border-violet-500',
					!info && 'border-border',
					isSelected && '-translate-y-2 border-primary shadow-md ring-2 ring-primary',
					noMove && !isSelected && 'opacity-45'
				)}
			>
				{#if fresh.has(key)}
					<NewCardMark dot="corner-right" />
				{/if}
				<span class="leading-none">{c.value}</span>
				{#if c.command}
					<CommandIcon
						command={c.command}
						class={cn(
							'size-3.5 sm:size-4',
							info?.kind === 'lightning'
								? 'text-red-600 dark:text-red-400'
								: 'text-violet-600 dark:text-violet-400'
						)}
					/>
				{/if}
			</button>
		</li>
	{/each}
</ul>
{#if items.length === 0}
	<p class="py-4 text-center text-sm text-muted-foreground">No cards in hand.</p>
{/if}
