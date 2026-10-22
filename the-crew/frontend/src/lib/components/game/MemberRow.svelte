<!--
	One crew member's row: name, cards in hand and tricks won, then the tasks in front of them in
	token order, and at the right end the communicated card (or a faint radio icon once its card is
	played). The same row for the other members (crew strip), JARVIS and the viewer (above the
	button row).
-->
<script lang="ts">
	import Crown from '@lucide/svelte/icons/crown';
	import Layers from '@lucide/svelte/icons/layers';
	import MicOff from '@lucide/svelte/icons/mic-off';
	import Trophy from '@lucide/svelte/icons/trophy';
	import { getCardMotion } from '@board-games/shared/cardMotion.js';
	import { cap, terms } from '#lib/edition.js';
	import { byTokenOrder, cardKey } from '#lib/game/cards.js';
	import { DUMMY } from '#lib/game/format.js';
	import { isTaskCard } from '#lib/game/taskChip.js';
	import { cn } from '#lib/utils.js';
	import type { Attempt, Member } from '#lib/types.js';
	import CommCard from './CommCard.svelte';
	import EditionGlyph from './EditionGlyph.svelte';
	import TaskList, { type TaskPick } from './TaskList.svelte';

	let {
		attempt,
		member: m,
		online = null,
		taskPick,
		class: className = ''
	}: {
		attempt: Attempt;
		member: Member;
		/** Whether the member is connected; null for the viewer and the dummy (no dot). */
		online?: boolean | null;
		taskPick?: TaskPick;
		class?: string;
	} = $props();

	const { send, receive } = getCardMotion();

	const you = $derived(m.seat === attempt.seat);
	const name = $derived(you ? 'You' : m.dummy ? DUMMY : m.name);
	const tasks = $derived(byTokenOrder(attempt.tasks.filter((t) => t.owner === m.seat)));
	/** Deep Sea's chips are wide: in a wrapping row (the 5-player grid) they move below the name. */
	const chips = $derived(tasks.some(isTaskCard));
	/** Rapture of the deep allows several communications; older views only have the first. */
	const comms = $derived(
		m.communications?.length ? m.communications : m.communication ? [m.communication] : []
	);
</script>

<!-- The dummy's stacks lie below its row, so its name and counts share one line. -->
<div class={cn('flex min-w-0 items-center gap-2', m.dummy ? 'min-h-7' : 'min-h-11', className)}>
	<div
		class={cn('flex min-w-13 shrink-0', m.dummy ? 'items-center gap-2.5 pr-1' : 'flex-col gap-px')}
	>
		<span
			class={cn(
				'flex items-center gap-1 text-sm leading-tight font-bold whitespace-nowrap',
				m.dummy && 'font-display tracking-wide'
			)}
		>
			{#if online !== null}
				<span
					class={cn(
						'size-1.5 shrink-0 rounded-full',
						online ? 'bg-success' : 'bg-muted-foreground/40'
					)}
					title={online ? 'Online' : 'Offline'}
				></span>
			{/if}
			<span class="max-w-24 truncate">{name}</span>
			{#if m.seat === attempt.commander}
				<Crown
					class="size-3.5 shrink-0 fill-current text-captain"
					aria-label={cap(terms.commander)}
				/>
			{/if}
			{#if m.seat === attempt.silenced}
				<MicOff class="size-3.5 shrink-0 text-danger" aria-label="Silenced" />
			{/if}
		</span>
		<span class="flex items-center gap-1.5 text-xs text-muted-foreground">
			<span class="flex items-center gap-0.5" title="Cards in hand">
				<Layers class="size-3" />{m.cardCount}
			</span>
			<span class="flex items-center gap-0.5" title="Tricks won">
				<Trophy class="size-3" />{m.tricksWon}
			</span>
		</span>
	</div>
	<div class={cn('flex min-w-0 flex-1 items-center gap-[7px]', chips && 'basis-auto')}>
		<TaskList
			{tasks}
			attempt={attempt.number}
			players={attempt.members.length}
			pick={taskPick}
			label={you ? 'Your tasks' : `${m.dummy ? terms.dummyPossessive : `${m.name}'s`} tasks`}
		/>
		{#if comms.length}
			<span class="mr-[7px] ml-auto flex shrink-0 gap-3">
				{#each comms as comm (cardKey(comm.card))}
					{#if you}
						<CommCard {comm} />
					{:else}
						{@const c = comm.card}
						<span
							class="block"
							in:receive={{ key: cardKey(c), offstage: 'top' }}
							out:send={{ key: cardKey(c), offstage: 'top' }}
						>
							<CommCard {comm} />
						</span>
					{/if}
				{/each}
			</span>
		{:else if m.communicated}
			<EditionGlyph
				kind="comm"
				class="ml-auto size-4 shrink-0 text-muted-foreground/60"
				aria-label="Has communicated"
			/>
		{/if}
	</div>
</div>
