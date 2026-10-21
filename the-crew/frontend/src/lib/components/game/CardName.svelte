<!-- A card named in a button: its value and symbol ("Play 3⟲"); the label says it in words. -->
<script lang="ts">
	import type { Card } from '#lib/types.js';
	import SuitGlyph from './SuitGlyph.svelte';

	let { card }: { card: Card } = $props();
</script>

<span class="inline-flex items-center gap-0.5 font-display font-bold">
	{card.rank}<SuitGlyph suit={card.suit} class="size-3.5" />
</span>
