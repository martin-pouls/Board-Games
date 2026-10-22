<!--
	The open tasks in the middle of the table, during the setup (token swap, poll, draft): in token
	order, so a swap or move reorders them, in a grid of five; past ten tasks the grid scrolls inside
	the panel. Deep Sea's task cards are chips with their difficulty in two columns (past four rows
	they scroll, on a short screen past three), and the line below reads out the tapped one's card. A task taken flies to its
	owner.
-->
<script lang="ts">
	import { flip } from 'svelte/animate';
	import RectangleVertical from '@lucide/svelte/icons/rectangle-vertical';
	import { getCardMotion } from '@board-games/shared/cardMotion.js';
	import { byTokenOrder } from '#lib/game/cards.js';
	import { isTaskCard } from '#lib/game/taskChip.js';
	import { getTaskSheet } from '#lib/game/taskSheet.js';
	import { taskTextOf } from '#lib/game/taskText.js';
	import { cn } from '#lib/utils.js';
	import { OWNER_UNASSIGNED, type Attempt, type Task } from '#lib/types.js';
	import TaskCard from './TaskCard.svelte';
	import TaskChip from './TaskChip.svelte';
	import type { TaskPick } from './TaskList.svelte';

	let { attempt, pick }: { attempt: Attempt; pick: TaskPick } = $props();

	const motion = getCardMotion();
	const { send, receive } = motion;
	const key = (t: Task) => `task-${attempt.number}-${t.id}`;

	const open = $derived(byTokenOrder(attempt.tasks.filter((t) => t.owner === OWNER_UNASSIGNED)));
	const players = $derived(attempt.members.length);
	/** Deep Sea: task cards, shown as chips. */
	const chips = $derived(attempt.tasks.some(isTaskCard));
	/** The height of the rows that hold ten tasks; more tasks scroll. */
	const tenRows = 2 * 44 + 14 + 16;
	/** Deep Sea: the rows of chips (30 px, 2 px apart); past four (three) the next one peeks out. */
	const rows = $derived(Math.ceil(open.length / 2));

	/**
	 * Deep Sea: in the draft and the assignment a tap picks, and the line below the chips reads out
	 * the picked one; its Card button opens the full card. A tap that picks nothing opens it directly.
	 */
	const picking = $derived(!!pick.read);
	const read = $derived(open.find((t) => t.id === pick.selected[0] && t.objective) ?? null);
	const sheet = getTaskSheet();
	function tap(t: Task) {
		if (pick.canPick(t)) pick.onpick(t);
		else sheet?.show(t);
	}
</script>

<section
	class="flex flex-col rounded-xl bg-panel px-1 pt-1 pb-0.5 [--tab-ring:var(--panel)]"
	aria-label="Open tasks"
>
	<div class="flex items-baseline justify-between px-1 text-xs text-muted-foreground">
		<span class="font-display font-bold tracking-wide uppercase">Open tasks</span>
		<span>{open.length} of {attempt.tasks.length}</span>
	</div>
	{#if open.length && chips}
		<ul
			class={cn(
				'grid max-h-[8.875rem] grid-cols-2 gap-x-1.5 gap-y-0.5 overflow-y-auto px-0.5 py-1 short:max-h-[6.8125rem]',
				rows > 4 && 'fade-end',
				rows > 3 && 'short:fade-end'
			)}
			aria-label="Open tasks"
		>
			{#each open as t (t.id)}
				{@const sel = pick.selected.includes(t.id)}
				<li
					class="min-w-0"
					animate:flip={motion.flip}
					in:receive={{ key: key(t), offstage: 'none' }}
					out:send={{ key: key(t), offstage: 'none' }}
				>
					<button
						type="button"
						class={cn(
							'flex min-h-7.5 w-full items-center gap-2 rounded-[0.5625rem] px-1.5 py-px text-left focus-visible:ring-3 focus-visible:ring-ring focus-visible:outline-none',
							sel && 'bg-card shadow-[inset_0_0_0_1.5px_var(--primary)]'
						)}
						aria-pressed={picking ? sel : undefined}
						aria-label={`${t.objective ? taskTextOf(t, players) : 'Face-down task'}, difficulty ${t.difficulty}`}
						onclick={() => tap(t)}
					>
						<span
							class="grid size-5 shrink-0 place-items-center rounded-full bg-foreground font-display text-[0.6875rem] font-extrabold text-background"
							aria-hidden="true">{t.difficulty}</span
						>
						<TaskChip task={t} {players} still class="min-w-0 shrink" />
					</button>
				</li>
			{/each}
		</ul>
		{#if picking}
			<div
				class="mx-0.5 mt-0.5 mb-0.5 flex min-h-8 items-center gap-1 rounded-lg bg-background py-0.5 pr-0.5 pl-2 shadow-[inset_3px_0_0_var(--primary)]"
			>
				<p
					class={cn(
						'min-w-0 flex-1 py-0.5 text-[0.8125rem] leading-tight',
						!read && 'text-muted-foreground'
					)}
					aria-live="polite"
				>
					{read ? `${taskTextOf(read, players)}.` : 'Tap a task to read its card.'}
				</p>
				{#if read}
					<!-- The printed card of the picked task, in the sheet. -->
					<button
						type="button"
						class="relative flex h-7 shrink-0 items-center gap-1 rounded-md border border-border bg-card px-1.5 font-display text-[0.6875rem] font-bold tracking-wide text-muted-foreground uppercase before:absolute before:-inset-1.5 focus-visible:ring-3 focus-visible:ring-ring focus-visible:outline-none"
						aria-label="Show the full card"
						onclick={() => sheet?.show(read)}
					>
						<RectangleVertical class="size-3.5" aria-hidden="true" />Card
					</button>
				{/if}
			</div>
		{/if}
	{:else if open.length}
		<ul
			class="grid grid-cols-5 justify-items-center gap-x-1.5 gap-y-3.5 overflow-y-auto px-1 pt-1.5 pb-2.5"
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
						selected={pick.selected.includes(t.id)}
						onpick={pick.canPick(t) ? () => pick.onpick(t) : undefined}
					/>
				</li>
			{/each}
		</ul>
	{/if}
</section>
