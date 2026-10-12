<!-- What the selected cards would do, as badges: the suit powers, a Jester's effect, a defeat. -->
<script lang="ts">
	import Drama from '@lucide/svelte/icons/drama';
	import Skull from '@lucide/svelte/icons/skull';
	import { playProblemText, suitGlyph, suitPowerName } from '#lib/game/format.js';
	import type { PlayPreview, SuitEffect } from '#lib/game/rules.js';
	import { cn } from '#lib/utils.js';

	let { preview }: { preview: PlayPreview } = $props();

	const check = $derived(preview.check);

	function label(e: SuitEffect): string {
		switch (e.suit) {
			case 'hearts':
				return `Heal ${e.amount}`;
			case 'diamonds':
				return `Draw ${e.amount}`;
			case 'clubs':
				return '×2';
			case 'spades':
				return `Shield ${e.amount}`;
		}
	}

	const badge = 'flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs whitespace-nowrap';
</script>

{#if check.ok}
	<ul class="flex flex-wrap gap-1">
		{#if check.kind === 'jester'}
			<li class={cn(badge, 'border-jester/60 text-jester')}>
				<Drama class="size-3.5" aria-hidden="true" />
				{preview.cancelsImmunity ? 'Cancel immunity' : 'Jester'}
			</li>
		{/if}
		{#each preview.effects as e (e.suit)}
			<li
				class={cn(
					badge,
					e.blocked ? 'text-muted-foreground line-through opacity-70' : 'border-primary/50'
				)}
				title={e.blocked ? 'Immune' : undefined}
			>
				<span style="font-family: system-ui, sans-serif">{suitGlyph(e.suit)}&#xFE0E;</span>
				{e.blocked ? suitPowerName(e.suit) : label(e)}
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
	<p class="text-xs leading-tight text-muted-foreground">{playProblemText(check.problem)}</p>
{/if}
