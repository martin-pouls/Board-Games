<!--
	The full card of a Deep Sea task, from the bottom: the printed card (difficulty, status) beside
	whose task it is, its sentence and how it stands. Opened by tapping a task chip.
-->
<script lang="ts">
	import * as Sheet from '#lib/components/ui/sheet/index.js';
	import { seatName } from '#lib/game/format.js';
	import { taskTextOf } from '#lib/game/taskText.js';
	import { OWNER_UNASSIGNED, type Attempt, type Task } from '#lib/types.js';
	import TaskCard from './TaskCard.svelte';

	let {
		attempt,
		task = $bindable(null)
	}: {
		attempt: Attempt;
		/** The task shown; null closes the sheet. */
		task?: Task | null;
	} = $props();

	const players = $derived(attempt.members.length);
	/** The task as the latest view has it (its status may change while the sheet is open). */
	const shown = $derived(task ? (attempt.tasks.find((t) => t.id === task!.id) ?? task) : null);
	const whose = $derived.by(() => {
		if (!shown || shown.owner === OWNER_UNASSIGNED) return 'Open task';
		if (shown.owner === attempt.seat) return 'Your task';
		return `${seatName(attempt.members, shown.owner, attempt.seat)}'s task`;
	});
	const state = $derived(
		!shown
			? ''
			: shown.status === 'fulfilled'
				? 'Done.'
				: shown.status === 'failed'
					? 'Failed.'
					: attempt.result
						? 'Not done.'
						: shown.owner === OWNER_UNASSIGNED
							? 'Not taken yet.'
							: 'Open.'
	);
</script>

<Sheet.Root
	open={!!task}
	onOpenChange={(o) => {
		if (!o) task = null;
	}}
>
	<Sheet.Content
		side="bottom"
		class="mx-auto max-w-md rounded-t-2xl px-4 pt-5 pb-[max(1.25rem,env(safe-area-inset-bottom))]"
	>
		{#if shown}
			<div class="flex items-center gap-4 pr-6">
				<div class="shrink-0 [zoom:1.6] p-1.5">
					<TaskCard task={shown} {players} />
				</div>
				<div class="flex min-w-0 flex-col gap-1.5">
					<Sheet.Title class="font-display text-base font-bold">{whose}</Sheet.Title>
					<Sheet.Description class="text-base leading-snug text-foreground">
						{taskTextOf(shown, players)}.
					</Sheet.Description>
					<p class="text-sm text-muted-foreground">{state}</p>
				</div>
			</div>
		{/if}
	</Sheet.Content>
</Sheet.Root>
