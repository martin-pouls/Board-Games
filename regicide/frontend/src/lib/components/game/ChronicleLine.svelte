<script lang="ts">
	import { eventText } from '#lib/game/format.js';
	import type { GameEvent, PlayerView } from '#lib/types.js';

	let { events, view }: { events: GameEvent[]; view: PlayerView } = $props();

	const lines = $derived(
		events
			.map((e) => eventText(e, view))
			.filter((l): l is string => l !== null)
			.slice(-3)
	);
</script>

<div class="min-h-5 text-center text-sm text-muted-foreground italic" aria-live="polite">
	{#each lines as l, i (i)}
		<p class="truncate" title={l}>{l}</p>
	{/each}
</div>
