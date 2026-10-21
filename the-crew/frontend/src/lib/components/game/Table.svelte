<script lang="ts">
	import HeaderButton, { headerRow } from '@board-games/shared/HeaderButton.svelte';
	import HeaderThemeToggle from '@board-games/shared/HeaderThemeToggle.svelte';
	import { toast } from 'svelte-sonner';
	import { resolve } from '$app/paths';
	import BookOpen from '@lucide/svelte/icons/book-open';
	import Flag from '@lucide/svelte/icons/flag';
	import House from '@lucide/svelte/icons/house';
	import { setCardMotion } from '@board-games/shared/cardMotion.js';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { errorText } from '#lib/errors.js';
	import { ActionError } from '#lib/game/connection.svelte.js';
	import type { GameClient } from '#lib/game/client.js';
	import { cardLabel, sameCard } from '#lib/game/cards.js';
	import type { TablePreset } from '#lib/game/tablePreset.js';
	import { cn } from '#lib/utils.js';
	import {
		OWNER_UNASSIGNED,
		type Attempt,
		type Card,
		type CommOption,
		type PlayerView,
		type Task
	} from '#lib/types.js';
	import ActionRow from './ActionRow.svelte';
	import CrewStrip from './CrewStrip.svelte';
	import Debrief from './Debrief.svelte';
	import DecisionPanel from './DecisionPanel.svelte';
	import Hand from './Hand.svelte';
	import MemberRow from './MemberRow.svelte';
	import MissionPanel from './MissionPanel.svelte';
	import RulesHelp from './RulesHelp.svelte';
	import type { TaskPick } from './TaskList.svelte';
	import TaskPool from './TaskPool.svelte';
	import TrickArea from './TrickArea.svelte';

	let {
		client,
		view,
		attempt,
		preset = null
	}: {
		client: GameClient;
		view: PlayerView;
		attempt: Attempt;
		/** Dev fixtures only: the UI state to start with. */
		preset?: TablePreset | null;
	} = $props();

	setCardMotion();

	// svelte-ignore state_referenced_locally
	const start = preset;
	let selected = $state<Card | null>(start?.selected ?? null);
	let selectedTasks = $state<number[]>(start?.tasks ?? []);
	let roles = $state<(number | null)[]>([]);
	let busy = $state(false);
	let confirmGiveUp = $state(false);
	let rulesOpen = $state(false);
	/** The communicate flow: Communicate was tapped; `shown` is the card picked to show. */
	let communicating = $state(start?.communicate ?? false);
	let shown = $state<Card | null>(start?.shown ?? null);

	const connected = $derived(client.status === 'open');
	const playing = $derived(view.phase === 'playing');
	const mission = $derived(view.missionDef);
	const step = $derived(attempt.step);
	const me = $derived(attempt.members[attempt.seat]);
	const mine = $derived(playing && attempt.actors.includes(attempt.seat));
	const dummy = $derived(attempt.members.find((m) => m.dummy));
	const forJarvis = $derived(mine && !!dummy && attempt.turn === dummy.seat);
	const ready = $derived(!busy && connected);
	const dealing = $derived(
		attempt.trickNumber === 1 && attempt.trick.length === 0 && step !== 'play'
	);

	// Drop the selection whenever the decision it belongs to changes.
	const assigned = $derived(attempt.tasks.filter((t) => t.owner !== OWNER_UNASSIGNED).length);
	const selectionKey = $derived(
		`${attempt.number}|${step}|${attempt.pick}|${attempt.trickNumber}|${attempt.trick.length}|${attempt.hand.length}|${assigned}|${mine}`
	);
	let lastKey: string | undefined;
	$effect(() => {
		// Not on the first run: a fixture may start with a card picked.
		if (lastKey !== undefined && lastKey !== selectionKey) {
			selected = null;
			selectedTasks = [];
			shown = null;
		}
		lastKey = selectionKey;
	});

	// Close the give-up dialog once the attempt is over.
	$effect(() => {
		if (!playing) confirmGiveUp = false;
	});
	// The communicate flow ends when the window closes (the leader played, the attempt ended).
	const canCommunicate = $derived(attempt.actions.communicate.length > 0);
	$effect(() => {
		if (!canCommunicate) {
			communicating = false;
			shown = null;
		}
	});

	const inPlayable = (c: Card) => attempt.playable.some((p) => sameCard(p, c));

	const showable = (c: Card) => attempt.actions.communicate.some((o) => sameCard(o.card, c));

	function handEnabled(c: Card): boolean {
		if (!ready || !playing) return false;
		if (step === 'play' && communicating) return showable(c);
		if (step === 'play') return mine && !forJarvis && inPlayable(c);
		if (step === 'pass') return !!attempt.pass && !attempt.pass.chosen && c.suit !== 'trump';
		return false;
	}
	function jarvisEnabled(c: Card): boolean {
		return ready && playing && step === 'play' && !communicating && forJarvis && inPlayable(c);
	}
	function pickCard(c: Card) {
		if (communicating) {
			shown = sameCard(c, shown) ? null : c;
			return;
		}
		// Tapping the selected card again plays it.
		if (step === 'play' && sameCard(c, selected)) {
			void run(() => client.playCard(c));
			return;
		}
		selected = sameCard(c, selected) ? null : c;
	}
	function startCommunicating() {
		communicating = true;
		shown = null;
		selected = null;
	}
	function stopCommunicating() {
		communicating = false;
		shown = null;
	}
	async function communicate(o: CommOption) {
		if (await run(() => client.communicate(o.card, o.position))) stopCommunicating();
	}

	// Deep Sea: before the first choice the captain may tap a displayed task to replace it. On their
	// own draft turn a tapped task offers both Take and Replace; otherwise (also in mission 25, where
	// the captain drafts last) a tap only selects it for replacing.
	const replacing = $derived(attempt.actions.replaceTasks && (step !== 'draft' || !mine));
	function taskEnabled(t: Task): boolean {
		if (!ready) return false;
		if (replacing) return t.owner === OWNER_UNASSIGNED;
		if (!mine) return false;
		if (step === 'draft' || step === 'assign') return t.owner === OWNER_UNASSIGNED;
		if (step === 'adjustTokens') return true;
		return false;
	}
	function pickTask(t: Task) {
		// Like a card: tap to select, tap again (or confirm in the decision panel) to take it.
		if (step === 'draft' && !replacing) {
			if (selectedTasks[0] !== t.id) selectedTasks = [t.id];
			else if (attempt.takeable.includes(t.id)) void run(() => client.pickTask(t.id));
			else selectedTasks = [];
			return;
		}
		if (replacing || step === 'assign') {
			selectedTasks = selectedTasks[0] === t.id ? [] : [t.id];
			return;
		}
		selectedTasks = selectedTasks.includes(t.id)
			? selectedTasks.filter((id) => id !== t.id)
			: [...selectedTasks, t.id].slice(-2);
	}

	const taskPick: TaskPick = {
		canPick: taskEnabled,
		get selected() {
			return selectedTasks;
		},
		onpick: pickTask
	};
	// During the setup the open tasks lie in the middle, where the trick goes later.
	const pool = $derived(
		step !== 'play' && step !== '' && attempt.tasks.some((t) => t.owner === OWNER_UNASSIGNED)
	);

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
	// The decision panel: every step but play, and in play the distress signal and the handover.
	const decision = $derived(
		step !== 'play' || attempt.actions.sendDistress || attempt.actions.handover
	);
