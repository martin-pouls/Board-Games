<!-- The Tavern deck and the discard pile (the Castle deck is the track in the enemy panel). -->
<script lang="ts">
	import Beer from '@lucide/svelte/icons/beer';
	import Layers from '@lucide/svelte/icons/layers';
	import { cn } from '#lib/utils.js';
	import type { PlayerView } from '#lib/types.js';

	let { view, class: className }: { view: PlayerView; class?: string } = $props();
</script>

<ul class={cn('flex items-center gap-3 text-sm text-muted-foreground tabular-nums', className)}>
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
</ul>
