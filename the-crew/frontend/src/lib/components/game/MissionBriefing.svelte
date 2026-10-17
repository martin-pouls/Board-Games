<script lang="ts">
	import Check from '@lucide/svelte/icons/check';
	import { Skeleton } from '#lib/components/ui/skeleton/index.js';
	import { briefing, missionLabel } from '#lib/game/format.js';
	import type { LogEntry, MissionDef } from '#lib/types.js';

	let {
		mission,
		def,
		entry,
		loading = false
	}: {
		mission: number;
		/** The mission's table row; null while it is not known. */
		def: MissionDef | null;
		/** The mission's logbook line, if it was played. */
		entry?: LogEntry;
		/** The catalogue is on its way. */
		loading?: boolean;
	} = $props();

	const lines = $derived(def ? briefing(def) : []);
	const name = $derived(def ? missionLabel(def) : String(mission));
</script>

<section aria-label={`Mission ${name} briefing`}>
	<h3 class="flex flex-wrap items-center gap-x-2 text-base font-semibold">
		Mission {name}
		{#if entry?.completed}
			<span class="flex items-center gap-0.5 text-sm font-medium text-success">
				<Check class="size-4" strokeWidth={3} />Completed
			</span>
		{:else if entry}
			<span class="text-sm font-medium text-muted-foreground">
				{entry.attempts}
				{entry.attempts === 1 ? 'attempt' : 'attempts'} so far
			</span>
		{/if}
	</h3>
	{#if lines.length}
		<ul class="mt-1 flex list-disc flex-col gap-1 pl-5 text-sm marker:text-muted-foreground">
			{#each lines as line, i (i)}
				<li>{line}</li>
			{/each}
		</ul>
	{:else if def}
		<p class="mt-1 text-sm text-muted-foreground">No special rules.</p>
	{:else if loading}
		<div class="mt-2 flex flex-col gap-1.5" aria-busy="true">
			<Skeleton class="h-4 w-3/4" />
			<Skeleton class="h-4 w-1/2" />
		</div>
	{:else}
		<p class="mt-1 text-sm text-muted-foreground">The briefing could not be loaded.</p>
	{/if}
	{#if entry}
		<p class="mt-2 text-xs text-muted-foreground">
			Logbook: {entry.attempts}
			{entry.attempts === 1 ? 'attempt' : 'attempts'}{entry.distress
				? ', distress signal used (+1)'
				: ''}, total {entry.total}.
		</p>
	{/if}
</section>
