<script lang="ts">
	import type { PlayerSummary } from '#lib/types.js';

	let { id, players }: { id: string; players: PlayerSummary[] } = $props();

	const name = $derived(players.find((p) => p.id === id)?.name ?? 'Someone');
</script>

<span
	class="flex size-5 shrink-0 items-center justify-center rounded-full bg-secondary text-xs font-semibold uppercase"
	title={name}
	aria-label={name}>{name.slice(0, 1)}</span
>
