<!--
	A row of task cards: the ones in front of a member, or the open pool during the setup. A task
	that changes place (taken in the draft, handed over) flies from one row to the other.
-->
<script lang="ts" module>
	import type { Task } from '#lib/types.js';

	/** Tapping tasks (draft, token setup): which can be tapped, which are selected, what a tap does. */
	export interface TaskPick {
		canPick: (t: Task) => boolean;
		selected: number[];
		onpick: (t: Task) => void;
	}
</script>

<script lang="ts">
	import { flip } from 'svelte/animate';
	import { getCardMotion } from '@board-games/shared/cardMotion.js';
	import { cn } from '#lib/utils.js';
	import TaskCard from './TaskCard.svelte';

	let {
		tasks,
		attempt,
		size = 'xs',
		pick,
		label,
		class: className
	}: {
		/** The tasks to show, already in token order. */
		tasks: Task[];
		/** The attempt number: tasks of a new attempt don't fly from the old ones. */
		attempt: number;
		size?: 'xs' | 'task';
		pick?: TaskPick;
		label: string;
		class?: string;
	} = $props();

	const motion = getCardMotion();
	const { send, receive } = motion;
	const key = (t: Task) => `task-${attempt}-${t.id}`;
</script>

{#if tasks.length}
	<ul
		class={cn('flex flex-wrap items-end', size === 'xs' ? 'gap-2' : 'gap-x-2.5 gap-y-2', className)}
		aria-label={label}
	>
		{#each tasks as t (t.id)}
			<li
				animate:flip={motion.flip}
				in:receive={{ key: key(t), offstage: 'none' }}
				out:send={{ key: key(t), offstage: 'none' }}
			>
				<TaskCard
					task={t}
					{size}
					selected={!!pick?.selected.includes(t.id)}
					onpick={pick?.canPick(t) ? () => pick.onpick(t) : undefined}
				/>
			</li>
		{/each}
	</ul>
{/if}
