<script lang="ts">
	import ArrowLeft from '@lucide/svelte/icons/arrow-left';
	import ArrowRight from '@lucide/svelte/icons/arrow-right';
	import LoaderCircle from '@lucide/svelte/icons/loader-circle';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import type { GameClient } from '#lib/game/client.js';
	import { cardLabel } from '#lib/game/cards.js';
	import { listText, objectiveText, seatName, stepText } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import type { Attempt, Card, Direction, MissionDef, Task } from '#lib/types.js';
	import TaskCard from './TaskCard.svelte';

	let {
		client,
		attempt,
		mission,
		selectedCard,
		selectedTasks,
		busy,
		connected,
		run
	}: {
		client: GameClient;
		attempt: Attempt;
		mission: MissionDef | null;
		/** The card picked in the hand or among JARVIS's stacks. */
		selectedCard: Card | null;
		/** Tasks tapped in the mission panel (token swap or move). */
		selectedTasks: number[];
		busy: boolean;
		connected: boolean;
		/** Runs an action; false when it failed (the error is shown already). */
		run: (action: () => Promise<void>) => Promise<boolean>;
	} = $props();

	const me = $derived(attempt.seat);
	const name = (s: number) => seatName(attempt.members, s, me);
	const isCommander = $derived(me === attempt.commander);
	const mine = $derived(attempt.actors.includes(me));
	const disabled = $derived(busy || !connected);
	const dummy = $derived(attempt.members.find((m) => m.dummy));
	const jarvisTurn = $derived(!!dummy && attempt.turn === dummy.seat);
	const forJarvis = $derived(jarvisTurn && mine);

	/** The next human seat in direction d (left = the next seat), skipping JARVIS. */
	function neighbour(d: Direction): number {
		const n = attempt.members.length;
		const step = d === 'left' ? 1 : n - 1;
		let s = (me + step) % n;
		while (attempt.members[s].dummy) s = (s + step) % n;
		return s;
	}

	const waitingFor = $derived(
		listText(
			attempt.actors.map((s) => (jarvisTurn ? `${name(s)} (for JARVIS)` : name(s))),
			'or'
		)
	);

	const status = $derived.by(() => {
		const step = attempt.step;
		switch (step) {
			case 'play':
				if (!mine) return `Waiting for ${waitingFor} to play.`;
				if (forJarvis) return "JARVIS's turn: tap one of its face-up cards.";
				return attempt.trick.length
					? 'Your turn: follow suit if you can.'
					: 'Your turn: lead a card.';
			case 'draft':
				if (!mine) return `Waiting for ${waitingFor} to take a task.`;
				return forJarvis ? 'Choose a task for JARVIS: tap it.' : 'Take a task: tap it.';
			case 'pass':
				return attempt.pass?.chosen
					? `Waiting for ${waitingFor} to choose a card.`
					: stepText(step, attempt.pick);
			case 'poll':
				if (mine) return 'The commander asks you.';
				return isCommander ? 'Waiting for the crew to answer.' : `Waiting for ${waitingFor}.`;
			case 'pick':
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
		return `Could you take all ${attempt.tasks.length} tasks?`;
	});
	const answerText = (a: boolean | null | undefined) =>
		a == null ? '…' : poll?.kind === 'wellIll' ? (a ? 'well' : 'ill') : a ? 'yes' : 'no';
	const pollers = $derived(
		attempt.members.filter((m) => !m.dummy && m.seat !== attempt.commander).map((m) => m.seat)
	);

	let roles = $state<number[]>([]);
	// A derived string changes only with the pick itself, not with every new view object.
	const pickKey = $derived(`${attempt.step}|${attempt.pick}`);
	$effect(() => {
		void pickKey;
		roles = [];
	});
	/** The seats the commander may pick now (the server applies the rules, e.g. the even split). */
	const pickable = $derived(attempt.pickable);
	function pickSeat(s: number) {
		if (attempt.pick !== 'roles') {
			void run(() => client.pick([s]));
			return;
		}
		roles = roles.includes(s) ? roles.filter((r) => r !== s) : [...roles, s].slice(-2);
	}

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
	<p
		class={cn(
			'flex items-center gap-2 leading-snug',
			mine ? 'font-semibold' : 'text-muted-foreground'
		)}
	>
		{#if !mine}<LoaderCircle class="size-4 shrink-0 animate-spin" />{/if}
		{status}
	</p>

	{#if attempt.step === 'play' && mine && selectedCard}
		<Button
			class="h-11 text-base"
			{disabled}
			onclick={() => run(() => client.playCard(selectedCard))}
		>
			Play {cardLabel(selectedCard)}{forJarvis ? ' for JARVIS' : ''}
		</Button>
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
		{#if attempt.pick === 'roles'}
			<p class="text-sm text-muted-foreground">
				{roles.length === 0
					? 'First pick the member who wins only the first four tricks.'
					: 'Then pick the member who wins only the last trick.'}
			</p>
		{/if}
		<div class="grid grid-cols-2 gap-2">
			{#each pickable as s (s)}
				{@const answer = poll?.answers[s]}
				{@const role = roles.indexOf(s)}
				<Button
					class={cn('h-auto min-h-11 flex-col gap-0 py-1', role >= 0 && 'ring-3 ring-signal')}
					variant={role >= 0 ? 'default' : 'secondary'}
					{disabled}
					onclick={() => pickSeat(s)}
				>
					<span>{name(s)}</span>
					{#if role >= 0}
						<span class="text-xs font-normal opacity-75">
							{role === 0 ? 'first four tricks' : 'last trick'}
						</span>
					{:else if answer != null}
						<span class="text-xs font-normal opacity-75">said {answerText(answer)}</span>
					{/if}
				</Button>
			{/each}
		</div>
		{#if attempt.pick === 'roles'}
			<Button
				class="h-11"
				disabled={disabled || roles.length !== 2}
				onclick={() => run(() => client.pick(roles))}
			>
				Confirm roles
			</Button>
		{/if}
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
			Choose a card to pass {attempt.pass.direction} to {to}. Rockets can't be passed.
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
			<span class="mr-auto text-muted-foreground">Distress signal?</span>
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
			{#each attempt.tasks as t (t.id)}
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
