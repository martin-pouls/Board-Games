<script lang="ts">
	import { cardLabel } from '#lib/game/cards.js';
	import { positionText } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import type { Communication } from '#lib/types.js';

	// The marker of a communicated card: ▲ highest, ● only, ▼ lowest, ? in a dead zone.
	let { comm, class: className = '' }: { comm: Communication; class?: string } = $props();

	const MARK = { highest: '▲', only: '●', lowest: '▼', '': '?' } as const;
	const title = $derived(`${cardLabel(comm.card)}: ${positionText(comm.position)}`);
</script>

<span
	class={cn(
		'grid size-4.5 place-items-center rounded-full bg-signal text-[0.6rem] leading-none font-bold text-primary-foreground shadow',
		className
	)}
	{title}
	aria-label={title}>{MARK[comm.position]}</span
>
