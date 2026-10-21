<!--
	The open tasks in the middle of the table, during the setup (token swap, poll, draft): in token
	order, so a swap or move reorders them, in a grid of five (Deep Sea's wider task cards: up to
	four a row on a phone). Past
	ten tasks the grid scrolls inside the panel. A task taken flies to its owner.
-->
<script lang="ts">
	import { flip } from 'svelte/animate';
	import { getCardMotion } from '@board-games/shared/cardMotion.js';
	import { byTokenOrder } from '#lib/game/cards.js';
	import { cn } from '#lib/utils.js';
	import { OWNER_UNASSIGNED, type Attempt, type Task } from '#lib/types.js';
	import TaskCard from './TaskCard.svelte';
	import type { TaskPick } from './TaskList.svelte';

	let { attempt, pick }: { attempt: Attempt; pick: TaskPick } = $props();

	const motion = getCardMotion();
	const { send, receive } = motion;
	const key = (t: Task) => `task-${attempt.number}-${t.id}`;

	const open = $derived(byTokenOrder(attempt.tasks.filter((t) => t.owner === OWNER_UNASSIGNED)));
	/** Deep Sea task cards (text faces) are wider than a playing card. */
	const cards = $derived(attempt.tasks.some((t) => !t.card && (!!t.objective || t.difficulty > 0)));
	/** The height of the rows that hold ten tasks; more tasks scroll. */
	const tenRows = $derived(cards ? 3 * 56 + 2 * 10 + 18 : 2 * 44 + 14 + 16);
</script>

<section
	class="flex flex-col rounded-xl bg-panel px-1 pt-1 pb-0.5 [--tab-ring:var(--panel)]"
	aria-label="Open tasks"
>
	<div class="flex items-baseline justify-between px-1 text-xs text-muted-foreground">
		<span class="font-display font-bold tracking-wide uppercase">Open tasks</span>
		<span>{open.length} of {attempt.tasks.length}</span>
	</div>
	{#if open.length}
		<ul
			class={cn(
				'overflow-y-auto px-1',
				cards
					? 'flex flex-wrap justify-center gap-x-3 gap-y-2.5 pt-2.5 pb-2'
					: 'grid grid-cols-5 justify-items-center gap-x-1.5 gap-y-3.5 pt-1.5 pb-2.5'
			)}
			style:max-height={open.length > 10 ? `${tenRows}px` : undefined}
			aria-label="Open tasks"
		>
			{#each open as t (t.id)}
				<li
					animate:flip={motion.flip}
					in:receive={{ key: key(t), offstage: 'none' }}
					out:send={{ key: key(t), offstage: 'none' }}
				>
					<TaskCard
						task={t}
						size="xs"
						players={attempt.members.length}
						selected={pick.selected.includes(t.id)}
						onpick={pick.canPick(t) ? () => pick.onpick(t) : undefined}
					/>
				</li>
			{/each}
		</ul>
	{/if}
</section>