</script>

<div class="mx-auto flex w-full max-w-md flex-1 flex-col gap-2 sm:max-w-xl">
	<header class={cn(headerRow, 'text-sm')}>
		<HeaderButton icon={House} label="Home" href={resolve('/')} />
		<span class="font-mono text-base font-semibold tracking-widest">{view.gameId}</span>
		<span
			class={cn('size-2 rounded-full', connected ? 'bg-success' : 'animate-pulse bg-amber-500')}
			title={connected ? 'Connected' : 'Reconnecting'}
		></span>
		<span class="ml-1.5 truncate font-display text-base font-bold">Mission {view.mission}</span>
		<HeaderButton
			class="ml-auto"
			icon={BookOpen}
			label="Rules"
			onclick={() => (rulesOpen = true)}
		/>
		{#if playing}
			<HeaderButton
				icon={Flag}
				label="Give up"
				title="Give up this attempt"
				disabled={!connected}
				onclick={() => (confirmGiveUp = true)}
			/>
		{/if}
		<HeaderThemeToggle />
	</header>

	<CrewStrip
		{attempt}
		online={client.online}
		{jarvisEnabled}
		{selected}
		onpick={pickCard}
		{taskPick}
	/>

	<MissionPanel {attempt} {mission} {roles} />

	{#if pool}
		<TaskPool {attempt} pick={taskPick} />
	{:else if step === 'play' || step === ''}
		<TrickArea {attempt} />
	{/if}

	<!-- In front of you: your row, the button row, your hand. -->
	<div class="mt-auto flex flex-col gap-2 pb-[env(safe-area-inset-bottom)]">
		{#if exchangeText}
			<p class="text-center text-xs text-muted-foreground">{exchangeText}</p>
		{/if}

		{#if me}
			<MemberRow {attempt} member={me} {taskPick} />
		{/if}

		{#if playing}
			{#if decision}
				<DecisionPanel
					{client}
					{attempt}
					{mission}
					selectedCard={selected}
					{selectedTasks}
					{busy}
					{connected}
					{run}
					bind:roles
				/>
			{/if}
			{#if step === 'play'}
				<ActionRow
					{attempt}
					{selected}
					{communicating}
					{shown}
					disabled={!ready}
					onplay={(c) => run(() => client.playCard(c))}
					oncommunicate={startCommunicating}
					oncancel={stopCommunicating}
					onshow={communicate}
				/>
			{/if}
			<Hand
				cards={attempt.hand}
				deal={`${view.gameId}:${view.mission}:${attempt.number}`}
				selected={communicating ? shown : selected}
				marked={attempt.pass?.chosen ? [attempt.pass.chosen] : []}
				enabled={handEnabled}
				{dealing}
				onpick={pickCard}
			/>
		{:else}
			<Debrief {client} {view} {attempt} {busy} {connected} {run} />
		{/if}
	</div>
</div>

<RulesHelp bind:open={rulesOpen} {mission} />

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
