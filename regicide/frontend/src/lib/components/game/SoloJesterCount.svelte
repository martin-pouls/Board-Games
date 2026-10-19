<!-- Solo only: the Jester powers, one cap each, faded once used. -->
<script lang="ts">
	import { cn } from '#lib/utils.js';
	import type { PlayerView } from '#lib/types.js';
	import JesterHat from './JesterHat.svelte';

	let { view, class: className }: { view: PlayerView; class?: string } = $props();
</script>

{#if view.rules.soloJesters > 0}
	<span
		class={cn('flex items-center gap-0.5', className)}
		role="img"
		title={`${view.soloJestersLeft} Jester powers left`}
		aria-label={`${view.soloJestersLeft} Jester powers left`}
	>
		{#each Array.from({ length: view.rules.soloJesters }, (_x, i) => i) as i (i)}
			<JesterHat
				class={cn('size-5', i < view.soloJestersLeft ? 'text-jester' : 'text-muted-foreground/30')}
			/>
		{/each}
	</span>
{/if}
