<!--
	What the mission asks beyond the tasks: the special objectives and the communication rules.
	The tasks themselves lie with their owners (crew strip, own row) or in the open pool.
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
	import PlayingCard from './PlayingCard.svelte';
	import SonarPing from './SonarPing.svelte';

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
		attempt.comm === 'rapture' && attempt.tokensLeft >= 0
			? { left: attempt.tokensLeft, total: attempt.members.length - 2 }
			: null
	);
	const terrain = $derived(attempt.terrain);
	const stage = $derived(
		attempt.stages > 0
			? `Stage ${String.fromCharCode(65 + attempt.stage)} of ${attempt.stages}`
			: ''
	);

	const rules = $derived(
		[
			mission?.deadZone && { text: 'dead zone', title: 'Communicate without position' },
			!terrain && attempt.comm === 'currents' && { text: 'currents', title: modeTitle.currents },
			!terrain && attempt.comm === 'none' && { text: 'no communication', title: modeTitle.none },
			attempt.commFirstTrickOnly && {
				text: 'communication only before the first trick',
				title: 'The tasks were passed on: communicate before trick 1 or never'
			},
			mission && mission.fromTrick > 1 && { text: `comms from trick ${mission.fromTrick}` },
			stage && { text: stage, title: 'Side quest with stages: each stage is a new deal' },
			attempt.distress && { text: 'distress signal used' }
		].filter((r): r is { text: string; title?: string } => !!r)
	);
	const hint = $derived(
		mission?.faceDownTasks ? 'Win exactly as many tricks as the difficulty of your tasks' : ''
	);
</script>

{#if attempt.objectives.length || rules.length || pool || terrain || hint}
	<section class="flex flex-col gap-1 rounded-xl bg-panel px-3 py-1.5" aria-label="Mission">
		{#if terrain}
			<p class="flex items-center gap-1.5 text-xs">
				<PlayingCard card={terrain} size="xs" />
				<span>
					Unfamiliar terrain: {cardLabel(terrain)} →
					<b class="font-semibold">{MODE_TEXT[attempt.comm]}</b>
				</span>
			</p>
		{/if}
		{#if rules.length || pool}
			<ul class="flex flex-wrap items-center gap-1.5 text-xs">
				{#if pool}
					<li
						class="flex items-center gap-1.5 rounded-full bg-muted px-2 py-0.5"
						title={modeTitle.rapture}
					>
						<SonarPing class="size-3.5" />
						<span>rapture of the deep</span>
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
				{#each rules as r (r.text)}
					<li class="rounded-full bg-muted px-2 py-0.5" title={r.title}>{r.text}</li>
				{/each}
			</ul>
		{/if}
		{#if hint}
			<p class="text-xs text-muted-foreground">{hint}</p>
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
