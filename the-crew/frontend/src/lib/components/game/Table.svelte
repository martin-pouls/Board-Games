<script lang="ts">
	import HeaderButton, { headerRow } from '@board-games/shared/HeaderButton.svelte';
	import HeaderThemeToggle from '@board-games/shared/HeaderThemeToggle.svelte';
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
	import { cap, terms } from '#lib/edition.js';
	import { errorText } from '#lib/errors.js';
	import { ActionError } from '#lib/game/connection.svelte.js';
	import type { GameClient } from '#lib/game/client.js';
	import { byTokenOrder, cardLabel, sameCard } from '#lib/game/cards.js';
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
	import TaskList, { type TaskPick } from './TaskList.svelte';
	import TaskPool from './TaskPool.svelte';
	import TrickArea from './TrickArea.svelte';

	let { client, view, attempt }: { client: GameClient; view: PlayerView; attempt: Attempt } =
		$props();

	setCardMotion();

	let selected = $state<Card | null>(null);
	let selectedTasks = $state<number[]>([]);
	let roles = $state<(number | null)[]>([]);
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
	const dealing = $derived(
		attempt.trickNumber === 1 && attempt.trick.length === 0 && step !== 'play'
	);

	// Drop the selection whenever the decision it belongs to changes.
	const assigned = $derived(attempt.tasks.filter((t) => t.owner !== OWNER_UNASSIGNED).length);
	const selectionKey = $derived(
		`${attempt.number}|${step}|${attempt.pick}|${attempt.trickNumber}|${attempt.trick.length}|${attempt.hand.length}|${assigned}|${mine}`
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
	const myTasks = $derived(byTokenOrder(attempt.tasks.filter((t) => t.owner === attempt.seat)));
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
	const canCommunicate = $derived(attempt.actions.communicate.length > 0 && ready);
	// Rapture of the deep lets a member communicate again while sonar tokens are left.
	const showCommunicate = $derived(
		playing &&
			step === 'play' &&
			attempt.comm !== 'none' &&
			attempt.seat !== attempt.silenced &&
			(!me?.communicated || attempt.actions.communicate.length > 0)
	);
	const myComms = $derived(
		me?.communications.length ? me.communications : me?.communication ? [me.communication] : []
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

	<div
		class="sticky bottom-0 -mx-3 mt-auto flex flex-col gap-1.5 px-3 pt-1 pb-[max(0.5rem,env(safe-area-inset-bottom))] before:absolute before:inset-0 before:-z-10 before:backdrop-blur"
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
				bind:roles
			/>
		{:else}
			<Debrief {client} {view} {attempt} {busy} {connected} {run} />
		{/if}

		{#if exchangeText}
			<p class="text-center text-xs text-muted-foreground">{exchangeText}</p>
		{/if}

		<!-- In front of you: your tasks, and the button to communicate. -->
		<div class="flex min-h-8 flex-wrap items-center gap-x-2 gap-y-1 text-xs text-muted-foreground">
			<span class="flex items-center gap-1">
				You
				{#if attempt.seat === attempt.commander}
					<Crown class="size-3.5 text-captain" aria-label={cap(terms.commander)} />
				{/if}
				{#if attempt.seat === attempt.silenced}
					<MicOff class="size-3.5 text-danger" aria-label="Silenced" />
				{/if}
			</span>
			<span class="flex items-center gap-0.5" title="Tricks won">
				<Trophy class="size-3.5" />{me?.tricksWon ?? 0}
			</span>
			{#each myComms as c (cardLabel(c.card))}
				<span class="flex items-center gap-1">
					<CommBadge comm={c} />
					{cardLabel(c.card)}
				</span>
			{/each}
			<TaskList
				tasks={myTasks}
				attempt={attempt.number}
				players={attempt.members.length}
				pick={taskPick}
				label="Your tasks"
				class="pt-1"
			/>
			{#if showCommunicate}
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

		{#if playing}
			<Hand
				cards={attempt.hand}
				deal={`${view.gameId}:${view.mission}:${attempt.number}`}
				{selected}
				marked={attempt.pass?.chosen ? [attempt.pass.chosen] : myComms.map((c) => c.card)}
				enabled={handEnabled}
				{dealing}
				onpick={pickCard}
			/>
		{/if}
	</div>
</div>

<RulesHelp bind:open={rulesOpen} {mission} />

<CommunicateSheet
	bind:open={commOpen}
	options={attempt.actions.communicate}
	deadZone={(mission?.deadZone ?? false) || attempt.comm === 'currents'}
	tokensLeft={attempt.comm === 'rapture' ? attempt.tokensLeft : -1}
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
