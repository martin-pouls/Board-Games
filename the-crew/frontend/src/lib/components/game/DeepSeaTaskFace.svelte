<!--
	The face of a Deep Sea task card, laid out like the printed card (taskFace.ts): a small head,
	rows of small cards, orange numbers and crew icons on sea-blue paper. The full sentence is the
	title. The difficulty pip and the status marks are TaskCard's, so they sit the same on the
	face-down card back.
-->
<script lang="ts">
	import Crown from '@lucide/svelte/icons/crown';
	import User from '@lucide/svelte/icons/user';
	import Users from '@lucide/svelte/icons/users';
	import { taskFace } from '#lib/game/taskFace.js';
	import { cn } from '#lib/utils.js';
	import type { Task } from '#lib/types.js';
	import TaskMini from './TaskMini.svelte';

	let {
		task,
		text,
		players,
		size = 'task',
		class: className = ''
	}: { task: Task; text: string; players: number; size?: 'xs' | 'task'; class?: string } = $props();

	const xs = $derived(size === 'xs');
	const face = $derived(task.objective ? taskFace(task.objective, players, task.predicted) : null);
	/** Kinds without a picture show the sentence. */
	const plain = $derived(!face || face.rows.length === 0);
	/** Only words (trick positions): set them larger, as printed. */
	const wordy = $derived(
		!!face && face.rows.every((r) => r.every((g) => g.t === 'text' && !g.big))
	);
</script>

<div
	data-face
	class={cn(
		'task-face relative flex shrink-0 flex-col items-center overflow-hidden border border-white/80 text-center text-task-ink shadow-sm select-none',
		xs
			? 'h-[3.5rem] w-[4.5rem] gap-px rounded-[0.3rem] px-0.5 pt-0.5'
			: 'h-[5.6rem] w-[5.25rem] gap-0.5 rounded-md px-1 pt-1 pb-0.5',
		className
	)}
	title={text}
>
	{#if plain}
		<p
			class={cn(
				'my-auto leading-[1.15] font-semibold break-words',
				xs ? 'line-clamp-4 text-[0.5rem]' : 'line-clamp-6 text-[0.62rem]'
			)}
		>
			{text}
		</p>
	{:else if face}
		{#if face.head}
			<!-- Indented on the left, clear of the difficulty pip. -->
			<p
				class={cn(
					'leading-none font-medium opacity-75',
					xs ? 'pl-2 text-[0.48rem]' : 'pl-2.5 text-[0.56rem]'
				)}
			>
				{face.head}
			</p>
		{/if}
		<div class="flex flex-1 flex-col items-center justify-center gap-0.5">
			{#each face.rows as row, r (r)}
				<div class={cn('flex flex-wrap items-center justify-center', xs ? 'gap-0.5' : 'gap-1')}>
					{#each row as seg, i (i)}
						{#if seg.t === 'text'}
							<span
								class={cn(
									'leading-[1.1]',
									seg.big
										? cn('font-display font-bold', xs ? 'text-sm' : 'text-lg')
										: cn(
												'font-semibold',
												xs ? 'text-[0.5rem]' : wordy ? 'text-[0.75rem]' : 'text-[0.62rem]'
											)
								)}>{seg.text}</span
							>
						{:else if seg.t === 'cards'}
							<span
								class={cn(
									'flex items-center',
									!seg.join && seg.cards.length > 2 && seg.cards.length < 5 && '-space-x-1',
									seg.cards.length > 4 && 'gap-px'
								)}
							>
								{#each seg.cards as c, k (k)}
									{#if k > 0 && seg.join}
										<span
											class={cn('px-0.5 font-semibold', xs ? 'text-[0.42rem]' : 'text-[0.5rem]')}
											>{seg.join}</span
										>
									{/if}
									<TaskMini
										card={c}
										size={seg.cards.length > 4 ? 'xs' : size}
										small={seg.cards.length > 3}
									/>
								{/each}
							</span>
						{:else if seg.t === 'num'}
							<span
								class={cn(
									'rounded-[0.2rem] bg-task-badge px-1 font-display leading-tight font-bold text-white shadow-sm',
									xs ? 'text-[0.6rem]' : 'text-sm'
								)}>{seg.value}</span
							>
						{:else if seg.icon === 'me'}
							<User class={xs ? 'size-2.5' : 'size-3.5'} strokeWidth={2.75} />
						{:else if seg.icon === 'others'}
							<Users class={xs ? 'size-3' : 'size-4'} strokeWidth={2.5} />
						{:else if seg.icon === 'crew'}
							<span class="flex"
								><Users class={xs ? 'size-3' : 'size-4'} strokeWidth={2.5} /><Users
									class={xs ? 'size-3' : 'size-4'}
									strokeWidth={2.5}
								/></span
							>
						{:else}
							<Crown class={xs ? 'size-2.5' : 'size-3.5'} strokeWidth={2.5} />
						{/if}
					{/each}
				</div>
			{/each}
		</div>
		{#if face.note && !xs}
			<p class="text-[0.45rem] leading-none opacity-70">{face.note}</p>
		{/if}
	{/if}
</div>
