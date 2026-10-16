<script lang="ts">
	import HeaderButton, { headerRow } from '@board-games/shared/HeaderButton.svelte';
	import HeaderThemeToggle from '@board-games/shared/HeaderThemeToggle.svelte';
	import { setCardMotion } from '@board-games/shared/cardMotion.js';
	import { toast } from 'svelte-sonner';
	import { resolve } from '$app/paths';
	import BookOpen from '@lucide/svelte/icons/book-open';
	import Flag from '@lucide/svelte/icons/flag';
	import House from '@lucide/svelte/icons/house';
	import type { Snippet } from 'svelte';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { errorText } from '#lib/errors.js';
	import { ActionError } from '#lib/game/connection.svelte.js';
	import type { GameClient } from '#lib/game/client.js';
	import { loadHandOrder, saveHandOrder } from '#lib/game/handOrder.js';
	import { checkPlay, previewPlay, sortHand, type HandOrder } from '#lib/game/rules.js';
	import { pickCards, selectionKey, toggleIndex } from '#lib/game/selection.js';
	import { cn } from '#lib/utils.js';
	import type { PlayerView } from '#lib/types.js';
	import ActionPanel from './ActionPanel.svelte';
	import EnemyCard from './EnemyCard.svelte';
	import Hand from './Hand.svelte';
	import HandOrderMenu from './HandOrderMenu.svelte';
	import LastTurnPiles from './LastTurnPiles.svelte';
	import PlayerStrip from './PlayerStrip.svelte';
	import RulesHelp from './RulesHelp.svelte';
	import TableCounters from './TableCounters.svelte';

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
		/** Replaces the action panel above the hand. */
		controls?: Snippet;
	} = $props();

	// Cards fly between the hand, the enemy panel and the last turn's piles.
	setCardMotion();

	let selected = $state<number[]>([]);
	let busy = $state(false);
	let confirmGiveUp = $state(false);
	let rulesOpen = $state(false);
	let handOrder = $state<HandOrder>(loadHandOrder());
	let headerIcons = $state<HTMLElement>();

	const connected = $derived(client.status === 'open');
	const playing = $derived(view.phase === 'playing');
	const step = $derived(view.turn?.step ?? '');
	const myTurn = $derived(!readonly && playing && view.turn?.activePlayerId === view.you);
	const canSelect = $derived(connected && myTurn && (step === 'play' || step === 'discard'));
	// The hand in the player's order: selection indices refer to this array, not to view.hand.
	const hand = $derived(sortHand(view.hand, handOrder));
	const selectedCards = $derived(pickCards(hand, selected));
	const preview = $derived(
		myTurn && step === 'play' && selected.length > 0 ? previewPlay(view, selectedCards) : null
	);

	// Drop the selection whenever the situation it belongs to changes (new hand, step, turn).
	const key = $derived(selectionKey(view, canSelect));
	$effect(() => {
		void key;
		selected = [];
	});

	// Close the give-up dialog once the game is over.
	$effect(() => {
		if (!playing) confirmGiveUp = false;
	});

	function hint(i: number): 'dim' | undefined {
		if (step !== 'play' || selected.length === 0) return undefined;
		return checkPlay([...selectedCards, hand[i]]).ok ? undefined : 'dim';
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

	function setHandOrder(order: HandOrder) {
		if (order === handOrder) return;
		handOrder = order;
		saveHandOrder(order);
		selected = [];
	}

	async function giveUp() {
		await run(() => client.giveUp());
		confirmGiveUp = false;
	}
</script>

<div class="mx-auto flex w-full max-w-md flex-1 flex-col gap-2 sm:max-w-xl">
	<header class="flex flex-col gap-1">
		<div bind:this={headerIcons} class={cn(headerRow, 'text-sm')}>
			<HeaderButton icon={House} label="Home" href={resolve('/')} />
			<span class="font-mono text-base font-semibold tracking-widest">{view.gameId}</span>
			<span
				class={cn(
					'size-2 rounded-full',
					connected ? 'bg-emerald-500' : 'animate-pulse bg-amber-500'
				)}
				title={connected ? 'Connected' : 'Reconnecting'}
			></span>
			<HandOrderMenu
				class="ml-auto"
				order={handOrder}
				anchor={headerIcons}
				onchange={setHandOrder}
			/>
			<HeaderButton icon={BookOpen} label="Rules" onclick={() => (rulesOpen = true)} />
			{#if playing && !readonly}
				<HeaderButton
					icon={Flag}
					label="Give up"
					disabled={!connected}
					onclick={() => (confirmGiveUp = true)}
				/>
			{/if}
			<HeaderThemeToggle />
		</div>
		<div class="flex flex-wrap items-center gap-x-3 gap-y-1">
			<TableCounters {view} />
			<PlayerStrip class="ml-auto" players={view.players} online={client.online} />
		</div>
	</header>

	{#if view.enemy}
		<EnemyCard enemy={view.enemy} players={view.players} {preview} />
	{:else}
		<p
			class="rounded-xl border bg-card/70 p-4 text-center font-display text-lg text-primary"
			role="status"
		>
			The castle has fallen.
		</p>
	{/if}

	<LastTurnPiles lastTurn={view.lastTurn} players={view.players} />

	<div
		class="sticky bottom-0 -mx-3 mt-auto flex flex-col gap-1 px-3 pt-1 pb-[max(0.5rem,env(safe-area-inset-bottom))] before:absolute before:inset-0 before:-z-10 before:backdrop-blur"
	>
		{#if controls}
			{@render controls()}
		{:else if playing && !readonly}
			<ActionPanel {client} {view} selected={selectedCards} {busy} {connected} {run} />
		{/if}
		<Hand
			cards={hand}
			deal={view.gameId}
			{selected}
			interactive={canSelect && !busy}
			ontoggle={(i) => (selected = toggleIndex(selected, i))}
			{hint}
		/>
	</div>
</div>

<RulesHelp bind:open={rulesOpen} />

<AlertDialog.Root bind:open={confirmGiveUp}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Give up for everyone?</AlertDialog.Title>
			<AlertDialog.Description>
				The game ends as a defeat for all players. You can start a new one from the results screen.
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
