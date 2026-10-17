<script lang="ts">
	import ArrowLeft from '@lucide/svelte/icons/arrow-left';
	import ArrowRight from '@lucide/svelte/icons/arrow-right';
	import LoaderCircle from '@lucide/svelte/icons/loader-circle';
	import Shuffle from '@lucide/svelte/icons/shuffle';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import { cap, terms } from '#lib/edition.js';
	import type { GameClient } from '#lib/game/client.js';
	import { byTokenOrder, cardLabel } from '#lib/game/cards.js';
	import { listText, objectiveText, seatName, stepText } from '#lib/game/format.js';
	import { taskTextOf } from '#lib/game/taskText.js';
	import { cn } from '#lib/utils.js';
	import type { Attempt, Card, Direction, MissionDef, Task } from '#lib/types.js';
	import EditionGlyph from './EditionGlyph.svelte';
	import TaskCard from './TaskCard.svelte';

	let {
		client,
		attempt,
		mission,
		selectedCard,
		selectedTasks,
		busy,
		connected,
		run,
		roles = $bindable([])
	}: {
		client: GameClient;
		attempt: Attempt;
		mission: MissionDef | null;
		/** The card picked in the hand or among the dummy's stacks. */
		selectedCard: Card | null;
		/** Tasks tapped in the pool or a member's row (draft, token swap or move). */
		selectedTasks: number[];
		busy: boolean;
		connected: boolean;
		/** Runs an action; false when it failed (the error is shown already). */
		run: (action: () => Promise<void>) => Promise<boolean>;
		/** Mission 50: the seats tapped for the two roles (null = open), not yet confirmed. */
		roles?: (number | null)[];
	} = $props();

	const me = $derived(attempt.seat);
	const name = (s: number) => seatName(attempt.members, s, me);
	const isCommander = $derived(me === attempt.commander);
	const mine = $derived(attempt.actors.includes(me));
	const disabled = $derived(busy || !connected);
	const dummy = $derived(attempt.members.find((m) => m.dummy));
	const jarvisTurn = $derived(!!dummy && attempt.turn === dummy.seat);
	const forJarvis = $derived(jarvisTurn && mine);
	const players = $derived(attempt.members.length);
	// Deep Sea: the task whose owner predicts now, and the task tapped in the pool.
	const predictTask = $derived(attempt.predict >= 0 ? attempt.tasks[attempt.predict] : null);
	const forDummyPrediction = $derived(!!dummy && predictTask?.owner === dummy.seat);
	const tappedTask = $derived(selectedTasks.length ? attempt.tasks[selectedTasks[0]] : null);
	const twoVolunteers = $derived(attempt.pick === 'volunteers' && (mission?.volunteers ?? 0) > 1);

	/** The next human seat in direction d (left = the next seat), skipping the dummy. */
	function neighbour(d: Direction): number {
		const n = attempt.members.length;
		const step = d === 'left' ? 1 : n - 1;
		let s = (me + step) % n;
		while (attempt.members[s].dummy) s = (s + step) % n;
		return s;
	}

	const waitingFor = $derived(
		listText(
			attempt.actors.map((s) => (jarvisTurn ? `${name(s)} (for ${terms.dummy})` : name(s))),
			'or'
		)
	);

	const status = $derived.by(() => {
		const step = attempt.step;
		switch (step) {
			case 'play':
				if (!mine) return `Waiting for ${waitingFor} to play.`;
				if (forJarvis) return `${terms.dummyPossessive} turn: tap one of its cards.`;
				return attempt.trick.length
					? 'Your turn: follow suit if you can.'
					: 'Your turn: lead a card.';
			case 'draft': {
				if (!mine) return `Waiting for ${waitingFor} to take a task.`;
				if (draftTask && !attempt.takeable.includes(draftTask.id))
					return 'This task cannot be taken now: tap another one.';
				const or = attempt.actions.pass ? ', or pass' : '';
				return forJarvis
					? `Choose a task for ${terms.dummy}: tap it${or}.`
					: `Take a task: tap it${or}.`;
			}
			case 'assign':
				if (!mine)
					return `${stepText(step, attempt.pick, mission)} Waiting for ${name(attempt.commander)}.`;
				return tappedTask ? 'Who takes this task?' : 'Tap a task, then choose who takes it.';
			case 'predict':
				if (!mine) return `Waiting for ${waitingFor} to predict.`;
				return forDummyPrediction
					? `How many tricks will ${terms.dummy} win?`
					: 'How many tricks will you win?';
			case 'pass':
				return attempt.pass?.chosen
					? `Waiting for ${waitingFor} to choose a card.`
					: stepText(step, attempt.pick);
			case 'poll':
				if (mine) return `The ${terms.commander} asks you.`;
				return isCommander ? 'Waiting for the crew to answer.' : `Waiting for ${waitingFor}.`;
			case 'pick':
				if (mine && twoVolunteers)
					return chosen.length < 2 ? 'Pick the two volunteers.' : 'Confirm the volunteers.';
				if (mine && attempt.pick === 'roles')
					return roles[0] == null
						? 'Pick who wins only the first four tricks.'
						: roles[1] == null
							? 'Now pick who wins only the last trick.'
							: 'Confirm the roles.';
				return mine
					? stepText(step, attempt.pick, mission)
					: `${stepText(step, attempt.pick, mission)} Waiting for ${name(attempt.commander)}.`;
			case 'adjustTokens':
				return mine
					? stepText(step, attempt.pick)
					: `${stepText(step, attempt.pick)} Waiting for ${name(attempt.commander)}.`;
			default:
				return '';
		}
	});

	// --- Poll and pick ---
	const poll = $derived(attempt.poll);
	const pollTask = $derived(poll && poll.task >= 0 ? attempt.tasks[poll.task] : null);
	const pollObjective = $derived(
		poll && poll.objective >= 0 ? attempt.objectives[poll.objective] : null
	);
	const question = $derived.by(() => {
		if (!poll) return '';
		if (poll.kind === 'wellIll') return 'How do you feel: well or ill?';
		if (pollObjective) return `Could you take this objective: ${objectiveText(pollObjective)}?`;
		if (pollTask?.card) return `Could you take ${cardLabel(pollTask.card)}?`;
		if (mission?.distribution === 'volunteers' && mission.volunteers > 1)
			return `Would you volunteer to share the ${attempt.tasks.length} tasks with another member?`;
		return `Could you take all ${attempt.tasks.length} tasks?`;
	});
	const answerText = (a: boolean | null | undefined) =>
		a == null ? '…' : poll?.kind === 'wellIll' ? (a ? 'well' : 'ill') : a ? 'yes' : 'no';
	const pollers = $derived(
		attempt.members.filter((m) => !m.dummy && m.seat !== attempt.commander).map((m) => m.seat)
	);

	// A derived string changes only with the pick itself, not with every new view object.
	const pickKey = $derived(`${attempt.step}|${attempt.pick}`);
	$effect(() => {
		void pickKey;
		roles = [];
		chosen = [];
	});
	/** Mission 26: the two volunteers tapped, not yet confirmed. */
	let chosen = $state<number[]>([]);
	const pickedRoles = $derived(roles[0] != null && roles[1] != null ? [roles[0], roles[1]] : null);
	/** The seats the commander may pick now (the server applies the rules, e.g. the even split). */
	const pickable = $derived(attempt.pickable);
	function pickSeat(s: number) {
		if (twoVolunteers) {
			chosen = chosen.includes(s) ? chosen.filter((c) => c !== s) : [...chosen, s].slice(-2);
			return;
		}
		if (attempt.pick !== 'roles') {
			void run(() => client.pick([s]));
			return;
		}
		// Each role keeps its seat: tapping a picked seat frees only its role, a new seat fills the
		// first open role (or replaces the last trick's).
		const next = [roles[0] ?? null, roles[1] ?? null];
		const slot = next.indexOf(s);
		if (slot >= 0) next[slot] = null;
		else next[next[0] == null ? 0 : 1] = s;
		roles = next;
	}

	// --- Draft: a tapped task is taken only on confirmation, like a played card ---
	const draftTask = $derived(
		attempt.step === 'draft' && selectedTasks.length ? attempt.tasks[selectedTasks[0]] : null
	);

	// --- Token swap / move ---
	const moveToken = $derived(mission?.setup.includes('moveToken') ?? false);
	const tokenReady = $derived(selectedTasks.length === 2);

	// --- Handover (5 players, golden frame) ---
	let handoverOpen = $state(false);
	let handoverTask = $state<Task | null>(null);
	$effect(() => {
		if (!attempt.actions.handover) handoverOpen = false;
	});
	async function handover(to: number) {
		const t = handoverTask;
		if (!t) return;
		if (await run(() => client.handover(t.id, to))) {
			handoverOpen = false;
			handoverTask = null;
		}
	}
