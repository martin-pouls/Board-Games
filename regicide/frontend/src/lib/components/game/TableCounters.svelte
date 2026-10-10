<script lang="ts">
	import Beer from '@lucide/svelte/icons/beer';
	import Castle from '@lucide/svelte/icons/castle';
	import Drama from '@lucide/svelte/icons/drama';
	import Layers from '@lucide/svelte/icons/layers';
	import { cn } from '#lib/utils.js';
	import type { PlayerView } from '#lib/types.js';

	let { view }: { view: PlayerView } = $props();

	const enemiesLeft = $derived(view.castleCount + (view.enemy ? 1 : 0));
	const solo = $derived(view.rules.soloJesters > 0);
</script>

<ul class="flex items-center gap-3 text-sm text-muted-foreground tabular-nums">
	<li class="flex items-center gap-1" title={`Castle: ${enemiesLeft} enemies left`}>
		<Castle class="size-4" aria-hidden="true" />
		<span class="sr-only">Castle</span>
		{view.castleCount}
		<span class="text-xs">({enemiesLeft} left)</span>
	</li>
	<li class="flex items-center gap-1" title="Tavern deck">
		<Beer class="size-4" aria-hidden="true" />
		<span class="sr-only">Tavern</span>
		{view.tavernCount}
	</li>
	<li class="flex items-center gap-1" title="Discard pile">
		<Layers class="size-4" aria-hidden="true" />
		<span class="sr-only">Discard</span>
		{view.discardCount}
	</li>
	{#if solo}
		<li
			class="flex items-center gap-0.5"
			title={`${view.soloJestersLeft} Jester powers left`}
			aria-label={`${view.soloJestersLeft} Jester powers left`}
		>
			{#each Array.from({ length: view.rules.soloJesters }, (_x, i) => i) as i (i)}
				<Drama
					class={cn(
						'size-4',
						i < view.soloJestersLeft ? 'text-jester' : 'text-muted-foreground/30'
					)}
				/>
			{/each}
		</li>
	{/if}
</ul>
