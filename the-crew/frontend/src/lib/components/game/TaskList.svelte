<!--
	A row of tasks: the ones in front of a member. Planet Nine's are small playing cards, Deep Sea's
	task cards are chips (TaskChip; a tap opens the full card). A task that changes place (taken in
	the draft, handed over) flies from one row to the other.
-->
<script lang="ts" module>
	import type { Task } from '#lib/types.js';

	/** Tapping tasks (draft, token setup): which can be tapped, which are selected, what a tap does. */
	export interface TaskPick {
		canPick: (t: Task) => boolean;
		selected: number[];
		onpick: (t: Task) => void;
		/** Deep Sea: the open tasks read out the selected one (draft, assignment). */
		read?: boolean;
	}
</script>

<script lang="ts">
	import { flip } from 'svelte/animate';
	import { getCardMotion } from '@board-games/shared/cardMotion.js';
	import { isTaskCard } from '#lib/game/taskChip.js';
	import { cn } from '#lib/utils.js';
	import TaskCard from './TaskCard.svelte';
	import TaskChip from './TaskChip.svelte';

	let {
		tasks,
		attempt,
		size = 'xs',
		pick,
		label,
		players,
		class: className
	}: {
		/** The tasks to show, already in token order. */
		tasks: Task[];
		/** The attempt number: tasks of a new attempt don't fly from the old ones. */
		attempt: number;
		size?: 'xs' | 'task';
		pick?: TaskPick;
		label: string;
		/** Deep Sea: the crew members (the dummy included), for the numbers in the card texts. */
		players?: number;
		class?: string;
	} = $props();

	const motion = getCardMotion();
	const { send, receive } = motion;
	const key = (t: Task) => `task-${attempt}-${t.id}`;
	const chips = $derived(tasks.some(isTaskCard));
</script>

{#if tasks.length}
	<ul
		class={cn(
			'flex flex-wrap items-end',
			chips ? 'gap-x-2 gap-y-2' : size === 'xs' ? 'gap-x-[7px] gap-y-1.5' : 'gap-x-2.5 gap-y-2',
			className
		)}
		aria-label={label}
	>
		{#each tasks as t (t.id)}
			<li
				class={chips ? 'max-w-full min-w-0' : undefined}
				animate:flip={motion.flip}
				in:receive={{ key: key(t), offstage: 'none' }}
				out:send={{ key: key(t), offstage: 'none' }}
			>
				{#if chips}
					<TaskChip
						task={t}
						players={players ?? 3}
						selected={!!pick?.selected.includes(t.id)}
						onpick={pick?.canPick(t) ? () => pick.onpick(t) : undefined}
					/>
				{:else}
					<TaskCard
						task={t}
						{size}
						{players}
						selected={!!pick?.selected.includes(t.id)}
						onpick={pick?.canPick(t) ? () => pick.onpick(t) : undefined}
					/>
				{/if}
			</li>
		{/each}
	</ul>
{/if}
