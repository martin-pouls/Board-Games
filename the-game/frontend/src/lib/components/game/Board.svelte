<script lang="ts">
	import type { Snippet } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { resolve } from '$app/paths';
	import House from '@lucide/svelte/icons/house';
	import Layers from '@lucide/svelte/icons/layers';
	import { errorText } from '#lib/errors.js';
	import { ActionError } from '#lib/game/connection.svelte.js';
	import type { GameClient } from '#lib/game/client.js';
	import { cn } from '#lib/utils.js';
	import type { HandCard, PlayerView } from '#lib/types.js';
	import PileCard from './PileCard.svelte';
	import Hand from './Hand.svelte';
	import PlayerStrip from './PlayerStrip.svelte';
	import TurnBar from './TurnBar.svelte';

	let {
		client,
		view,
		readonly = false,
		controls
	}: {
		client: GameClient;
		view: PlayerView;
		/** Read-only board (first-player choice, game over): no card selection or plays. */
		readonly?: boolean;
		/** Replaces the turn bar above the hand. */
		controls?: Snippet;
	} = $props();

	let selected = $state<number | null>(null);
	let pending = $state(false);

	const connected = $derived(client.status === 'open');
	const myTurn = $derived(
		!readonly && view.phase === 'playing' && view.turn?.activePlayerId === view.you
	);
	const interactive = $derived(myTurn && connected && !pending);
	const selectedCard = $derived(
		selected === null ? undefined : view.hand.find((c) => c.value === selected)
	);
	// Ascending piles on the top row, descending below.
	const piles = $derived(
		[...view.piles].sort((a, b) =>
			a.direction === b.direction ? a.index - b.index : a.direction === 'up' ? -1 : 1
		)
	);
	const playsByPile = $derived(
		(view.turn?.plays ?? []).reduce<Record<number, number>>((acc, p) => {
			acc[p.pile] = (acc[p.pile] ?? 0) + 1;
			return acc;
		}, {})
	);

	// Drop a selection that no longer makes sense (card played elsewhere, turn over, …).
	$effect(() => {
		if (selected !== null && (!myTurn || !selectedCard)) selected = null;
	});

	// A short buzz when it becomes your turn (phones only; ignored elsewhere).
	$effect(() => {
		const events = client.lastEvents;
		if (readonly || view.players.length < 2) return;
		if (events.some((e) => e.type === 'turnStarted' && e.playerId === view.you)) {
			try {
				navigator.vibrate?.(40);
			} catch {
				// not supported
			}
		}
	});

	function fail(e: unknown) {
		toast.error(errorText(e instanceof ActionError ? e.code : 'internal'));
	}

	function onselect(card: HandCard) {
		if (!interactive) return;
		if (selected === card.value) {
			selected = null;
			return;
		}
		if (card.legalPiles.length === 0) {
			selected = null;
			toast('No legal pile for this card', { id: 'no-legal-pile' });
			return;
		}
		selected = card.value;
	}

	async function onpile(index: number) {
		if (!interactive || !selectedCard) return;
		if (!selectedCard.legalPiles.includes(index)) {
			toast.error(errorText('illegal_placement'), { id: 'illegal' });
			return;
		}
		const value = selectedCard.value;
		selected = null;
		pending = true;
		try {
			await client.playCard(value, index);
		} catch (e) {
			fail(e);
		} finally {
			pending = false;
		}
	}

	async function endTurn() {
		if (pending) return;
		selected = null;
		pending = true;
		try {
			await client.endTurn();
		} catch (e) {
			fail(e);
		} finally {
			pending = false;
		}
	}
</script>

<div class="mx-auto flex w-full max-w-md flex-1 flex-col gap-2">
	<header class="flex items-center gap-2 text-sm">
		<a
			href={resolve('/')}
			class="-ml-1 rounded-md p-1 text-muted-foreground hover:text-foreground"
			aria-label="Home"
		>
			<House class="size-5" />
		</a>
		<span class="font-mono text-base font-semibold tracking-widest">{view.gameId}</span>
		<span
			class={cn('size-2 rounded-full', connected ? 'bg-emerald-500' : 'animate-pulse bg-amber-500')}
			title={connected ? 'Connected' : 'Reconnecting'}
		></span>
		{#if view.options.pro}
			<span class="rounded bg-muted px-1.5 py-0.5 text-xs font-semibold">PRO</span>
		{/if}
		<span
			class="ml-auto flex items-center gap-1 text-muted-foreground tabular-nums"
			title="Draw pile"
		>
			<Layers class="size-4" />
			Deck {view.drawPileCount}
		</span>
	</header>

	<PlayerStrip players={view.players} online={client.online} showCounts={view.phase !== 'lobby'} />

	<div class="grid max-h-[34rem] min-h-56 flex-1 grid-cols-2 grid-rows-2 gap-2 sm:gap-3">
		{#each piles as pile (pile.index)}
			{@const legal = selectedCard?.legalPiles.includes(pile.index) ?? false}
			<PileCard
				{pile}
				highlight={legal}
				dimmed={!!selectedCard && !legal}
				disabled={!interactive || !selectedCard}
				playsThisTurn={playsByPile[pile.index] ?? 0}
				onclick={() => onpile(pile.index)}
			/>
		{/each}
	</div>

	<div
		class="sticky bottom-0 -mx-3 mt-auto flex flex-col gap-1 bg-background/95 px-3 pt-1 pb-[max(0.25rem,env(safe-area-inset-bottom))] backdrop-blur"
	>
		{#if controls}
			{@render controls()}
		{:else}
			<TurnBar {view} isMyTurn={myTurn} disabled={!connected} {pending} onendturn={endTurn} />
		{/if}
		<Hand cards={view.hand} {selected} {interactive} {onselect} />
		<p class="h-4 text-center text-xs text-muted-foreground">
			{#if myTurn && !selectedCard && view.turn?.cardsPlayed === 0 && view.hand.length > 0}
				Tap a card, then a highlighted pile.
			{/if}
		</p>
	</div>
</div>
