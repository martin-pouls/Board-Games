<script lang="ts">
	import { toast } from 'svelte-sonner';
	import { resolve } from '$app/paths';
	import BookOpen from '@lucide/svelte/icons/book-open';
	import Flag from '@lucide/svelte/icons/flag';
	import Gem from '@lucide/svelte/icons/gem';
	import House from '@lucide/svelte/icons/house';
	import Trees from '@lucide/svelte/icons/trees';
	import type { Snippet } from 'svelte';
	import { setCardMotion } from '@board-games/shared/cardMotion.js';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { errorText } from '#lib/errors.js';
	import { ActionError } from '#lib/game/connection.svelte.js';
	import type { GameClient } from '#lib/game/client.js';
	import { cardLabel, describeEvents, plural, sameCard, seats } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import type { Card, PlayerView } from '#lib/types.js';
	import DecisionPanel from './DecisionPanel.svelte';
	import ForestBoard from './ForestBoard.svelte';
	import Hand from './Hand.svelte';
	import RulesHelp from './RulesHelp.svelte';
	import TrickRow from './TrickRow.svelte';

	let {
		client,
		view,
		readonly = false,
		controls
	}: {
		client: GameClient;
		view: PlayerView;
		/** Read-only table (game over): no card selection or actions. */
		readonly?: boolean;
		/** Replaces the decision panel above the hand. */
		controls?: Snippet;
	} = $props();

	setCardMotion();

	let selected = $state<Card | null>(null);
	let highlight = $state<number | null>(null);
	let busy = $state(false);
	let confirmGiveUp = $state(false);
	let rulesOpen = $state(false);

	const connected = $derived(client.status === 'open');
	const playing = $derived(view.phase === 'playing');
	const pending = $derived(view.pending);
	const mine = $derived(!readonly && playing && !!pending?.playerIds.includes(view.you));
	const { me, mate } = $derived(seats(view));
	const step = $derived(pending?.step);
	const dealing = $derived(
		view.trick?.number === 1 && !view.trick.lead && !view.lastGift && view.hand.length === 11
	);
	const chronicle = $derived(describeEvents(client.lastEvents, view));

	// Drop the selection whenever the decision it belongs to changes.
	const selectionKey = $derived(
		`${view.round}|${view.trick?.number}|${step}|${mine}|${view.hand.length}`
	);
	$effect(() => {
		void selectionKey;
		selected = null;
		highlight = null;
	});

	// Close the give-up dialog once the game is over.
	$effect(() => {
		if (!playing) confirmGiveUp = false;
	});

	function enabled(c: Card): boolean {
		if (!mine || busy || !connected) return false;
		switch (step) {
			case 'lead':
			case 'follow':
				return view.playable.some((p) => sameCard(p, c));
			case 'foxesSwap':
				return true;
			case 'gift':
				return !pending?.yourGift;
			default:
				return false;
		}
	}

	function pick(c: Card) {
		// Tapping the selected card again plays it.
		if (sameCard(c, selected) && (step === 'lead' || step === 'follow')) {
			void run(() => client.playCard(c));
			return;
		}
		selected = sameCard(c, selected) ? null : c;
	}

	async function run(action: () => Promise<void>) {
		if (busy) return;
		busy = true;
		try {
			await action();
		} catch (e) {
			toast.error(errorText(e instanceof ActionError ? e.code : 'internal'));
		} finally {
			busy = false;
		}
	}

	async function giveUp() {
		await run(() => client.giveUp());
		confirmGiveUp = false;
	}
</script>

