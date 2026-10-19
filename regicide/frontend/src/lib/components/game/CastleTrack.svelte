<!--
	The Castle deck as a track of its twelve enemies in order: four Jacks, four Queens, four Kings.
	Defeated ones are filled in, the current one is ringed in red; the count on the right includes
	the current enemy. Only shown while there is one.
-->
<script lang="ts">
	import { cn } from '#lib/utils.js';

	let { castleCount }: { castleCount: number } = $props();

	const left = $derived(castleCount + 1);
	const defeated = $derived(12 - left);
	const groups = ['J', 'Q', 'K'];
</script>

<div
	class="flex h-5 items-center gap-0.5"
	role="img"
	aria-label={`Castle: ${defeated} of 12 enemies defeated`}
>
	{#each groups as rank, g (rank)}
		{#if g > 0}<span class="w-2"></span>{/if}
		{#each [0, 1, 2, 3] as i (i)}
			{@const n = g * 4 + i}
			<span
				class={cn(
					'grid size-4.5 place-items-center rounded-[4px] border font-display text-xs leading-none font-bold',
					n < defeated
						? 'border-muted-foreground bg-muted-foreground text-background'
						: n === defeated
							? 'border-[1.5px] border-health bg-health/12 text-foreground'
							: 'text-muted-foreground'
				)}
				aria-hidden="true">{rank}</span
			>
		{/each}
	{/each}
	<span class="ml-auto text-xs text-muted-foreground tabular-nums">{left} left</span>
</div>
