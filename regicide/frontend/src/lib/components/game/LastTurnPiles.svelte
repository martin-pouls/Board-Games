<!--
	The last turn's Played and Discarded piles, so a turn can be followed after the fact (also
	once the enemy it hit is gone), each with its label below: "Yielded" for a yield, and the
	enemy it defeated as a pile of its own, "Exact kill" in red when it went onto the Tavern deck.
	Bigger piles overlap closer (both by the same amount), so the row always fits on one line.

	Card motion: your own cards fly in from your hand (same keys as there), other players' cards
	come in from their hidden hands above, and a defeated enemy flies over from the enemy panel.
	The previous turn's cards simply give way.
-->
<script lang="ts">
	import Hourglass from '@lucide/svelte/icons/hourglass';
	import { getCardMotion, type Offstage } from '@board-games/shared/cardMotion.js';
	import { handKeys } from '#lib/game/rules.js';
	import { cn } from '#lib/utils.js';
	import type { Card, LastTurn, PlayerSummary } from '#lib/types.js';
	import PlayerInitial from './PlayerInitial.svelte';
	import PlayingCard from './PlayingCard.svelte';

	let { lastTurn, players }: { lastTurn: LastTurn | null; players: PlayerSummary[] } = $props();

	const { receive } = getCardMotion();

	const yielded = $derived(!!lastTurn && lastTurn.played === null);
	/** Where cards without a counterpart in your hand come from. */
	const from = $derived<Offstage>(
		players.find((p) => p.id === lastTurn?.playerId)?.isYou ? 'bottom' : 'top'
	);

	const CARD = 48;
	const GAP = 12;
	const MAX_STEP = 26;

	let width = $state(358);
	/** How much of each card but the last in a pile shows: as much as the row's width allows. */
	const step = $derived.by(() => {
		if (!lastTurn) return MAX_STEP;
		const piles = [
			lastTurn.played,
			lastTurn.defeated ? [lastTurn.defeated] : null,
			lastTurn.discarded
		];
		const extra = piles.reduce((n, p) => n + Math.max(0, (p?.length ?? 0) - 1), 0);
		if (extra === 0) return MAX_STEP;
		const shown = piles.filter((p, i) => i !== 1 || p).length;
		// The row's padding, the player's initial, the piles' last cards and the gaps.
		const fixed = 8 + 20 + GAP + shown * CARD + (shown - 1) * GAP;
		return Math.max(10, Math.min(MAX_STEP, Math.floor((width - fixed) / extra)));
	});
</script>

{#snippet label(text: string, exact = false)}
	<span
		class={cn(
			'flex h-4 items-center text-xs whitespace-nowrap',
			exact ? 'font-bold text-health' : 'text-muted-foreground'
		)}>{text}</span
	>
{/snippet}

{#snippet stack(cards: Card[] | null | undefined, empty: 'yield' | 'none')}
	{@const keys = handKeys(cards ?? [])}
	{@const items = (cards ?? []).map((card, i) => ({ card, key: keys[i] }))}
	<!-- Always rendered, so a pile that fills up after being empty still animates. -->
	<span class="flex">
		{#each items as { card, key }, i (key)}
			<span
				class="shrink-0"
				style:margin-left={i > 0 ? `${step - CARD}px` : undefined}
				in:receive={{ key, offstage: from }}
			>
				<PlayingCard {card} size="sm" />
			</span>
		{/each}
		{#if !cards || cards.length === 0}
			<span
				class="grid aspect-[5/7] w-12 shrink-0 place-items-center rounded-md border-[1.5px] border-dashed border-border text-xs text-muted-foreground"
			>
				{#if empty === 'yield'}<Hourglass class="size-4" aria-hidden="true" />{:else}None{/if}
			</span>
		{/if}
	</span>
{/snippet}

<section
	class="flex h-22 shrink-0 items-start gap-3 px-1"
	aria-label="Last turn"
	bind:clientWidth={width}
>
	{#if lastTurn}
		<span class="mt-[18px]"><PlayerInitial id={lastTurn.playerId} {players} /></span>
	{/if}
	<div class="flex flex-col items-center gap-1" aria-label={yielded ? 'Yielded' : 'Played'}>
		{@render stack(lastTurn?.played, yielded ? 'yield' : 'none')}
		{@render label(yielded ? 'Yielded' : 'Played')}
	</div>
	{#if lastTurn?.defeated}
		{@const key = handKeys([lastTurn.defeated])[0]}
		<div
			class="flex flex-col items-center gap-1"
			aria-label={lastTurn.exact ? 'Exact kill: the enemy went onto the Tavern deck' : 'Defeated'}
			title={lastTurn.exact ? 'Exact kill: the enemy went onto the Tavern deck' : undefined}
		>
			{#key key}
				<span in:receive={{ key, offstage: 'none' }}>
					<PlayingCard card={lastTurn.defeated} size="sm" />
				</span>
			{/key}
			{@render label(lastTurn.exact ? 'Exact kill' : 'Defeated', lastTurn.exact)}
		</div>
	{/if}
	<div class="flex flex-col items-center gap-1" aria-label="Discarded">
		{@render stack(lastTurn?.discarded, 'none')}
		{@render label('Discarded')}
	</div>
</section>
