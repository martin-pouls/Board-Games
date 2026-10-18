<script lang="ts">
	import type { Snippet } from 'svelte';
	import InfinityIcon from '@lucide/svelte/icons/infinity';
	import Zap from '@lucide/svelte/icons/zap';
	import { commandInfo } from '#lib/game/commands.js';
	import { cn } from '#lib/utils.js';
	import type { CommandName } from '#lib/types.js';
	import CommandBadge from './CommandBadge.svelte';

	/**
	 * A card face: navy, the value large in the middle and small in two corners, a command as its
	 * printed badge on top with its kind (Infinity, lightning) beside the corner values. Sized by
	 * the inherited `--w` (card width); everything scales with it. Decorative: the element around
	 * it carries the label.
	 */
	let {
		value,
		command,
		mode = 'face',
		inactive = false,
		overlap = false,
		covered = false,
		class: className = '',
		children
	}: {
		value: number;
		command?: CommandName;
		/** `face`: a full card. `start`: an empty pile's start value as a dashed outline. `mini`: 28 px, value only. */
		mode?: 'face' | 'start' | 'mini';
		/** The command has no effect any more (lightning card from an earlier turn): badge faded. */
		inactive?: boolean;
		/** Lies on the card to its left (overlapping hand): a shadow on its left edge. */
		overlap?: boolean;
		/** The next card covers its right part: no bottom corner value (only a sliver would show). */
		covered?: boolean;
		class?: string;
		/** Rendered inside the card (e.g. NewCardMark). */
		children?: Snippet;
	} = $props();

	const kind = $derived(command ? commandInfo(command).kind : undefined);
</script>

{#snippet index(corner: 'tl' | 'br')}
	<span class={cn('ix font-display', corner)}>
		{value}
		{#if kind === 'infinity'}
			<InfinityIcon class="km infinity" strokeWidth={3} />
		{:else if kind === 'lightning'}
			<Zap class="km lightning" fill="currentColor" strokeWidth={0} />
		{/if}
	</span>
{/snippet}

<span class={cn('game-card', mode, overlap && 'overlap', className)} aria-hidden="true">
	{#if mode === 'face'}
		{@render index('tl')}
		{#if !covered}{@render index('br')}{/if}
		{#if command}
			<CommandBadge {command} size="card" class={cn('bd', inactive && 'opacity-40')} />
		{/if}
	{:else if mode === 'mini' && kind}
		{#if kind === 'infinity'}
			<InfinityIcon class="km mini-km infinity" strokeWidth={3} />
		{:else}
			<Zap class="km mini-km lightning" fill="currentColor" strokeWidth={0} />
		{/if}
	{/if}
	<span class="v font-display">{value}</span>
	{@render children?.()}
</span>

<style>
	.game-card {
		position: relative;
		display: block;
		flex: none;
		width: var(--w);
		height: calc(var(--w) * 1.4);
		border-radius: calc(var(--w) * 0.1);
		background: var(--card-face);
		color: var(--card-ink);
		box-shadow:
			inset 0 0 0 1px var(--card-edge),
			0 1px 2px var(--card-shade);
		font-variant-numeric: tabular-nums;
	}
	.overlap {
		box-shadow:
			inset 0 0 0 1px var(--card-edge),
			-3px 0 5px var(--card-shade);
	}
	.ix {
		position: absolute;
		top: calc(var(--w) * 0.06);
		left: calc(var(--w) * 0.1);
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: calc(var(--w) * 0.02);
		font-size: calc(var(--w) * 0.2);
		line-height: 0.9;
	}
	.ix.br {
		top: auto;
		left: auto;
		bottom: calc(var(--w) * 0.06);
		right: calc(var(--w) * 0.1);
		transform: rotate(180deg);
	}
	.game-card :global(.km) {
		display: block;
		width: calc(var(--w) * 0.13);
		height: calc(var(--w) * 0.13);
	}
	.game-card :global(.km.infinity) {
		color: var(--infinity);
	}
	.game-card :global(.km.lightning) {
		color: var(--lightning);
	}
	.game-card :global(.bd) {
		position: absolute;
		top: calc(var(--w) * 0.07);
		left: 50%;
		transform: translateX(-50%);
	}
	/* The numeral sits at the same height on every card, badge or not. */
	.v {
		position: absolute;
		left: 0;
		right: 0;
		top: calc(var(--w) * 0.7);
		transform: translateY(-44%);
		text-align: center;
		font-size: calc(var(--w) * 0.66);
		line-height: 1;
	}

	/* An empty pile: its start value as a dashed outline, no card yet. */
	.start {
		background: none;
		box-shadow: none;
		border: 1.5px dashed color-mix(in oklch, var(--pile) 55%, transparent);
		color: var(--pile-label);
		opacity: 0.85;
	}

	.mini {
		--w: 28px;
		border-radius: 5px;
	}
	.mini .v {
		top: 50%;
		transform: translateY(-46%);
		font-size: 20px;
	}
	.game-card :global(.mini-km) {
		position: absolute;
		top: 2px;
		right: 2px;
		width: 8px;
		height: 8px;
	}
</style>
