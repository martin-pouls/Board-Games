<script lang="ts">
	import { tokenLabel, tokenMeaning } from '#lib/game/cards.js';
	import { cn } from '#lib/utils.js';
	import type { Token } from '#lib/types.js';

	let { token, class: className = '' }: { token: Token; class?: string } = $props();
</script>

{#if token}
	<span
		class={cn(
			'inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-primary px-1 font-display text-xs leading-none font-bold text-primary-foreground shadow-sm',
			className
		)}
		title={`Token ${tokenLabel(token)}: ${tokenMeaning(token)}`}
		aria-label={`Token ${tokenLabel(token)}`}>{tokenLabel(token)}</span
	>
{/if}
