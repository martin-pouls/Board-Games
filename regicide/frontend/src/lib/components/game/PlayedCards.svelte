<script lang="ts">
	import { playStacks } from '#lib/game/format.js';
	import type { PlayedSet, PlayerSummary } from '#lib/types.js';
	import PlayerInitial from './PlayerInitial.svelte';
	import PlayingCard from './PlayingCard.svelte';

	let { plays, players }: { plays: PlayedSet[]; players: PlayerSummary[] } = $props();

	const stacks = $derived(playStacks(plays));
</script>

{#if stacks.length > 0}
	<ul class="flex flex-wrap gap-x-3 gap-y-1.5" aria-label="Cards played against the enemy">
		{#each stacks as s (s.playerId)}
			<li class="flex items-center gap-1">
				<PlayerInitial id={s.playerId} {players} />
				<span class="flex -space-x-4">
					{#each s.cards as c, j (j)}
						<PlayingCard card={c} size="xs" />
					{/each}
				</span>
			</li>
		{/each}
	</ul>
{:else}
	<p class="text-xs text-muted-foreground italic">No cards played yet.</p>
{/if}
