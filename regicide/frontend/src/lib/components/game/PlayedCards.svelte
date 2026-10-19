<!--
	The cards played against the current enemy, one overlapping stack per player, on one line. The
	player's initial sits beside the stack; when that doesn't fit (three or four players), it moves
	onto the stack's blank bottom left corner. The more cards there are, the closer they overlap; the
	corner index always stays visible.
-->
<script lang="ts">
	import { playStacks } from '#lib/game/format.js';
	import type { PlayedSet, PlayerSummary } from '#lib/types.js';
	import PlayerInitial from './PlayerInitial.svelte';
	import PlayingCard from './PlayingCard.svelte';

	let { plays, players }: { plays: PlayedSet[]; players: PlayerSummary[] } = $props();

	const CARD = 34;
	const BESIDE = 22; // the initial and its gap
	const GAP = 10;
	const MAX_STEP = 18;
	const MIN_STEP = 9;

	const PAD = 8; // the list's side padding, so the cards' rings and shadows aren't clipped

	let width = $state(240 + PAD);
	const stacks = $derived(playStacks(plays));
	const extra = $derived(stacks.reduce((n, s) => n + s.cards.length - 1, 0));
	const fixed = (initial: number) => stacks.length * (initial + CARD) + (stacks.length - 1) * GAP;
	const inner = $derived(width - PAD);
	const beside = $derived(fixed(BESIDE) + extra * MAX_STEP <= inner);
	/** How much of each card but the last in a stack shows. */
	const step = $derived(
		extra === 0
			? MAX_STEP
			: Math.max(
					MIN_STEP,
					Math.min(MAX_STEP, Math.floor((inner - fixed(beside ? BESIDE : 0)) / extra))
				)
	);
</script>

{#if stacks.length > 0}
	<ul
		class="-mx-1 flex min-h-12 items-center gap-2.5 overflow-x-clip px-1"
		aria-label="Cards played against the enemy"
		bind:clientWidth={width}
	>
		{#each stacks as s (s.playerId)}
			<li class="relative flex shrink-0 items-center gap-0.5">
				<PlayerInitial
					id={s.playerId}
					{players}
					class={beside ? undefined : 'absolute bottom-0.5 left-0.5 z-10 ring-2 ring-stock'}
				/>
				<span class="flex">
					{#each s.cards as c, j (j)}
						<span class="shrink-0" style:margin-left={j > 0 ? `${step - CARD}px` : undefined}>
							<PlayingCard card={c} size="xs" />
						</span>
					{/each}
				</span>
			</li>
		{/each}
	</ul>
{:else}
	<p class="flex min-h-12 items-center text-sm text-muted-foreground">No cards played yet.</p>
{/if}
