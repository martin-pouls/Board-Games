<!--
	What the mission asks beyond the tasks: the special objectives and the communication rules
	(only while the attempt runs: the debrief says what counted). The tasks themselves lie with
	their owners (crew strip, own row) or in the open pool.
-->
<script lang="ts">
	import Check from '@lucide/svelte/icons/check';
	import Circle from '@lucide/svelte/icons/circle';
	import X from '@lucide/svelte/icons/x';
	import { terms } from '#lib/edition.js';
	import { cardLabel } from '#lib/game/cards.js';
	import { objectiveText, ownerText } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import { OWNER_UNASSIGNED, type Attempt, type MissionDef } from '#lib/types.js';
	import EditionGlyph from './EditionGlyph.svelte';
	import TaskMini from './TaskMini.svelte';

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

	/** The debrief shows the objectives only. */
	const over = $derived(!!attempt.result);

	// Deep Sea: the communication mode of this attempt (the terrain card has already decided).
	const MODE_TEXT = {
		'': 'normal communication',
		currents: 'currents',
		rapture: 'rapture of the deep',
		none: 'no communication'
	} as const;
	const modeTitle = {
		currents: 'Currents: communicate a card without its position',
		rapture: `Rapture of the deep: the crew shares ${terms.commToken}s, and communicates again while any are left`,
		none: 'Nobody may communicate'
	} as const;
	/** The shared sonar token pool of rapture of the deep: used and left. */
	const pool = $derived(
		!over && attempt.comm === 'rapture' && attempt.tokensLeft >= 0
			? { left: attempt.tokensLeft, total: attempt.members.length - 2 }
			: null
	);
	const terrain = $derived(over ? null : attempt.terrain);

	/** The communication rules (and the distress signal), each a label and what it means. */
	const rules = $derived(
		over
			? []
			: [
					mission?.deadZone && {
						label: 'Dead zone',
						text: 'communicate without the position',
						title: 'Nobody says whether a communicated card is the highest, only or lowest'
					},
					!terrain &&
						attempt.comm === 'currents' && {
							label: 'Currents',
							text: 'communicate without the position',
							title: modeTitle.currents
						},
					!terrain &&
						attempt.comm === 'none' && {
							label: 'No communication',
							text: '',
							title: modeTitle.none
						},
					attempt.commFirstTrickOnly && {
						label: 'Passed on',
						text: 'communicate only before the first trick',
						title: 'The tasks were passed on: communicate before trick 1 or never'
					},
					mission &&
						mission.fromTrick > 1 && {
							label: 'Disruption',
							text: `nobody communicates before trick ${mission.fromTrick}`
						},
					attempt.distress && { label: 'Distress signal', text: 'used' }
				].filter((r): r is { label: string; text: string; title?: string } => !!r)
	);
</script>

{#if attempt.objectives.length || rules.length || pool || terrain}
	<section
		class="flex flex-col gap-1 rounded-[0.625rem] bg-panel px-2 py-1 text-xs"
		aria-label="Mission"
	>
		{#if terrain}
			<p class="flex items-center gap-1.5">
				<b class="font-bold">Unfamiliar terrain:</b>
				<TaskMini card={terrain} size="chip" />
				<span class="sr-only">{cardLabel(terrain)}</span>
				<span>→ <b class="font-semibold">{MODE_TEXT[attempt.comm]}</b></span>
			</p>
		{/if}
		{#if rules.length || pool}
			<ul class="flex flex-wrap items-center gap-x-3 gap-y-0.5">
				{#if pool}
					<li class="flex items-center gap-1.5" title={modeTitle.rapture}>
						<EditionGlyph kind="comm" class="size-3.5" />
						<b class="font-bold">Rapture of the deep</b>
						<span
							class="flex gap-0.5"
							role="img"
							aria-label={`${pool.left} of ${pool.total} ${terms.commToken}s left`}
						>
							{#each Array.from({ length: pool.total }, (_, i) => i) as i (i)}
								<span
									class={cn(
										'size-2 rounded-full border border-signal',
										i < pool.left && 'bg-signal'
									)}
								></span>
							{/each}
						</span>
					</li>
				{/if}
				{#each rules as r (r.label)}
					<li title={r.title}>
						<b class="font-bold">{r.label}{r.text ? ':' : ''}</b>
						{r.text}
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
							<b class="font-semibold">{ownerText(owner(i), attempt.members, attempt.seat)}:</b>
							{objectiveText(o)}
						</span>
					</li>
				{/each}
			</ul>
		{/if}
	</section>
{/if}
