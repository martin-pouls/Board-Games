<!--
	The draw pile as a face-down stack of navy cards at the start of the button row, right above the
	hand it refills. The count is printed right on the back, as on Regicide's decks; the framed,
	hatched back keeps it from reading as a card value. Fixed size; an empty pile is a dashed outline.
-->
<script lang="ts">
	import { cn } from '#lib/utils.js';

	let { count }: { count: number } = $props();
</script>

<span
	class="flex shrink-0 items-center pr-1"
	title={count > 0 ? 'Draw pile' : 'Draw pile empty'}
	role="img"
	aria-label={count > 0 ? `Draw pile: ${count} cards` : 'Draw pile empty'}
>
	<span
		class={cn(
			'deck grid h-[44.8px] w-8 place-items-center',
			count === 0 && 'empty',
			count >= 2 && 'layer-1',
			count >= 3 && 'layer-2'
		)}
	>
		<!-- A patch of the back's own navy behind the digits, so the hatching doesn't run through. -->
		<b
			class={cn(
				'relative z-1 rounded-sm px-0.5 font-display text-xl leading-none font-normal tabular-nums',
				count > 0 ? 'bg-card-face text-card-ink' : 'text-muted-foreground'
			)}>{count}</b
		>
	</span>
</span>

<style>
	.deck {
		position: relative;
		border-radius: 3.2px;
		background: var(--card-face);
		box-shadow: inset 0 0 0 1px var(--card-edge);
	}
	/* The cards underneath, offset down and right, each with a line of the page between. */
	.deck.layer-1 {
		box-shadow:
			inset 0 0 0 1px var(--card-edge),
			1.5px 1.5px 0 0 var(--background),
			2.5px 2.5px 0 0 var(--card-face);
	}
	.deck.layer-2 {
		box-shadow:
			inset 0 0 0 1px var(--card-edge),
			1.5px 1.5px 0 0 var(--background),
			2.5px 2.5px 0 0 var(--card-face),
			3.5px 3.5px 0 0 var(--background),
			4.5px 4.5px 0 0 var(--card-face);
	}
	/* The back: a thin frame with fine hatching. */
	.deck::before {
		content: '';
		position: absolute;
		inset: 4px;
		border-radius: 2px;
		border: 1px solid var(--card-edge);
		background: repeating-linear-gradient(45deg, transparent 0 3px, var(--card-edge) 3px 4px);
	}
	.deck.empty {
		background: none;
		box-shadow: none;
		border: 1.5px dashed color-mix(in oklch, var(--muted-foreground) 55%, transparent);
	}
	.deck.empty::before {
		display: none;
	}
</style>
