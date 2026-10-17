<!--
	The open tasks in the middle of the table, during the setup (token swap, poll, draft): in token
	order, so a swap or move reorders them. A task taken flies to its owner.
-->
<script lang="ts">
	import { byTokenOrder } from '#lib/game/cards.js';
	import { OWNER_UNASSIGNED, type Attempt } from '#lib/types.js';
	import TaskList, { type TaskPick } from './TaskList.svelte';

	let { attempt, pick }: { attempt: Attempt; pick: TaskPick } = $props();

	const open = $derived(byTokenOrder(attempt.tasks.filter((t) => t.owner === OWNER_UNASSIGNED)));
</script>

<section class="flex flex-col gap-2 rounded-xl bg-panel px-3 py-2" aria-label="Open tasks">
	<div class="flex items-baseline justify-between text-xs text-muted-foreground">
		<span class="font-display font-semibold tracking-wide uppercase">Open tasks</span>
		<span>{open.length} of {attempt.tasks.length}</span>
	</div>
	<TaskList
		tasks={open}
		attempt={attempt.number}
		size="task"
		players={attempt.members.length}
		{pick}
		label="Open tasks"
		class="justify-center pt-1"
	/>
</section>
