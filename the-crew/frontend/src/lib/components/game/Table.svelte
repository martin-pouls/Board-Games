<script lang="ts">
	import { toast } from 'svelte-sonner';
	import { resolve } from '$app/paths';
	import BookOpen from '@lucide/svelte/icons/book-open';
	import Crown from '@lucide/svelte/icons/crown';
	import Flag from '@lucide/svelte/icons/flag';
	import House from '@lucide/svelte/icons/house';
	import MessageSquare from '@lucide/svelte/icons/message-square';
	import MicOff from '@lucide/svelte/icons/mic-off';
	import Trophy from '@lucide/svelte/icons/trophy';
	import { setCardMotion } from '@board-games/shared/cardMotion.js';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import { errorText } from '#lib/errors.js';
	import { ActionError } from '#lib/game/connection.svelte.js';
	import type { GameClient } from '#lib/game/client.js';
	import { cardLabel, sameCard } from '#lib/game/cards.js';
	import { describeEvents } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import {
		OWNER_UNASSIGNED,
		type Attempt,
		type Card,
		type PlayerView,
		type Task
	} from '#lib/types.js';
	import CommBadge from './CommBadge.svelte';
	import CommunicateSheet from './CommunicateSheet.svelte';
	import CrewStrip from './CrewStrip.svelte';
	import Debrief from './Debrief.svelte';
	import DecisionPanel from './DecisionPanel.svelte';
	import Hand from './Hand.svelte';
	import MissionPanel from './MissionPanel.svelte';
	import RulesHelp from './RulesHelp.svelte';
	import TrickArea from './TrickArea.svelte';

	let { client, view, attempt }: { client: GameClient; view: PlayerView; attempt: Attempt } =
		$props();

	setCardMotion();

	let selected = $state<Card | null>(null);
	let selectedTasks = $state<number[]>([]);
	let busy = $state(false);
	let confirmGiveUp = $state(false);
	let rulesOpen = $state(false);
	let commOpen = $state(false);

	const connected = $derived(client.status === 'open');
	const playing = $derived(view.phase === 'playing');
	const mission = $derived(view.missionDef);
	const step = $derived(attempt.step);
	const me = $derived(attempt.members[attempt.seat]);
	const mine = $derived(playing && attempt.actors.includes(attempt.seat));
	const dummy = $derived(attempt.members.find((m) => m.dummy));
	const forJarvis = $derived(mine && !!dummy && attempt.turn === dummy.seat);
	const ready = $derived(!busy && connected);
	const chronicle = $derived(describeEvents(client.lastEvents, attempt.members, attempt.seat));
	const dealing = $derived(
		attempt.trickNumber === 1 && attempt.trick.length === 0 && step !== 'play'
	);

	// Drop the selection whenever the decision it belongs to changes.
	const selectionKey = $derived(
		`${attempt.number}|${step}|${attempt.pick}|${attempt.trickNumber}|${attempt.trick.length}|${attempt.hand.length}|${mine}`
	);
	$effect(() => {
		void selectionKey;
		selected = null;
		selectedTasks = [];
	});

	// Close dialogs once the attempt is over.
	$effect(() => {
		if (!playing) {
			confirmGiveUp = false;
			commOpen = false;
		}
	});

	const inPlayable = (c: Card) => attempt.playable.some((p) => sameCard(p, c));

	function handEnabled(c: Card): boolean {
		if (!ready || !playing) return false;
		if (step === 'play') return mine && !forJarvis && inPlayable(c);
		if (step === 'pass') return !!attempt.pass && !attempt.pass.chosen && c.suit !== 'trump';
		return false;
	}
	function jarvisEnabled(c: Card): boolean {
		return ready && playing && step === 'play' && forJarvis && inPlayable(c);
	}
	function pickCard(c: Card) {
		// Tapping the selected card again plays it.
		if (step === 'play' && sameCard(c, selected)) {
			void run(() => client.playCard(c));
			return;
		}
		selected = sameCard(c, selected) ? null : c;
	}

	function taskEnabled(t: Task): boolean {
		if (!ready || !mine) return false;
		if (step === 'draft') return t.owner === OWNER_UNASSIGNED;
		if (step === 'adjustTokens') return true;
		return false;
	}
	function pickTask(t: Task) {
		if (step === 'draft') {
			void run(() => client.pickTask(t.id));
			return;
		}
		selectedTasks = selectedTasks.includes(t.id)
			? selectedTasks.filter((id) => id !== t.id)
			: [...selectedTasks, t.id].slice(-2);
	}

	/** Runs an action; false when it failed (the error is shown as a toast). */
	async function run(action: () => Promise<void>): Promise<boolean> {
		if (busy) return false;
		busy = true;
		try {
			await action();
			return true;
		} catch (e) {
			toast.error(errorText(e instanceof ActionError ? e.code : 'internal'));
			return false;
		} finally {
			busy = false;
		}
	}

	async function giveUp() {
		await run(() => client.giveUp());
		confirmGiveUp = false;
	}

	const exchangeText = $derived.by(() => {
		const ex = attempt.exchange;
		if (!ex || !ex.gave || !ex.got || attempt.trickNumber > (ex.kind === 'draw' ? 2 : 1)) return '';
		return ex.kind === 'pass'
			? `Distress pass: you gave ${cardLabel(ex.gave)} and received ${cardLabel(ex.got)}.`
			: `Draw: ${cardLabel(ex.gave)} was taken from you; you drew ${cardLabel(ex.got)}.`;
	});
	const canCommunicate = $derived(attempt.actions.communicate.length > 0 && ready);
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
				class={cn('size-2 rounded-full', connected ? 'bg-success' : 'animate-pulse bg-amber-500')}
				title={connected ? 'Connected' : 'Reconnecting'}
			></span>
			<button
				type="button"
				class="ml-auto rounded-md p-1 text-muted-foreground hover:text-foreground"
				aria-label="Rules"
				title="Rules"
				onclick={() => (rulesOpen = true)}
			>
				<BookOpen class="size-5" />
			</button>
			{#if playing}
				<button
					type="button"
					class="-mr-1 flex items-center gap-1 rounded-md p-1 text-muted-foreground hover:text-destructive disabled:opacity-50"
					disabled={!connected}
					title="Give up this attempt"
					onclick={() => (confirmGiveUp = true)}
				>
					<Flag class="size-4" />
					<span class="max-[24rem]:sr-only">Give up</span>
				</button>
			{/if}
		</div>
		<div class="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm">
			<span class="font-display text-base font-bold">Mission {view.mission}</span>
			<span class="text-muted-foreground">Attempt {attempt.number}</span>
			{#if mission?.deadZone}
				<span
					class="rounded-full bg-muted px-2 py-0.5 text-xs"
					title="Communicate without position"
				>
					dead zone
				</span>
			{/if}
			{#if mission && mission.fromTrick > 1}
				<span class="rounded-full bg-muted px-2 py-0.5 text-xs">
					comms from trick {mission.fromTrick}
				</span>
			{/if}
			{#if attempt.distress}
				<span class="rounded-full bg-muted px-2 py-0.5 text-xs">distress signal used</span>
			{/if}
		</div>
	</header>

	<CrewStrip {attempt} online={client.online} {jarvisEnabled} {selected} onpick={pickCard} />

	<MissionPanel {attempt} canPick={taskEnabled} selected={selectedTasks} onpick={pickTask} />

	{#if step === 'play' || step === ''}
		<TrickArea {attempt} />
		<p class="min-h-5 text-center text-sm text-muted-foreground" aria-live="polite">{chronicle}</p>
	{/if}

	<div
		class="sticky bottom-0 -mx-3 mt-auto flex flex-col gap-1.5 bg-background/95 px-3 pt-1 pb-[max(0.5rem,env(safe-area-inset-bottom))] backdrop-blur"
	>
		{#if playing}
			<DecisionPanel
				{client}
				{attempt}
				{mission}
				selectedCard={selected}
				{selectedTasks}
				{busy}
				{connected}
				{run}
			/>
		{:else}
			<Debrief {client} {view} {attempt} {busy} {connected} {run} />
		{/if}

		{#if exchangeText}
			<p class="text-center text-xs text-muted-foreground">{exchangeText}</p>
		{/if}

		{#if playing}
			<Hand
				cards={attempt.hand}
				{selected}
				marked={attempt.pass?.chosen ?? me?.communication?.card ?? null}
				enabled={handEnabled}
				{dealing}
				onpick={pickCard}
			/>
			<div class="flex items-center gap-2 text-xs text-muted-foreground">
				<span class="flex items-center gap-1">
					You
					{#if attempt.seat === attempt.commander}
						<Crown class="size-3.5 text-rocket-ink" aria-label="Commander" />
					{/if}
					{#if attempt.seat === attempt.silenced}
						<MicOff class="size-3.5 text-danger" aria-label="Silenced" />
					{/if}
				</span>
				<span class="flex items-center gap-0.5" title="Tricks won">
					<Trophy class="size-3.5" />{me?.tricksWon ?? 0}
				</span>
				{#if me?.communication}
					<span class="flex items-center gap-1">
						<CommBadge comm={me.communication} />
						{cardLabel(me.communication.card)}
					</span>
				{/if}
				{#if step === 'play' && !me?.communicated && attempt.seat !== attempt.silenced}
					<Button
						size="sm"
						variant="outline"
						class="ml-auto h-8"
						disabled={!canCommunicate}
						title={canCommunicate ? 'Communicate a card' : 'Not possible right now'}
						onclick={() => (commOpen = true)}
					>
						<MessageSquare />Communicate
					</Button>
				{/if}
			</div>
		{/if}
	</div>
</div>

<RulesHelp bind:open={rulesOpen} {mission} />

<CommunicateSheet
	bind:open={commOpen}
	options={attempt.actions.communicate}
	deadZone={mission?.deadZone ?? false}
	disabled={!ready}
	oncommunicate={(o) => run(() => client.communicate(o.card, o.position))}
/>

<AlertDialog.Root bind:open={confirmGiveUp}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Give up this attempt?</AlertDialog.Title>
			<AlertDialog.Description>
				It counts as a failed attempt in the logbook. The host can then start the next attempt.
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
