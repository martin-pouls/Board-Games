<script lang="ts">
	import type { Snippet } from 'svelte';
	import Zap from '@lucide/svelte/icons/zap';
	import InfinityIcon from '@lucide/svelte/icons/infinity';
	import { commandInfo } from '#lib/game/commands.js';
	import { cn } from '#lib/utils.js';
	import type { CommandName } from '#lib/types.js';
	import CommandIcon from './CommandIcon.svelte';

	let {
		command,
		muted = false,
		size = 'sm',
		label,
		class: className = '',
		children
	}: {
		command: CommandName;
		/** The command has no effect any more (lightning card from an earlier turn). */
		muted?: boolean;
		size?: 'sm' | 'xs';
		/** Replaces the short chip text (e.g. the full label). */
		label?: string;
		class?: string;
		/** Extra content after the label (e.g. the pile direction). */
		children?: Snippet;
	} = $props();

	const info = $derived(commandInfo(command));
	const icon = $derived(size === 'xs' ? 'size-3' : 'size-3.5');
</script>

<span
	class={cn(
		'inline-flex max-w-full min-w-0 items-center gap-0.5 rounded-full leading-none font-semibold whitespace-nowrap',
		size === 'xs' ? 'px-1.5 py-0.5 text-[0.65rem]' : 'px-2 py-1 text-xs',
		info.kind === 'lightning'
			? 'bg-red-500/15 text-red-800 dark:text-red-200'
			: 'bg-violet-500/15 text-violet-800 dark:text-violet-200',
		muted && 'opacity-50',
		className
	)}
	title={muted ? 'No effect any more' : `${info.label}: ${info.description}`}
>
	{#if info.kind === 'lightning'}
		<Zap class={cn(icon, 'shrink-0 opacity-70')} aria-hidden="true" />
	{:else}
		<InfinityIcon class={cn(icon, 'shrink-0 opacity-70')} aria-hidden="true" />
	{/if}
	{#if command !== 'exactlyThree'}
		<!-- the 3! glyph would just repeat its label -->
		<CommandIcon {command} class={icon} />
	{/if}
	<span class="truncate">{label ?? info.short}</span>
	{@render children?.()}
</span>
