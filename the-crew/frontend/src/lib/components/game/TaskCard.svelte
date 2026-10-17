<script lang="ts">
	import Check from '@lucide/svelte/icons/check';
	import X from '@lucide/svelte/icons/x';
	import { cardLabel, tokenLabel, tokenMeaning } from '#lib/game/cards.js';
	import { taskTextOf } from '#lib/game/taskText.js';
	import { cn } from '#lib/utils.js';
	import type { Task } from '#lib/types.js';
	import CardBack from './CardBack.svelte';
	import DeepSeaTaskFace from './DeepSeaTaskFace.svelte';
	import PlayingCard from './PlayingCard.svelte';
	import TokenBadge from './TokenBadge.svelte';

	let {
		task,
		owner = '',
		size = 'task',
		selected = false,
		players = 3,
		onpick
	}: {
		task: Task;
		/** The owner's name, shown under the card ("" hides it). */
		owner?: string;
		size?: 'xs' | 'task';
		selected?: boolean;
		/** Deep Sea: the crew members (the dummy included), for the numbers in the card text. */
		players?: number;
		/** Makes the card a button. */
		onpick?: () => void;
	} = $props();

	/** A Deep Sea task card: a text card, or face down with only its difficulty. */
	const deepSea = $derived(!task.card && (!!task.objective || task.difficulty > 0));
	const text = $derived(deepSea ? taskTextOf(task, players) : '');

	const label = $derived(
		[
			deepSea
				? task.objective
					? `Task: ${text}, difficulty ${task.difficulty}`
					: `Face-down task, difficulty ${task.difficulty}`
				: task.card
					? `Task ${cardLabel(task.card)}`
					: 'Face-down task',
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
		{:else if deepSea}
			{#if task.objective}
				<DeepSeaTaskFace
					{task}
					{text}
					{players}
					{size}
					class={cn(task.status === 'failed' && 'opacity-60 grayscale')}
				/>
			{:else}
				<CardBack
					{size}
					label={`Face-down task, difficulty ${task.difficulty}`}
					class={size === 'xs' ? 'h-[3.5rem] w-[4.5rem]' : 'h-[5.6rem] w-[5.25rem]'}
				/>
			{/if}
			<span
				class="absolute -top-1.5 -left-1.5 grid size-5 place-items-center rounded-full border border-ink/40 bg-paper font-display text-[0.65rem] leading-none font-bold text-ink shadow"
				title={`Difficulty ${task.difficulty}`}
				aria-hidden="true">{task.difficulty}</span
			>
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
			class={cn(
				'block truncate text-center text-[0.65rem] leading-tight text-muted-foreground',
				deepSea ? 'max-w-20' : 'max-w-12'
			)}
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
				? '-translate-y-1 [&_[data-face]]:ring-3 [&_[data-face]]:ring-signal [&_[role=img]]:ring-3 [&_[role=img]]:ring-signal'
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
