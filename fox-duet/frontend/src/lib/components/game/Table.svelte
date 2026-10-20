<script lang="ts">
	import HeaderButton, { headerRow } from '@board-games/shared/HeaderButton.svelte';
	import HeaderThemeToggle from '@board-games/shared/HeaderThemeToggle.svelte';
	import { toast } from 'svelte-sonner';
	import { resolve } from '$app/paths';
	import BookOpen from '@lucide/svelte/icons/book-open';
	import Flag from '@lucide/svelte/icons/flag';
	import House from '@lucide/svelte/icons/house';
	import type { Snippet } from 'svelte';
	import { setCardMotion } from '@board-games/shared/cardMotion.js';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { errorText } from '#lib/errors.js';
	import { ActionError } from '#lib/game/connection.svelte.js';
	import type { GameClient } from '#lib/game/client.js';
	import { trackBoardMarks } from '#lib/game/boardMarks.svelte.js';
	import { sameCard, seats } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import type { Card, PlayerView } from '#lib/types.js';
	import ForestBoard from './ForestBoard.svelte';
	import Hand from './Hand.svelte';
	import RulesHelp from './RulesHelp.svelte';
	import TrickRow from './TrickRow.svelte';
	import TurnBar from './TurnBar.svelte';

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
		/** Replaces the turn bar above the hand. */
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
	const { me } = $derived(seats(view));
	const step = $derived(pending?.step);
	const dealing = $derived(
		view.trick?.number === 1 && !view.trick.lead && !view.lastGift && view.hand.length === 11
	);
	// What changed on the board without a move (round-end gems, forest tokens), until the next lead.
	const marks = trackBoardMarks(
		() => view.board,
		() => !!view.trick?.lead,
		() => view.gameId
	);

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

<div class="mx-auto flex w-full max-w-md flex-1 flex-col gap-1.5 sm:max-w-xl">
	<header class={cn(headerRow, 'text-sm whitespace-nowrap')}>
		<HeaderButton icon={House} label="Home" href={resolve('/')} />
		<span class="font-mono text-base font-semibold tracking-widest">{view.gameId}</span>
		<span
			class={cn('size-2 shrink-0 rounded-full', connected ? 'bg-online' : 'animate-pulse bg-fox')}
			title={connected ? 'Connected' : 'Reconnecting'}
		></span>
		{#if view.round > 0}
			<span class="ml-1.5 text-muted-foreground">
				Round <b class="font-extrabold text-foreground">{view.round}</b> of 3
			</span>
		{/if}
		<HeaderButton
			class="ml-auto"
			icon={BookOpen}
			label="Rules"
			onclick={() => (rulesOpen = true)}
		/>
		{#if playing && !readonly}
			<HeaderButton
				icon={Flag}
				label="Give up"
				disabled={!connected}
				onclick={() => (confirmGiveUp = true)}
			/>
		{/if}
		<HeaderThemeToggle />
	</header>

	{#if view.board}
		<ForestBoard
			board={view.board}
			myEnd={me.end}
			players={view.players}
			online={client.online}
			marks={marks.current}
			options={mine && step === 'move' ? (pending?.moveOptions ?? []) : []}
			{highlight}
			onchoose={(delta) => run(() => client.chooseMove(delta))}
			onhover={(d) => (highlight = d)}
		/>
	{/if}

	<div class="flex-1"></div>
	<TrickRow {view} />
	<div class="flex-1"></div>

	<div class="flex flex-col gap-1.5 pb-[env(safe-area-inset-bottom)]">
		{#if controls}
			{@render controls()}
		{:else if playing && !readonly}
			<TurnBar {client} {view} {selected} {busy} {connected} {run} bind:highlight />
		{/if}
		<Hand
			cards={view.hand}
			deal={`${view.gameId}:${view.round}`}
			{selected}
			marked={pending?.yourGift ?? null}
			{enabled}
			{dealing}
			onpick={pick}
		/>
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
