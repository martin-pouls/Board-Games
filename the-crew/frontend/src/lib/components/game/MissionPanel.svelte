<script lang="ts">
	import Check from '@lucide/svelte/icons/check';
	import Circle from '@lucide/svelte/icons/circle';
	import X from '@lucide/svelte/icons/x';
	import { objectiveText, ownerText, seatName } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import { OWNER_UNASSIGNED, type Attempt, type Task } from '#lib/types.js';
	import TaskCard from './TaskCard.svelte';

	let {
		attempt,
		canPick = () => false,
		selected = [],
		onpick
	}: {
		attempt: Attempt;
		/** Whether a task can be tapped now (draft, token setup). */
		canPick?: (t: Task) => boolean;
		selected?: number[];
		onpick?: (t: Task) => void;
	} = $props();

	const done = $derived(attempt.tasks.filter((t) => t.status === 'fulfilled').length);
	const owner = (t: Task) =>
		t.owner === OWNER_UNASSIGNED ? '' : seatName(attempt.members, t.owner, attempt.seat);
</script>

{#if attempt.tasks.length || attempt.objectives.length}
	<section class="flex flex-col gap-2 rounded-xl bg-panel px-3 py-2" aria-label="Mission">
		{#if attempt.tasks.length}
			<div class="flex items-baseline justify-between text-xs text-muted-foreground">
				<span class="font-display font-semibold tracking-wide uppercase">Tasks</span>
				<span>{done}/{attempt.tasks.length} done</span>
			</div>
			<ul class="flex flex-wrap justify-center gap-x-2.5 gap-y-1.5 pt-1">
				{#each attempt.tasks as t (t.id)}
					<li>
						<TaskCard
							task={t}
							owner={owner(t)}
							selected={selected.includes(t.id)}
							onpick={canPick(t) ? () => onpick?.(t) : undefined}
						/>
					</li>
				{/each}
			</ul>
		{/if}
		{#if attempt.objectives.length}
			<ul class="flex flex-col gap-1 text-sm leading-snug">
				{#each attempt.objectives as o, i (i)}
					<li class="flex items-start gap-1.5">
						{#if o.status === 'fulfilled'}
							<Check class="mt-0.5 size-4 shrink-0 text-success" strokeWidth={3} />
						{:else if o.status === 'failed'}
							<X class="mt-0.5 size-4 shrink-0 text-danger" strokeWidth={3} />
						{:else}
							<Circle class="mt-0.5 size-4 shrink-0 text-muted-foreground" />
						{/if}
						<span class={cn(o.status === 'failed' && 'text-muted-foreground line-through')}>
							<b class="font-semibold">{ownerText(o.owner, attempt.members, attempt.seat)}:</b>
							{objectiveText(o)}
						</span>
					</li>
				{/each}
			</ul>
		{/if}
	</section>
{/if}
