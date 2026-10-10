<script lang="ts">
	import { suitGlyph } from '#lib/game/format.js';
	import type { PlayedSet, PlayerSummary } from '#lib/types.js';
	import PlayingCard from './PlayingCard.svelte';

	let { plays, players }: { plays: PlayedSet[]; players: PlayerSummary[] } = $props();

	const initial = (id: string) => (players.find((p) => p.id === id)?.name ?? '?').slice(0, 1);
	const nameOf = (id: string) => players.find((p) => p.id === id)?.name ?? 'Someone';
</script>

{#if plays.length > 0}
	<ul class="flex flex-wrap gap-x-3 gap-y-1.5" aria-label="Cards played against the enemy">
		{#each plays as p, i (i)}
			<li class="flex items-center gap-1" title={`${nameOf(p.playerId)}: ${p.value}`}>
				<span
					class="flex size-5 shrink-0 items-center justify-center rounded-full bg-secondary text-xs font-semibold uppercase"
					aria-label={nameOf(p.playerId)}>{initial(p.playerId)}</span
				>
				<span class="flex -space-x-2">
					{#each p.cards as c, j (j)}
						<PlayingCard card={c} size="xs" />
					{/each}
				</span>
				<span class="text-xs text-muted-foreground tabular-nums">{p.value}</span>
				{#if p.doubled}
					<span class="rounded bg-primary px-1 text-xs font-bold text-primary-foreground">×2</span>
				{/if}
				{#each p.blockedSuits as s (s)}
					<span class="text-xs opacity-40" style="font-family: system-ui, sans-serif" title="Immune"
						>{suitGlyph(s)}&#xFE0E;</span
					>
				{/each}
			</li>
		{/each}
	</ul>
{:else}
	<p class="text-xs text-muted-foreground italic">No cards played yet.</p>
{/if}
