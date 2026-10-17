<script lang="ts" module>
	import { cap, terms } from '#lib/edition.js';
	import type { MissionDef } from '#lib/types.js';

	export type MarkerKind = 'deadZone' | 'disruption' | 'objective' | 'commander';

	export const MARKERS: Record<MarkerKind, { label: string }> = {
		deadZone: { label: 'Dead zone' },
		disruption: { label: 'Disruption' },
		objective: { label: 'Special objective' },
		commander: { label: `${cap(terms.commander)} decides` }
	};

	/** The special rules of a mission worth a marker in the grid. */
	export function missionMarkers(def: MissionDef | undefined): MarkerKind[] {
		if (!def) return [];
		const out: MarkerKind[] = [];
		if (def.deadZone) out.push('deadZone');
		if (def.fromTrick > 1) out.push('disruption');
		if (def.objectives.length > 0) out.push('objective');
		if (def.distribution === 'commanderDecision' || def.distribution === 'commanderDistribution') {
			out.push('commander');
		}
		return out;
	}
</script>

<script lang="ts">
	import Check from '@lucide/svelte/icons/check';
	import Clock from '@lucide/svelte/icons/clock';
	import EyeOff from '@lucide/svelte/icons/eye-off';
	import Target from '@lucide/svelte/icons/target';
	import UserStar from '@lucide/svelte/icons/user-star';
	import { missionLabelOf } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import type { LogEntry } from '#lib/types.js';

	const MISSIONS = Array.from({ length: 50 }, (_, i) => i + 1);

	let {
		selected,
		logbook,
		missions = null,
		editable = false,
		disabled = false,
		onselect
	}: {
		/** The selected mission, 1–50. */
		selected: number;
		logbook: LogEntry[];
		/** The mission table, for the markers; the grid works without it. */
		missions?: MissionDef[] | null;
		/** Only the host can pick. */
		editable?: boolean;
		disabled?: boolean;
		onselect?: (mission: number) => void;
	} = $props();

	const entries = $derived(new Map(logbook.map((e) => [e.mission, e])));
	const defs = $derived(new Map((missions ?? []).map((m) => [m.number, m])));

	/** What the logbook says about a mission, for the title and the screen reader. */
	function stateText(e: LogEntry | undefined): string {
		if (!e) return 'not played yet';
		const tries = `${e.attempts} ${e.attempts === 1 ? 'attempt' : 'attempts'}`;
		if (!e.completed) return `${tries}, not completed`;
		return `completed, logbook total ${e.total}${e.distress ? ' (distress signal used)' : ''}`;
	}

	function label(n: number): string {
		const marks = missionMarkers(defs.get(n)).map((k) => MARKERS[k].label.toLowerCase());
		return `Mission ${missionLabelOf(n, missions)}: ${stateText(entries.get(n))}${marks.length ? `; ${marks.join(', ')}` : ''}`;
	}

	const cellClass =
		'relative flex h-14 w-full flex-col items-center justify-center rounded-lg border text-center leading-none';
</script>

{#snippet marker(kind: MarkerKind)}
	{#if kind === 'deadZone'}
		<EyeOff class="size-3" />
	{:else if kind === 'disruption'}
		<Clock class="size-3" />
	{:else if kind === 'objective'}
		<Target class="size-3" />
	{:else}
		<UserStar class="size-3" />
	{/if}
{/snippet}

{#snippet cellBody(n: number)}
	{@const e = entries.get(n)}
	<span class="absolute top-1 left-1 flex gap-px text-muted-foreground" aria-hidden="true">
		{#each missionMarkers(defs.get(n)) as kind (kind)}
			{@render marker(kind)}
		{/each}
	</span>
	<span class="font-display text-base font-bold">{missionLabelOf(n, missions)}</span>
	<span
		class={cn(
			'mt-1 flex h-3 items-center gap-px text-[0.68rem] font-semibold',
			e?.completed ? 'text-success' : 'text-muted-foreground'
		)}
		aria-hidden="true"
	>
		{#if e?.completed}
			<Check class="size-3" strokeWidth={3} />{e.total}
		{:else if e}
			{e.attempts}×
		{/if}
	</span>
{/snippet}

<ul class="grid grid-cols-7 gap-1" aria-label="Missions">
	{#each MISSIONS as n (n)}
		{@const e = entries.get(n)}
		{@const isSelected = n === selected}
		{@const look = cn(
			cellClass,
			isSelected
				? 'border-signal bg-signal/20 text-foreground ring-2 ring-signal'
				: e?.completed
					? 'border-success/40 bg-success/10'
					: 'border-border bg-card'
		)}
		<li class="contents">
			{#if editable}
				<button
					type="button"
					class={cn(look, 'transition-colors hover:border-signal disabled:opacity-60')}
					{disabled}
					aria-pressed={isSelected}
					aria-label={label(n)}
					title={label(n)}
					onclick={() => onselect?.(n)}
				>
					{@render cellBody(n)}
				</button>
			{:else}
				<span
					class={look}
					role="img"
					aria-label={`${label(n)}${isSelected ? ' (selected)' : ''}`}
					title={label(n)}
				>
					{@render cellBody(n)}
				</span>
			{/if}
		</li>
	{/each}
</ul>

<ul class="mt-2 flex flex-wrap gap-x-3 gap-y-1 text-xs text-muted-foreground">
	<li class="flex items-center gap-1">
		<Check class="size-3 text-success" strokeWidth={3} />Completed (logbook total)
	</li>
	<li class="flex items-center gap-1"><span class="font-semibold">2×</span>Attempts so far</li>
	{#each Object.entries(MARKERS) as [kind, m] (kind)}
		<li class="flex items-center gap-1">
			{@render marker(kind as MarkerKind)}{m.label}
		</li>
	{/each}
</ul>
