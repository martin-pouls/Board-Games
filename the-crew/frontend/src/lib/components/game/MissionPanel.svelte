<!--
	What the mission asks beyond the tasks: the special objectives and the communication rules.
	The tasks themselves lie with their owners (crew strip, own row) or in the open pool.
-->
<script lang="ts">
	import Check from '@lucide/svelte/icons/check';
	import Circle from '@lucide/svelte/icons/circle';
	import X from '@lucide/svelte/icons/x';
	import { objectiveText, ownerText } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import { OWNER_UNASSIGNED, type Attempt, type MissionDef } from '#lib/types.js';

	let {
		attempt,
		mission,
		roles = []
	}: {
		attempt: Attempt;
		mission: MissionDef | null;
		/** Mission 50: the seats the commander tapped for the roles (null = open), not yet confirmed. */
		roles?: (number | null)[];
	} = $props();

	/** The objective's owner, or the seat tapped for its role while the commander picks. */
	function owner(i: number): number {
		const rule = mission?.objectives[i]?.ownerRule;
		const pending = rule === 'roleA' ? roles[0] : rule === 'roleB' ? roles[1] : undefined;
		return attempt.objectives[i].owner === OWNER_UNASSIGNED && pending != null
			? pending
			: attempt.objectives[i].owner;
	}

	const rules = $derived(
		[
			mission?.deadZone && { text: 'dead zone', title: 'Communicate without position' },
			mission && mission.fromTrick > 1 && { text: `comms from trick ${mission.fromTrick}` },
			attempt.distress && { text: 'distress signal used' }
		].filter((r): r is { text: string; title?: string } => !!r)
	);
</script>

{#if attempt.objectives.length || rules.length}
	<section class="flex flex-col gap-1 rounded-xl bg-panel px-3 py-1.5" aria-label="Mission">
		{#if rules.length}
			<ul class="flex flex-wrap gap-1.5 text-xs">
				{#each rules as r (r.text)}
					<li class="rounded-full bg-muted px-2 py-0.5" title={r.title}>{r.text}</li>
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
							<b class="font-semibold">{ownerText(owner(i), attempt.members, attempt.seat)}:</b>
							{objectiveText(o)}
						</span>
					</li>
				{/each}
			</ul>
		{/if}
	</section>
{/if}
