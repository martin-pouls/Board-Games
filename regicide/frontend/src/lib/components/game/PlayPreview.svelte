<!--
	What the selected cards would do, as badges on one line: the suit powers (struck through where
	the enemy's immunity blocks them), a Jester's effect, a defeat. Or why the cards can't be played.
-->
<script lang="ts">
	import Skull from '@lucide/svelte/icons/skull';
	import { playProblemText, suitGlyph } from '#lib/game/format.js';
	import type { PlayPreview, SuitEffect } from '#lib/game/rules.js';
	import { cn } from '#lib/utils.js';
	import JesterHat from './JesterHat.svelte';

	let { preview }: { preview: PlayPreview } = $props();

	const check = $derived(preview.check);

	/** A blocked power is shown without its amount, struck through. */
	function label(e: SuitEffect): string {
		const n = e.blocked ? '' : ` ${e.amount}`;
		switch (e.suit) {
			case 'hearts':
				return `Heal${n}`;
			case 'diamonds':
				return `Draw${n}`;
			case 'clubs':
				return '×2';
			case 'spades':
				return `Shield${n}`;
		}
	}

	const badge =
		'flex shrink-0 items-center gap-0.5 rounded-full border bg-card px-[7px] text-xs leading-[18px] whitespace-nowrap';
</script>

{#if check.ok}
	<ul class="flex gap-1">
		{#if check.kind === 'jester'}
			<li class={cn(badge, 'border-jester/60 text-jester')}>
				<JesterHat class="size-3.5" />
				{preview.cancelsImmunity ? 'Cancel immunity' : 'Jester'}
			</li>
		{/if}
		{#each preview.effects as e (e.suit)}
			<li
				class={cn(badge, e.blocked ? 'text-muted-foreground line-through' : 'border-primary/50')}
				title={e.blocked ? "Blocked by the enemy's immunity" : undefined}
			>
				<i class="suit-glyph">{suitGlyph(e.suit)}&#xFE0E;</i>
				{label(e)}
				{#if e.blocked}<span class="sr-only">(blocked by immunity)</span>{/if}
			</li>
		{/each}
		{#if preview.defeats}
			<li
				class={cn(badge, 'border-health/60')}
				title={preview.exact ? 'Exact: the enemy goes onto the Tavern deck' : undefined}
			>
				<Skull class="size-3.5 text-health" aria-hidden="true" />
				{preview.exact ? 'Exact' : 'Defeat'}
			</li>
		{/if}
	</ul>
{:else}
	<p class="line-clamp-2 text-right text-xs leading-tight whitespace-normal text-muted-foreground">
		{playProblemText(check.problem)}
	</p>
{/if}