<div class="mx-auto flex w-full max-w-md flex-1 flex-col gap-2 sm:max-w-xl">
	<header class="flex flex-col gap-1">
		<div class="flex items-center gap-2 text-sm">
			<a
				href={resolve('/')}
				class="-ml-1 rounded-md p-1 text-muted-foreground hover:text-foreground"
				aria-label="Home"
			>
				<House class="size-5" />
			</a>
			<span class="font-mono text-base font-semibold tracking-widest">{view.gameId}</span>
			<span
				class={cn(
					'size-2 rounded-full',
					connected ? 'bg-emerald-500' : 'animate-pulse bg-amber-500'
				)}
				title={connected ? 'Connected' : 'Reconnecting'}
			></span>
			<span class="text-muted-foreground">Level {view.level}</span>
			<button
				type="button"
				class="ml-auto rounded-md p-1 text-muted-foreground hover:text-foreground"
				aria-label="Rules"
				title="Rules"
				onclick={() => (rulesOpen = true)}
			>
				<BookOpen class="size-5" />
			</button>
			{#if playing && !readonly}
				<button
					type="button"
					class="-mr-1 flex items-center gap-1 rounded-md p-1 text-muted-foreground hover:text-destructive disabled:opacity-50"
					disabled={!connected}
					title="Give up"
					onclick={() => (confirmGiveUp = true)}
				>
					<Flag class="size-4" />
					<span class="max-[24rem]:sr-only">Give up</span>
				</button>
			{/if}
		</div>
		<div class="flex flex-wrap items-center gap-x-3 gap-y-0.5 text-sm">
			<span>Round <b>{view.round}</b>/3</span>
			{#if view.trick?.number}<span>Trick <b>{view.trick.number}</b>/11</span>{/if}
			{#if view.board}
				<span class="flex items-center gap-1" title="Gems left in the forest">
					<Gem class="size-4 text-gem" /><b>{view.board.gemsLeft}</b>
				</span>
				<span class="flex items-center gap-1" title="Forest tokens left">
					<Trees class="size-4 text-moss" /><b>{view.board.forestTokens}</b>/{view.board
						.forestTokensTotal}
				</span>
			{/if}
			{#if mate}
				<span class="ml-auto flex items-center gap-1.5 text-muted-foreground">
					<span
						class={cn(
							'size-2 rounded-full',
							client.online.includes(mate.id) ? 'bg-emerald-500' : 'bg-muted-foreground/40'
						)}
					></span>
					{mate.name} · {plural(mate.handCount, 'card')}{mate.isDealer ? ' · dealer' : ''}
				</span>
			{/if}
		</div>
	</header>

	{#if view.board}
		<ForestBoard
			board={view.board}
			myEnd={me.end}
			players={view.players}
			options={mine && step === 'move' ? (pending?.moveOptions ?? []) : []}
			{highlight}
			onchoose={(delta) => run(() => client.chooseMove(delta))}
			onhover={(d) => (highlight = d)}
		/>
	{/if}

	<TrickRow {view} />

	<p class="min-h-5 text-center text-sm text-muted-foreground" aria-live="polite">{chronicle}</p>

	<div
		class="sticky bottom-0 -mx-3 mt-auto flex flex-col gap-1 bg-background/95 px-3 pt-1 pb-[max(0.5rem,env(safe-area-inset-bottom))] backdrop-blur"
	>
		{#if controls}
			{@render controls()}
		{:else if playing && !readonly}
			<DecisionPanel {client} {view} {selected} {busy} {connected} {run} bind:highlight />
		{/if}
		{#if view.lastGift && !view.trick?.follow && step !== 'gift'}
			<p class="text-center text-xs text-muted-foreground">
				Gift: you gave {cardLabel(view.lastGift.gave)} and received {cardLabel(
					view.lastGift.received
				)}.
			</p>
		{/if}
		<Hand
			cards={view.hand}
			{selected}
			marked={pending?.yourGift ?? null}
			{enabled}
			{dealing}
			onpick={pick}
		/>
		<p class="text-center text-xs text-muted-foreground">
			You{me.isDealer ? ' · dealer' : ''}{me.isHost ? ' · host' : ''}
		</p>
	</div>
</div>

<RulesHelp bind:open={rulesOpen} />

<AlertDialog.Root bind:open={confirmGiveUp}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Give up for both of you?</AlertDialog.Title>
			<AlertDialog.Description>
				The game ends as a defeat. You can start a new one from the results screen.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={busy}>Keep playing</AlertDialog.Cancel>
			<AlertDialog.Action
				class="bg-destructive text-white hover:bg-destructive/90"
				disabled={busy || !connected}
				onclick={giveUp}
			>
				Give up
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
