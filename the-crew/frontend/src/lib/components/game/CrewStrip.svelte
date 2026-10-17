<script lang="ts">
	import Crown from '@lucide/svelte/icons/crown';
	import Layers from '@lucide/svelte/icons/layers';
	import MicOff from '@lucide/svelte/icons/mic-off';
	import Trophy from '@lucide/svelte/icons/trophy';
	import { getCardMotion } from '@board-games/shared/cardMotion.js';
	import { byTokenOrder, cardKey, sameCard } from '#lib/game/cards.js';
	import { cap, terms } from '#lib/edition.js';
	import { DUMMY } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import type { Attempt, Card, Member } from '#lib/types.js';
	import CardBack from './CardBack.svelte';
	import CommBadge from './CommBadge.svelte';
	import EditionGlyph from './EditionGlyph.svelte';
	import PlayingCard from './PlayingCard.svelte';
	import TaskList, { type TaskPick } from './TaskList.svelte';

	let {
		attempt,
		online,
		jarvisEnabled = () => false,
		selected = null,
		onpick,
		taskPick
	}: {
		attempt: Attempt;
		/** Player IDs with a live connection. */
		online: string[];
		/** Whether the viewer can pick this dummy card now (the commander on the dummy's turn). */
		jarvisEnabled?: (c: Card) => boolean;
		selected?: Card | null;
		onpick?: (c: Card) => void;
		taskPick?: TaskPick;
	} = $props();

	const { send, receive } = getCardMotion();

	// The other members in seat order, starting left of the viewer.
	const others = $derived.by(() => {
		const n = attempt.members.length;
		const out: Member[] = [];
		for (let i = 1; i < n; i++) out.push(attempt.members[(attempt.seat + i) % n]);
		return out;
	});
	const humans = $derived(others.filter((m) => !m.dummy));
	const jarvis = $derived(others.find((m) => m.dummy));
	const turnSeat = $derived(
		attempt.step === 'play' || attempt.step === 'draft' ? attempt.turn : -1
	);
	const tasks = $derived(byTokenOrder(attempt.tasks));
	const tasksOf = (m: Member) => tasks.filter((t) => t.owner === m.seat);
	const anyJarvisEnabled = $derived(!!jarvis?.stacks.some((s) => s.up && jarvisEnabled(s.up)));
</script>

{#snippet stats(m: Member)}
	<span class="flex items-center gap-2 text-xs text-muted-foreground">
		<span class="flex items-center gap-0.5" title="Cards in hand">
			<Layers class="size-3.5" />{m.cardCount}
		</span>
		<span class="flex items-center gap-0.5" title="Tricks won">
			<Trophy class="size-3.5" />{m.tricksWon}
		</span>
	</span>
{/snippet}

{#snippet badges(m: Member)}
	{#if m.seat === attempt.commander}
		<Crown class="size-3.5 shrink-0 text-rocket-ink" aria-label={cap(terms.commander)} />
	{/if}
	{#if m.seat === attempt.silenced}
		<MicOff class="size-3.5 shrink-0 text-danger" aria-label="Silenced" />
	{/if}
{/snippet}

<section class="flex flex-col gap-1.5" aria-label="Crew">
	{#if humans.length}
		<ul class={cn('grid gap-1.5', humans.length < 3 ? 'grid-cols-1' : 'grid-cols-2')}>
			{#each humans as m (m.seat)}
				{@const up = online.includes(m.playerId)}
				<li
					class={cn(
						'flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 rounded-xl bg-panel px-2 py-1.5 ring-2 ring-transparent transition-shadow',
						(m.mustAct || turnSeat === m.seat) && 'ring-signal/70'
					)}
				>
					<span class="flex min-w-0 items-center gap-1 text-sm font-semibold">
						<span
							class={cn(
								'size-1.5 shrink-0 rounded-full',
								up ? 'bg-success' : 'bg-muted-foreground/40'
							)}
							title={up ? 'Online' : 'Offline'}
						></span>
						<span class="truncate">{m.name}</span>
						{@render badges(m)}
					</span>
					{@render stats(m)}
					{#if m.communication}
						{@const c = m.communication.card}
						<span class="relative shrink-0">
							{#key cardKey(c)}
								<span
									class="block"
									in:receive={{ key: cardKey(c), offstage: 'top' }}
									out:send={{ key: cardKey(c), offstage: 'top' }}
								>
									<PlayingCard card={c} size="xs" />
								</span>
							{/key}
							<CommBadge comm={m.communication} class="absolute -top-1.5 -right-1.5" />
						</span>
					{:else if m.communicated}
						<EditionGlyph
							kind="comm"
							class="size-4 shrink-0 text-muted-foreground/60"
							aria-label="Has communicated"
						/>
					{/if}
					<TaskList
						tasks={tasksOf(m)}
						attempt={attempt.number}
						pick={taskPick}
						label={`${m.name}'s tasks`}
						class="ml-auto pt-1"
					/>
				</li>
			{/each}
		</ul>
	{/if}

	{#if jarvis}
		<div
			class={cn(
				'flex flex-col gap-1 rounded-xl bg-panel px-2 py-1.5 ring-2 ring-transparent',
				(turnSeat === jarvis.seat || anyJarvisEnabled) && 'ring-signal/70'
			)}
		>
			<div class="flex min-h-8 flex-wrap items-center gap-x-2 gap-y-1">
				<span class="font-display text-sm font-semibold tracking-wider">{DUMMY}</span>
				{@render badges(jarvis)}
				{@render stats(jarvis)}
				<TaskList
					tasks={tasksOf(jarvis)}
					attempt={attempt.number}
					pick={taskPick}
					label={`${terms.dummyPossessive} tasks`}
					class="ml-auto pt-1"
				/>
			</div>
			<ul class="flex justify-between gap-1" aria-label={`${terms.dummyPossessive} stacks`}>
				{#each jarvis.stacks as s, i (i)}
					{@const on = !!s.up && jarvisEnabled(s.up)}
					<li class="relative h-[4.1rem] w-[2.6rem] shrink-0">
						{#if s.hidden}
							<CardBack
								size="stack"
								label="Hidden card"
								class={cn('absolute top-0 left-0', s.up ? 'translate-x-0.5 translate-y-1' : '')}
							/>
						{/if}
						{#if s.up}
							{@const c = s.up}
							{#key cardKey(c)}
								<button
									type="button"
									class={cn(
										'absolute top-0 left-0 rounded-md transition-[translate,opacity] duration-150 focus-visible:ring-3 focus-visible:ring-ring focus-visible:outline-none',
										sameCard(c, selected) && '-translate-y-1.5 ring-3 ring-signal',
										on && !sameCard(c, selected) && 'hover:-translate-y-1',
										anyJarvisEnabled && !on && 'opacity-40 saturate-50'
									)}
									disabled={!on}
									aria-pressed={sameCard(c, selected)}
									onclick={() => onpick?.(c)}
									in:receive={{ key: cardKey(c), offstage: 'none' }}
									out:send={{ key: cardKey(c), offstage: 'none' }}
								>
									<PlayingCard card={c} size="stack" />
								</button>
							{/key}
						{:else if !s.hidden}
							<span
								class="absolute top-0 left-0 h-[3.6rem] w-[2.6rem] rounded-md border border-dashed border-foreground/15"
							></span>
						{/if}
					</li>
				{/each}
			</ul>
		</div>
	{/if}
</section>
