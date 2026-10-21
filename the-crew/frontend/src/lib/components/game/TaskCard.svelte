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
	{#if deepSea}
		<span class="relative block">
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
	{:else}
		<!-- The task card on its square task token: the order mark (a hollow square without one), then
		     ✓ or ✗. The token hangs below the card, inside this box. -->
		<span
			class={cn(
				'relative block rounded-[0.25rem] pb-2',
				selected && 'ring-2 ring-primary outline-3 outline-offset-2 outline-primary/25'
			)}
		>
			{#if task.card}
				<PlayingCard card={task.card} {size} />
			{:else}
				<CardBack {size} label="Face-down task" />
			{/if}
			{#if task.status !== 'open'}
				<span class="absolute inset-x-0 top-0 bottom-2 z-[3] rounded-[inherit] bg-background/40"
				></span>
			{/if}
			<span
				class={cn(
					'absolute bottom-0 left-1/2 z-[4] grid h-3.5 min-w-4 -translate-x-1/2 place-items-center rounded-[3px] px-[3px] font-display text-[10px] leading-none font-bold shadow-[0_0_0_1.5px_var(--tab-ring,var(--background))]',
					task.status === 'fulfilled'
						? 'bg-success text-ok-ink'
						: task.status === 'failed'
							? 'bg-danger text-ok-ink'
							: 'bg-chip text-chip-ink'
				)}
				aria-hidden="true"
			>
				{#if task.status === 'fulfilled'}
					<Check class="size-2.5" strokeWidth={3.5} />
				{:else if task.status === 'failed'}
					<X class="size-2.5" strokeWidth={3.5} />
				{:else if task.token}
					{tokenLabel(task.token)}
				{:else}
					<span class="size-1.5 rounded-[1.5px] border-[1.5px] border-chip-ink opacity-70"></span>
				{/if}
			</span>
		</span>
	{/if}
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
			deepSea &&
				(selected
					? '-translate-y-1 [&_[data-face]]:ring-3 [&_[data-face]]:ring-signal [&_[role=img]]:ring-3 [&_[role=img]]:ring-signal'
					: 'hover:-translate-y-0.5')
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
