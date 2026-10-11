<script lang="ts">
	import Check from '@lucide/svelte/icons/check';
	import X from '@lucide/svelte/icons/x';
	import { cardLabel, tokenLabel, tokenMeaning } from '#lib/game/cards.js';
	import { cn } from '#lib/utils.js';
	import type { Task } from '#lib/types.js';
	import CardBack from './CardBack.svelte';
	import PlayingCard from './PlayingCard.svelte';
	import TokenBadge from './TokenBadge.svelte';

	let {
		task,
		owner = '',
		size = 'task',
		selected = false,
		onpick
	}: {
		task: Task;
		/** The owner's name, shown under the card ("" hides it). */
		owner?: string;
		size?: 'xs' | 'task';
		selected?: boolean;
		/** Makes the card a button. */
		onpick?: () => void;
	} = $props();

	const label = $derived(
		[
			task.card ? `Task ${cardLabel(task.card)}` : 'Face-down task',
			task.token && `token ${tokenLabel(task.token)} (${tokenMeaning(task.token)})`,
			owner && `for ${owner}`,
			task.status !== 'open' && task.status
		]
			.filter(Boolean)
			.join(', ')
	);
</script>

{#snippet face()}
	<span class="relative block">
		{#if task.card}
			<PlayingCard
				card={task.card}
				{size}
				class={cn(task.status === 'failed' && 'opacity-60 grayscale')}
			/>
		{:else}
			<CardBack {size} label="Face-down task" />
		{/if}
		{#if task.token}
			<TokenBadge
				token={task.token}
				class={cn('absolute -top-1.5 -right-1.5', size === 'xs' && 'h-4 min-w-4 text-[0.6rem]')}
			/>
		{/if}
		{#if task.status === 'fulfilled'}
			<span
				class="absolute -bottom-1 -left-1 grid size-4.5 place-items-center rounded-full bg-success text-white shadow"
			>
				<Check class="size-3" strokeWidth={3} />
			</span>
		{:else if task.status === 'failed'}
			<span
				class="absolute -bottom-1 -left-1 grid size-4.5 place-items-center rounded-full bg-danger text-white shadow"
			>
				<X class="size-3" strokeWidth={3} />
			</span>
		{/if}
	</span>
	{#if owner}
		<span
			class="block max-w-12 truncate text-center text-[0.65rem] leading-tight text-muted-foreground"
		>
			{owner}
		</span>
	{/if}
{/snippet}

{#if onpick}
	<button
		type="button"
		class={cn(
			'flex flex-col items-center gap-0.5 rounded-md transition-transform focus-visible:ring-3 focus-visible:ring-ring focus-visible:outline-none',
			selected
				? '-translate-y-1 [&_[role=img]]:ring-3 [&_[role=img]]:ring-signal'
				: 'hover:-translate-y-0.5'
		)}
		aria-label={label}
		aria-pressed={selected}
		onclick={onpick}
	>
		{@render face()}
	</button>
{:else}
	<div class="flex flex-col items-center gap-0.5" role="group" aria-label={label}>
		{@render face()}
	</div>
{/if}
