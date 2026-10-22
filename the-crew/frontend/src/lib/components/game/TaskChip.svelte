<!--
	A Deep Sea task in a row or the draft: a low chip of printed task paper with the short phrase and
	the small cards as printed (taskChip.ts), 28 px high. A done task fades under a green ✓ on its
	corner, a failed one gets a red ✗. Tapping it picks it when it can be picked now (draft,
	assignment), otherwise it opens the full card (TaskSheet). A face-down task is a dark chip.
-->
<script lang="ts">
	import Check from '@lucide/svelte/icons/check';
	import Crown from '@lucide/svelte/icons/crown';
	import User from '@lucide/svelte/icons/user';
	import Users from '@lucide/svelte/icons/users';
	import X from '@lucide/svelte/icons/x';
	import { taskChip } from '#lib/game/taskChip.js';
	import { getTaskSheet } from '#lib/game/taskSheet.js';
	import { taskTextOf } from '#lib/game/taskText.js';
	import { cn } from '#lib/utils.js';
	import type { Task } from '#lib/types.js';
	import TaskMini from './TaskMini.svelte';

	let {
		task,
		players,
		selected = false,
		onpick,
		still = false,
		class: className = ''
	}: {
		task: Task;
		/** The crew members (the dummy included), for the numbers in the sentence. */
		players: number;
		selected?: boolean;
		/** Makes a tap pick the task instead of opening its card. */
		onpick?: () => void;
		/** Not tappable at all (inside a button). */
		still?: boolean;
		class?: string;
	} = $props();

	const sheet = getTaskSheet();
	const segs = $derived(task.objective ? taskChip(task.objective, task.predicted) : []);
	const text = $derived(taskTextOf(task, players));
	const label = $derived(
		[text, task.status === 'fulfilled' ? 'done' : task.status === 'failed' ? 'failed' : '']
			.filter(Boolean)
			.join(', ')
	);
	const chipClass = $derived(
		cn(
			'chip inline-flex h-7 max-w-full items-center gap-[3px] overflow-hidden rounded-md px-1.5 text-[0.78rem] leading-none font-bold whitespace-nowrap',
			task.objective ? 'task-face text-task-ink' : 'bg-rocket text-rocket-ink',
			selected && 'outline-2 outline-offset-1 outline-primary'
		)
	);
	const tap = $derived(
		still ? undefined : (onpick ?? (sheet ? () => sheet.show(task) : undefined))
	);
</script>

{#snippet content()}
	{#if !task.objective}
		<span class="font-semibold">Face down</span>
	{/if}
	{#each segs as seg, i (i)}
		{#if seg.t === 'text'}
			<span class="px-px">{seg.text}</span>
		{:else if seg.t === 'card'}
			<span class="flex items-center">
				<TaskMini card={seg.card} size="chip" />
			</span>
		{:else if seg.icon === 'captain'}
			<Crown class="size-3.5 fill-current" strokeWidth={2} />
		{:else if seg.icon === 'me'}
			<User class="size-3.5" strokeWidth={2.75} />
		{:else}
			<Users class="size-3.5" strokeWidth={2.5} />
		{/if}
	{/each}
{/snippet}

<span class={cn('relative inline-flex max-w-full shrink-0', className)}>
	{#if tap}
		<button
			type="button"
			class={cn(
				chipClass,
				'focus-visible:ring-3 focus-visible:ring-ring focus-visible:outline-none'
			)}
			title={text}
			aria-label={label}
			aria-pressed={onpick ? selected : undefined}
			onclick={tap}
		>
			{@render content()}
		</button>
	{:else}
		<span class={chipClass} title={text}>{@render content()}</span>
	{/if}
	{#if task.status === 'fulfilled'}
		<span class="pointer-events-none absolute -inset-px z-[3] rounded-[0.4375rem] bg-background/40"
		></span>
	{/if}
	{#if task.status !== 'open'}
		<span
			class={cn(
				'absolute -top-1.5 -right-1.5 z-[4] grid size-[0.9375rem] place-items-center rounded-full text-ok-ink shadow-[0_0_0_1.5px_var(--tab-ring,var(--background))]',
				task.status === 'fulfilled' ? 'bg-success' : 'bg-danger'
			)}
			aria-hidden="true"
		>
			{#if task.status === 'fulfilled'}
				<Check class="size-2.5" strokeWidth={3.5} />
			{:else}
				<X class="size-2.5" strokeWidth={3.5} />
			{/if}
		</span>
	{/if}
</span>

<style>
	/* The printed paper's white edge and a soft shadow. */
	.chip {
		box-shadow:
			0 0 0 1px oklch(1 0 0 / 85%),
			0 1px 2px var(--card-shadow);
	}
</style>