</script>

<div
	class="flex min-h-14 flex-col justify-center gap-2 rounded-xl bg-muted/60 px-3 py-2"
	role="status"
>
	{#if tappedTask?.objective}
		<!-- Deep Sea: the tapped task card in words. -->
		<p class="text-sm leading-snug">{taskTextOf(tappedTask, players)}</p>
	{/if}
	<!-- The status; a tapped card or task replaces it with the button that confirms it. -->
	{#if attempt.step === 'play' && mine && selectedCard}
		<Button
			class="h-10 w-full text-base"
			{disabled}
			onclick={() => run(() => client.playCard(selectedCard))}
		>
			Play {cardLabel(selectedCard)}{forJarvis ? ` for ${terms.dummy}` : ''}
		</Button>
	{:else if attempt.step === 'draft' && mine && draftTask && (attempt.takeable.includes(draftTask.id) || attempt.actions.replaceTasks)}
		<div class="flex gap-2">
			{#if attempt.takeable.includes(draftTask.id)}
				<Button
					class="h-10 flex-1 text-base"
					{disabled}
					onclick={() => run(() => client.pickTask(draftTask.id))}
				>
					Take {draftTask.card ? cardLabel(draftTask.card) : 'this task'}{forJarvis
						? ` for ${terms.dummy}`
						: ''}
				</Button>
			{:else}
				<p class="mr-auto self-center text-sm text-muted-foreground">You cannot take this task.</p>
			{/if}
			{#if attempt.actions.replaceTasks}
				<Button
					class="h-10"
					variant="secondary"
					{disabled}
					onclick={() => run(() => client.replaceTask(draftTask.id))}
				>
					<Shuffle />Replace
				</Button>
			{/if}
		</div>
	{:else}
		<p
			class={cn(
				'flex items-center gap-2 leading-snug',
				mine ? 'font-semibold' : 'text-muted-foreground'
			)}
		>
			{#if !mine}<LoaderCircle class="size-4 shrink-0 animate-spin" />{/if}
			{status}
		</p>
	{/if}

	{#if attempt.step === 'draft' && mine && attempt.actions.pass && !draftTask}
		<Button
			class="h-10"
			variant="secondary"
			{disabled}
			onclick={() => run(() => client.passTask())}
		>
			Pass{forJarvis ? ` for ${terms.dummy}` : ''}
		</Button>
	{/if}
	{#if attempt.actions.replaceTasks && attempt.step !== 'draft'}
		<!-- Before the first choice the captain may swap a displayed task for another of the same difficulty. -->
		<div class="flex items-center gap-2 text-sm">
			<span class="mr-auto text-muted-foreground">
				{tappedTask ? 'Replace the tapped task?' : 'Tap a task to replace it.'}
			</span>
			<Button
				size="sm"
				variant="secondary"
				disabled={disabled || !tappedTask}
				onclick={() => tappedTask && run(() => client.replaceTask(tappedTask.id))}
			>
				<Shuffle />Replace
			</Button>
		</div>
	{/if}

	{#if attempt.step === 'assign' && mine && tappedTask}
		<div class="flex flex-wrap items-center gap-2">
			{#each attempt.assignable[tappedTask.id] ?? [] as s (s)}
				<Button
					class="h-9"
					variant="secondary"
					{disabled}
					onclick={() => run(() => client.assignTask(tappedTask.id, s))}
				>
					{name(s)}
				</Button>
			{/each}
		</div>
	{:else if attempt.step === 'predict' && mine && predictTask}
		<p class="text-sm">
			{predictTask.objective?.hidden
				? 'Only you will see the number until the debrief.'
				: 'Everyone will see the number.'}
		</p>
		<div class="grid grid-cols-7 gap-1.5">
			{#each Array.from({ length: attempt.totalTricks + 1 }, (_, i) => i) as n (n)}
				<Button
					class="h-10 px-0"
					variant="secondary"
					{disabled}
					onclick={() => run(() => client.predict(predictTask.id, n))}
				>
					{n}
				</Button>
			{/each}
		</div>
	{:else if attempt.step === 'poll' && poll}
		<p class="text-sm">{question}</p>
		<ul class="flex flex-wrap gap-x-3 gap-y-0.5 text-sm">
			{#each pollers as s (s)}
				<li>
					<span class="text-muted-foreground">{name(s)}:</span>
					<b>{answerText(poll.answers[s])}</b>
				</li>
			{/each}
		</ul>
		{#if mine}
			<div class="grid grid-cols-2 gap-2">
				<Button class="h-11" {disabled} onclick={() => run(() => client.answerPoll(true))}>
					{poll.kind === 'wellIll' ? 'Well' : 'Yes'}
				</Button>
				<Button
					class="h-11"
					variant="secondary"
					{disabled}
					onclick={() => run(() => client.answerPoll(false))}
				>
					{poll.kind === 'wellIll' ? 'Ill' : 'No'}
				</Button>
			</div>
		{/if}
	{:else if attempt.step === 'pick' && mine}
		{#if pollTask}
			<div class="flex items-center gap-2 text-sm">
				<TaskCard task={pollTask} />
				<span>Who takes this task?</span>
			</div>
		{:else if pollObjective}
			<p class="text-sm">{objectiveText(pollObjective)}</p>
		{/if}
		<div class="flex flex-wrap items-center gap-2">
			{#each pickable as s (s)}
				{@const answer = poll?.answers[s]}
				<!-- A tapped role shows in the mission panel's objectives. -->
				<Button
					class="h-9 gap-1.5"
					variant={roles.includes(s) || chosen.includes(s) ? 'default' : 'secondary'}
					{disabled}
					onclick={() => pickSeat(s)}
				>
					{name(s)}
					{#if answer != null}
						<span class="text-xs font-normal opacity-75">said {answerText(answer)}</span>
					{/if}
				</Button>
			{/each}
			{#if attempt.pick === 'roles'}
				<Button
					class="ml-auto h-9"
					disabled={disabled || !pickedRoles}
					onclick={() => pickedRoles && run(() => client.pick(pickedRoles))}
				>
					Confirm
				</Button>
			{:else if twoVolunteers}
				<Button
					class="ml-auto h-9"
					disabled={disabled || chosen.length < 2}
					onclick={() => run(() => client.pick(chosen))}
				>
					Confirm
				</Button>
			{/if}
		</div>
	{:else if attempt.step === 'adjustTokens' && mine}
		<p class="text-sm text-muted-foreground">
			{moveToken
				? 'Tap a task with a token, then a task without one.'
				: 'Tap two tasks to swap their tokens.'}
		</p>
		<div class="grid grid-cols-2 gap-2">
			<Button
				class="h-11"
				disabled={disabled || !tokenReady}
				onclick={() => run(() => client.adjustTokens(selectedTasks[0], selectedTasks[1]))}
			>
				{moveToken ? 'Move token' : 'Swap tokens'}
			</Button>
			<Button
				class="h-11"
				variant="secondary"
				{disabled}
				onclick={() => run(() => client.skipTokens())}
			>
				Keep as is
			</Button>
		</div>
	{:else if attempt.step === 'pass' && attempt.pass && !attempt.pass.chosen}
		{@const to = name(neighbour(attempt.pass.direction))}
		<p class="text-sm">
			Choose a card to pass {attempt.pass.direction} to {to}. {cap(terms.trumps)} can't be passed.
		</p>
		<Button
			class="h-11"
			disabled={disabled || !selectedCard}
			onclick={() => selectedCard && run(() => client.choosePassCard(selectedCard))}
		>
			{selectedCard ? `Pass ${cardLabel(selectedCard)}` : 'Pick a card in your hand'}
		</Button>
	{/if}

	{#if attempt.step === 'play' && attempt.actions.sendDistress}
		<div class="flex flex-wrap items-center gap-1.5 border-t border-foreground/10 pt-2 text-sm">
			<span class="mr-auto flex items-center gap-1.5 text-muted-foreground">
				<EditionGlyph kind="distress" class="size-4 shrink-0" />Distress signal?
			</span>
			<Button
				size="sm"
				variant="secondary"
				{disabled}
				onclick={() => run(() => client.sendDistress('left'))}
			>
				<ArrowLeft class="size-4" />Left
			</Button>
			<Button
				size="sm"
				variant="secondary"
				{disabled}
				onclick={() => run(() => client.sendDistress('right'))}
			>
				Right<ArrowRight class="size-4" />
			</Button>
			<Button size="sm" variant="ghost" {disabled} onclick={() => run(() => client.skipDistress())}>
				No
			</Button>
		</div>
	{/if}
	{#if attempt.step === 'play' && attempt.actions.handover}
		<Button size="sm" variant="outline" {disabled} onclick={() => (handoverOpen = true)}>
			Hand over a task
		</Button>
	{/if}
</div>

<Dialog.Root bind:open={handoverOpen}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Hand over a task</Dialog.Title>
			<Dialog.Description>
				Once per attempt, before the first card: one task, with its token, goes to another member.
			</Dialog.Description>
		</Dialog.Header>
		<ul class="flex flex-wrap justify-center gap-2.5 py-1">
			{#each byTokenOrder(attempt.tasks) as t (t.id)}
				<li>
					<TaskCard
						task={t}
						owner={name(t.owner)}
						selected={handoverTask?.id === t.id}
						onpick={() => (handoverTask = t)}
					/>
				</li>
			{/each}
		</ul>
		{#if handoverTask}
			{@const from = handoverTask.owner}
			<p class="text-sm">Give it to:</p>
			<div class="grid grid-cols-2 gap-2">
				{#each attempt.members.filter((m) => m.seat !== from) as m (m.seat)}
					<Button class="h-11" variant="secondary" {disabled} onclick={() => handover(m.seat)}>
						{name(m.seat)}
					</Button>
				{/each}
			</div>
		{/if}
	</Dialog.Content>
</Dialog.Root>
