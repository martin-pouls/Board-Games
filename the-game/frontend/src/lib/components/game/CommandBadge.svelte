<script lang="ts">
	import { commandInfo } from '#lib/game/commands.js';
	import { cn } from '#lib/utils.js';
	import type { CommandName } from '#lib/types.js';
	import CommandIcon from './CommandIcon.svelte';

	/**
	 * A command as printed on its card: Infinity commands in a blue square, lightning commands in
	 * an orange triangle pointing down. Decorative; the surrounding element carries the text.
	 */
	let {
		command,
		size = 'page',
		class: className = ''
	}: {
		command: CommandName;
		/** `card`: on a card face, scaled with the card's `--w`. `page`: 24 px, on the page. */
		size?: 'card' | 'page';
		class?: string;
	} = $props();

	const lightning = $derived(commandInfo(command).kind === 'lightning');
</script>

<span class={cn('badge', size, lightning ? 'lightning' : 'infinity', className)} aria-hidden="true">
	{#if lightning}
		<svg class="tri" viewBox="0 0 30 25">
			<path d="M3 3h24L15 22Z" stroke-width="4" stroke-linejoin="round" />
		</svg>
	{/if}
	<CommandIcon {command} class="glyph" />
</span>

<style>
	.badge {
		position: relative;
		display: inline-grid;
		flex: none;
		place-items: center;
	}
	.tri {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
	}
	.badge :global(.glyph) {
		position: relative;
	}

	.card.infinity {
		width: calc(var(--w) * 0.27);
		height: calc(var(--w) * 0.27);
		border-radius: calc(var(--w) * 0.065);
		background: var(--infinity);
		color: var(--infinity-ink);
	}
	.card.infinity :global(.glyph) {
		width: calc(var(--w) * 0.17);
		height: calc(var(--w) * 0.17);
	}
	.card.lightning {
		width: calc(var(--w) * 0.34);
		height: calc(var(--w) * 0.28);
		color: var(--lightning-ink);
	}
	.card.lightning path {
		fill: var(--lightning);
		stroke: var(--lightning);
	}
	.card.lightning :global(.glyph) {
		width: calc(var(--w) * 0.15);
		height: calc(var(--w) * 0.15);
		margin-top: calc(var(--w) * -0.08);
	}

	.page.infinity {
		width: 24px;
		height: 24px;
		border-radius: 6px;
		background: var(--infinity-page);
		color: var(--infinity-ink);
	}
	.page.infinity :global(.glyph) {
		width: 15px;
		height: 15px;
	}
	.page.lightning {
		width: 29px;
		height: 24px;
		color: var(--lightning-ink);
	}
	.page.lightning path {
		fill: var(--lightning-page);
		stroke: var(--lightning-page);
	}
	.page.lightning :global(.glyph) {
		width: 12px;
		height: 12px;
		margin-top: -6px;
	}
</style>
