<script lang="ts">
	import { cn } from '#lib/utils.js';
	import type { PlayerSummary } from '#lib/types.js';

	let {
		id,
		players,
		class: className
	}: { id: string; players: PlayerSummary[]; class?: string } = $props();

	const name = $derived(players.find((p) => p.id === id)?.name ?? 'Someone');
</script>

<!-- In the card's mist colours in both modes, so it looks the same beside a stack and on a card. -->
<span
	class={cn(
		'flex size-5 shrink-0 items-center justify-center rounded-full bg-ridge-1 text-xs font-semibold text-suit-black uppercase',
		className
	)}
	title={name}
	aria-label={name}>{name.slice(0, 1)}</span
>
