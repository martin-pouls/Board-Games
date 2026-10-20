<script lang="ts">
	import HeaderButton, { headerRow } from '@board-games/shared/HeaderButton.svelte';
	import HeaderThemeToggle from '@board-games/shared/HeaderThemeToggle.svelte';
	import { setCardMotion } from '@board-games/shared/cardMotion.js';
	import type { Snippet } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { resolve } from '$app/paths';
	import House from '@lucide/svelte/icons/house';
	import Flag from '@lucide/svelte/icons/flag';
	import BookOpen from '@lucide/svelte/icons/book-open';
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
	import NoCommunicationFrame from './NoCommunicationFrame.svelte';
	import RulesHelp from './RulesHelp.svelte';

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

	// One card motion for the board: cards fly between the hand and the piles.
	setCardMotion();

	let selected = $state<number | null>(null);
	let pending = $state(false);
	let confirmGiveUp = $state(false);
	let rulesOpen = $state(false);

	const connected = $derived(client.status === 'open');
	const extreme = $derived(view.options.variant === 'extreme');
	// Turn markers (bound pile, skull to cover) only matter while the game runs.
	const playing = $derived(view.phase === 'playing');
	const myTurn = $derived(
		!readonly && view.phase === 'playing' && view.turn?.activePlayerId === view.you
	);
	const interactive = $derived(myTurn && connected && !pending);
	// Anyone may give up while the game runs, whoever's turn it is.
	const running = $derived(view.phase === 'playing' || view.phase === 'choosingFirstPlayer');
	const stuck = $derived(myTurn && !!view.turn?.stuck);
	const selectedCard = $derived(
		selected === null ? undefined : view.hand.find((c) => c.value === selected)
	);
	// Ascending piles on the top row, descending below.
	const piles = $derived(
		[...view.piles].sort((a, b) =>
			a.direction === b.direction ? a.index - b.index : a.direction === 'up' ? -1 : 1
		)
	);
	// Pile card size: one --w for all four piles, from the grid's size, which stays the same all
	// game (the hand's height depends only on the hand size). The card sits between the corner
	// pills or, when that gives a bigger card, below them.
	const GAP = 8;
	/** Corner pill: inset 6 px, 28 × 22 px; the card keeps 4 px from it and 6 px from the mat edge. */
	const PILL_SIDE = 6 + 28 + 4;
	const PILL_BOTTOM = 6 + 22 + 4;
	const PAD = 6;
	const MAX_PILE_CARD = 112;
	let gridW = $state(0);
	let gridH = $state(0);
	const pileCard = $derived.by(() => {
		const w = (gridW - GAP) / 2;
		const h = (gridH - GAP) / 2;
		const beside = Math.min(w - 2 * PILL_SIDE, (h - 2 * PAD) / 1.4);
		const below = Math.min(w - 2 * PAD, (h - PILL_BOTTOM - PAD) / 1.4);
		const size = Math.max(0, Math.floor(Math.min(MAX_PILE_CARD, Math.max(beside, below))));
		// Below the pills: pushed down only as far as needed to clear them.
		const top = size > beside ? Math.max(0, 2 * PILL_BOTTOM - (h - size * 1.4)) : 0;
		return { size, top };
	});
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

	// Close the give-up dialog once the game is over (someone else gave up, or it ended).
	$effect(() => {
		if (confirmGiveUp && !running) confirmGiveUp = false;
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
			// A command veto names its rule; otherwise the base rules forbid the pile.
			const veto = selectedCard.vetoes?.find((v) => v.pile === index);
			toast.error(errorText(veto?.code ?? 'illegal_placement'), { id: 'illegal' });
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

	/** Runs one turn action (undo, end turn, give up) with the board locked meanwhile. */
	async function turnAction(action: () => Promise<void>) {
		if (pending) return;
		selected = null;
		pending = true;
		try {
			await action();
		} catch (e) {
			fail(e);
		} finally {
			pending = false;
		}
	}

	const undo = () => turnAction(() => client.undoPlay());
	const endTurn = () => turnAction(() => client.endTurn());

	async function giveUp() {
		await turnAction(() => client.giveUp());
		confirmGiveUp = false;
	}
</script>

<div class="mx-auto flex w-full max-w-md flex-1 flex-col gap-2">
	<header class={cn(headerRow, 'gap-1 text-sm sm:gap-1.5')}>
		<HeaderButton icon={House} label="Home" href={resolve('/')} />
		<span class="font-mono text-base font-semibold tracking-widest">{view.gameId}</span>
		<span
			class={cn('size-2 rounded-full', connected ? 'bg-go' : 'animate-pulse bg-extreme')}
			title={connected ? 'Connected' : 'Reconnecting'}
		></span>
		{#if view.options.pro}
			<span class="rounded bg-muted px-1.5 py-0.5 text-xs font-semibold">PRO</span>
		{/if}
		{#if extreme}
			<span
				class="ml-1 -rotate-3 bg-extreme px-1.5 py-0.5 text-xs font-extrabold tracking-wide text-extreme-ink"
				title="Extreme"
			>
				<!-- Abbreviated on phones, where the header row has room for nothing more. -->
				<span class="sm:hidden" aria-hidden="true">EXT</span><span class="max-sm:sr-only"
					>EXTREME</span
				>
			</span>
		{/if}
		<HeaderButton
			class="ml-auto"
			icon={BookOpen}
			label="Rules"
			onclick={() => (rulesOpen = true)}
		/>
		{#if running}
			<HeaderButton
				icon={Flag}
				label="Give up"
				disabled={!connected}
				onclick={() => (confirmGiveUp = true)}
			/>
		{/if}
		<HeaderThemeToggle />
	</header>

	{#if view.communicationRestricted}
		<NoCommunicationFrame />
	{/if}

	<PlayerStrip players={view.players} online={client.online} showCounts={view.phase !== 'lobby'} />

	<div
		bind:clientWidth={gridW}
		bind:clientHeight={gridH}
		class="grid max-h-[34rem] min-h-56 flex-1 grid-cols-2 grid-rows-2 gap-2"
		style:--w="{pileCard.size}px"
		style:--mt="{pileCard.top}px"
	>
		{#each piles as pile (pile.index)}
			{@const legal = selectedCard?.legalPiles.includes(pile.index) ?? false}
			{@const playable = interactive && !!selectedCard}
			<PileCard
				{pile}
				highlight={legal}
				dimmed={!!selectedCard && !legal}
				disabled={!playable}
				playsThisTurn={playsByPile[pile.index] ?? 0}
				bound={playing && view.turn?.constraints.boundPile === pile.index}
				skull={playing && (view.turn?.constraints.skullPiles.includes(pile.index) ?? false)}
				onclick={() => onpile(pile.index)}
			/>
		{/each}
	</div>

	<!-- Turn bar, buttons and hand: fixed heights all game, so the piles never resize. -->
	<div
		class="sticky bottom-0 -mx-3 mt-auto flex flex-col gap-2 px-3 pt-1 pb-[max(0.25rem,env(safe-area-inset-bottom))] before:absolute before:inset-0 before:-z-10 before:backdrop-blur"
	>
		{#if controls}
			{@render controls()}
		{:else}
			<TurnBar
				{view}
				isMyTurn={myTurn}
				disabled={!connected}
				{pending}
				onundo={undo}
				onendturn={endTurn}
				ongiveup={() => (confirmGiveUp = true)}
			/>
		{/if}
		<Hand
			cards={view.hand}
			handSize={view.rules.handSize}
			deal={view.gameId}
			{selected}
			{interactive}
			{onselect}
		/>
	</div>
</div>

<RulesHelp bind:open={rulesOpen} {extreme} />

<AlertDialog.Root bind:open={confirmGiveUp}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Give up and lose the game?</AlertDialog.Title>
			<AlertDialog.Description>
				{#if stuck}
					No card fits and your turn can't end. You can still undo your cards from this turn and try
					another way.
				{:else}
					The game ends as a defeat for everyone and the cards left are counted. You can start a new
					game from the results screen.
				{/if}
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={pending}
				>{stuck ? 'Keep trying' : 'Keep playing'}</AlertDialog.Cancel
			>
			<AlertDialog.Action
				class="bg-destructive text-white hover:bg-destructive/90"
				disabled={pending || !connected}
				onclick={giveUp}
			>
				Give up
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
