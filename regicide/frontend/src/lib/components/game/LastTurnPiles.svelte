<!--
	The last turn's Played and Discarded piles, so a turn can be followed after the fact (also
	once the enemy it hit is gone). Icons instead of labels: swords for the play (an hourglass in
	the slot for a yield, a skull before the enemy it defeated, with "Exact" like the play preview
	when it went onto the Tavern deck), layers for the discard, like the discard pile counter.

	Card motion: your own cards fly in from your hand (same keys as there), other players' cards
	come in from their hidden hands above, and a defeated enemy flies over from the enemy panel.
	The previous turn's cards simply give way.
-->
<script lang="ts">
	import Hourglass from '@lucide/svelte/icons/hourglass';
	import Layers from '@lucide/svelte/icons/layers';
	import Skull from '@lucide/svelte/icons/skull';
	import Swords from '@lucide/svelte/icons/swords';
	import { getCardMotion, type Offstage } from '@board-games/shared/cardMotion.js';
	import { handKeys } from '#lib/game/rules.js';
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
</script>

{#snippet stack(cards: Card[] | null | undefined, empty: 'yield' | 'none')}
	{@const keys = handKeys(cards ?? [])}
	{@const items = (cards ?? []).map((card, i) => ({ card, key: keys[i] }))}
	<!-- Always rendered, so a pile that fills up after being empty still animates. -->
	<span class="flex -space-x-6">
		{#each items as { card, key } (key)}
			<span in:receive={{ key, offstage: from }}>
				<PlayingCard {card} size="sm" />
			</span>
		{/each}
		{#if !cards || cards.length === 0}
			<span
				class="flex aspect-[5/7] w-11 shrink-0 items-center justify-center rounded-lg border border-dashed border-muted-foreground/40 text-muted-foreground"
			>
				{#if empty === 'yield'}<Hourglass class="size-4" aria-label="Yielded" />{/if}
			</span>
		{/if}
	</span>
{/snippet}

<section class="flex flex-wrap items-center gap-x-4 gap-y-2 px-1" aria-label="Last turn">
	{#if lastTurn}
		<PlayerInitial id={lastTurn.playerId} {players} />
	{/if}
	<div class="flex flex-auto items-center gap-1.5" aria-label="Played">
		<span title="Played">
			<Swords class="size-4 text-muted-foreground" aria-hidden="true" />
		</span>
		{@render stack(lastTurn?.played, yielded ? 'yield' : 'none')}
		{#if lastTurn?.defeated}
			{@const key = handKeys([lastTurn.defeated])[0]}
			{#if lastTurn.exact}
				<span
					class="ml-1 flex items-center gap-1 rounded-full border border-health/60 px-2 py-0.5 text-xs whitespace-nowrap"
					title="Exact: the enemy went onto the Tavern deck"
				>
					<Skull class="size-3.5 text-health" aria-hidden="true" />
					Exact
				</span>
			{:else}
				<span title="Defeated"><Skull class="ml-1 size-4 text-health" aria-label="Defeated" /></span
				>
			{/if}
			{#key key}
				<span in:receive={{ key, offstage: 'none' }}>
					<PlayingCard card={lastTurn.defeated} size="sm" />
				</span>
			{/key}
		{/if}
	</div>
	<div class="flex flex-auto items-center gap-1.5" aria-label="Discarded">
		<span title="Discarded">
			<Layers class="size-4 text-muted-foreground" aria-hidden="true" />
		</span>
		{@render stack(lastTurn?.discarded, 'none')}
	</div>
</section>
