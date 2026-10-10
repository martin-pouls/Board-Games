<script lang="ts">
	import { untrack, type Snippet } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { resolve } from '$app/paths';
	import House from '@lucide/svelte/icons/house';
	import Layers from '@lucide/svelte/icons/layers';
	import Flag from '@lucide/svelte/icons/flag';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { errorText } from '#lib/errors.js';
	import { ActionError } from '#lib/game/connection.svelte.js';
	import type { GameClient } from '#lib/game/client.js';
	import { cn } from '#lib/utils.js';
	import type { HandCard, PlayerView } from '#lib/types.js';
	import PileCard from './PileCard.svelte';
	import Hand from './Hand.svelte';
	import PlayerStrip from './PlayerStrip.svelte';
	import TurnBar from './TurnBar.svelte';
	import ActiveCommands from './ActiveCommands.svelte';
	import NoCommunicationBanner from './NoCommunicationBanner.svelte';

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
	let confirmEnd = $state(false);
	let ending = $state(false);

	const connected = $derived(client.status === 'open');
	const extreme = $derived(view.options.variant === 'extreme');
	// Turn markers (bound pile, skull to cover) only matter while the game runs.
	const playing = $derived(view.phase === 'playing');
	const myTurn = $derived(
		!readonly && view.phase === 'playing' && view.turn?.activePlayerId === view.you
	);
	const interactive = $derived(myTurn && connected && !pending);
	const canEndGame = $derived(
		view.hostId === view.you && (view.phase === 'playing' || view.phase === 'choosingFirstPlayer')
	);
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

	// Extreme: toasts for command cards that change the flow of play.
	$effect(() => {
		const events = client.lastEvents;
		if (readonly) return;
		untrack(() => {
			for (const e of events) {
				if (e.type !== 'cardPlayed') continue;
				if (e.card.command === 'stop') {
					const name = view.players.find((p) => p.id === e.playerId)?.name ?? 'Someone';
					toast(e.playerId === view.you ? 'STOP: your turn is over.' : `${name} played STOP.`, {
						id: 'stop'
					});
				} else if (e.card.command === 'noCommunication') {
					toast('No communication until the card is covered.', { id: 'no-comm' });
				}
			}
		});
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
			toast(card.blocked ? errorText(card.blocked) : 'No legal pile for this card', {
				id: 'no-legal-pile'
			});
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

	async function endGame() {
		if (ending) return;
		ending = true;
		try {
			await client.endGame();
			confirmEnd = false;
		} catch (e) {
			fail(e);
		} finally {
			ending = false;
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
		{#if extreme}
			<span class="rounded bg-violet-600 px-1.5 py-0.5 text-xs font-semibold text-white">
				EXTREME
			</span>
		{/if}
		<span
			class="ml-auto flex items-center gap-1 text-muted-foreground tabular-nums"
			title="Draw pile"
		>
			<Layers class="size-4" />
			<span class={cn(extreme && 'max-[24rem]:sr-only')}>Deck</span>
			{view.drawPileCount}
		</span>
		{#if canEndGame}
			<button
				type="button"
				class="-mr-1 flex items-center gap-1 rounded-md p-1 text-muted-foreground hover:text-destructive disabled:opacity-50"
				disabled={!connected}
				title="End game"
				onclick={() => (confirmEnd = true)}
			>
				<Flag class="size-4" />
				<!-- The EXTREME badge leaves no room for the label on phones. -->
				<span class={cn(extreme && 'max-[24rem]:sr-only')}>End game</span>
			</button>
		{/if}
	</header>

	{#if view.communicationRestricted}
		<NoCommunicationBanner />
	{/if}

	<PlayerStrip players={view.players} online={client.online} showCounts={view.phase !== 'lobby'} />

	<ActiveCommands commands={view.activeCommands} piles={view.piles} />

	<div class="grid max-h-[34rem] min-h-56 flex-1 grid-cols-2 grid-rows-2 gap-2 sm:gap-3">
		{#each piles as pile (pile.index)}
			{@const legal = selectedCard?.legalPiles.includes(pile.index) ?? false}
			<PileCard
				{pile}
				highlight={legal}
				dimmed={!!selectedCard && !legal}
				disabled={!interactive || !selectedCard}
				playsThisTurn={playsByPile[pile.index] ?? 0}
				bound={playing && view.turn?.constraints.boundPile === pile.index}
				skull={playing && (view.turn?.constraints.skullPiles.includes(pile.index) ?? false)}
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

<AlertDialog.Root bind:open={confirmEnd}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>End the game for everyone?</AlertDialog.Title>
			<AlertDialog.Description>
				The round stops now and the cards left are counted. You can start a new game from the
				results screen.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={ending}>Keep playing</AlertDialog.Cancel>
			<AlertDialog.Action
				class="bg-destructive text-white hover:bg-destructive/90"
				disabled={ending || !connected}
				onclick={endGame}
			>
				End game
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
