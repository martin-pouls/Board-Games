<!--
	The turn bar above the action row, like The Game's: whose turn in bold, then what matters for
	it (the selected cards' effects, the damage to cover). Your own turn is outlined in the primary
	colour. It slides in when the turn passes to someone else.
-->
<script lang="ts">
	import { fly } from 'svelte/transition';
	import { prefersReducedMotion } from 'svelte/motion';
	import type { Snippet } from 'svelte';
	import { cn } from '#lib/utils.js';

	let {
		title,
		mine,
		turnKey,
		hideTitle = false,
		children
	}: {
		title: string;
		mine: boolean;
		/** Changes when the turn passes on. */
		turnKey: string;
		/** Keeps the title for screen readers only, when the right side needs the room. */
		hideTitle?: boolean;
		children?: Snippet;
	} = $props();
</script>

{#key turnKey}
	<div
		in:fly={{ y: 8, duration: prefersReducedMotion.current ? 0 : 200 }}
		class={cn(
			'flex h-10 shrink-0 items-center gap-2 overflow-hidden rounded-xl border pr-2 pl-3 whitespace-nowrap',
			mine
				? 'border-primary bg-[color-mix(in_oklab,var(--primary)_9%,var(--card))] shadow-[inset_0_0_0_1px_var(--primary)]'
				: 'bg-card'
		)}
		aria-live="polite"
	>
		<b class={cn('shrink-0 text-sm font-bold', hideTitle && 'sr-only')}>{title}</b>
		<div class="ml-auto flex min-w-0 items-center justify-end">
			{@render children?.()}
		</div>
	</div>
{/key}
