<!--
	A communicated card at the end of its owner's row: a little larger than a task, with a ring and
	the round radio token (Deep Sea: sonar token) at the top for highest, in the middle for only, at
	the bottom for lowest, and a "?" in the middle when the position is hidden (dead zone,
	currents). The ring keeps it apart from the tasks once the row fills up.
-->
<script lang="ts">
	import { cardLabel } from '#lib/game/cards.js';
	import { positionText } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import type { Communication } from '#lib/types.js';
	import PlayingCard from './PlayingCard.svelte';

	let { comm, class: className = '' }: { comm: Communication; class?: string } = $props();

	const title = $derived(`${cardLabel(comm.card)}: ${positionText(comm.position)}`);
	const PLACE = {
		highest: '-top-[5px]',
		only: 'top-[calc(50%-7px)]',
		lowest: '-bottom-[5px]',
		'': 'top-[calc(50%-7px)]'
	} as const;
</script>

<span class={cn('relative block shrink-0', className)} {title}>
	<span class="block rounded-[0.3125rem] ring-2 ring-radio">
		<PlayingCard card={comm.card} size="comm" />
	</span>
	<span
		class={cn(
			'absolute -right-[7px] z-[4] grid size-3.5 place-items-center rounded-full bg-radio font-sans text-[10px] leading-none font-extrabold text-radio-ink shadow-[0_0_0_2px_var(--tab-ring,var(--background)),0_1px_2px_var(--card-shadow)]',
			PLACE[comm.position]
		)}
		aria-hidden="true"
	>
		{#if comm.position}
			<span class="size-[3.5px] rounded-full bg-radio-ink/85"></span>
		{:else}
			?
		{/if}
	</span>
</span>
