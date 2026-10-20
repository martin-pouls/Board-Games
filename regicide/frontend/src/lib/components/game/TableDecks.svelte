<!--
	The Tavern deck (face down) and the discard pile (face up) as small card stacks in a row of
	their own above the enemy, the count on the card and the name beside it. The Castle deck is the
	track in the enemy panel. Fixed sizes: a changing count never moves anything; an empty deck is a
	dashed outline.
-->
<script lang="ts">
	import { cn } from '#lib/utils.js';
	import type { PlayerView } from '#lib/types.js';

	let { view, class: className }: { view: PlayerView; class?: string } = $props();

	const decks = $derived([
		{ name: 'Tavern', title: 'Tavern deck', count: view.tavernCount, back: true },
		{ name: 'Discard', title: 'Discard pile', count: view.discardCount, back: false }
	]);
</script>

<ul class={cn('-my-1 flex h-10 shrink-0 items-center gap-7 px-1.5', className)} aria-label="Decks">
	{#each decks as d (d.name)}
		<li class="flex items-center gap-2.5" title={d.title}>
			<span
				class={cn(
					'deck grid aspect-[5/7] w-6.5 place-items-center font-display text-base leading-none font-extrabold tabular-nums',
					d.count === 0
						? 'border-[1.5px] border-dashed border-border text-muted-foreground'
						: d.back
							? 'back bg-back text-back-ink [--deck-edge:var(--back-edge)] [--deck:var(--back)]'
							: 'bg-stock text-suit-black [--deck-edge:var(--stock-edge)] [--deck:var(--stock)]',
					d.count >= 2 && 'layer-1',
					d.count >= 3 && 'layer-2'
				)}
				aria-hidden="true"
			>
				<span class="relative">{d.count}</span>
			</span>
			<span class="text-sm text-muted-foreground"
				>{d.name}<span class="sr-only">: {d.count}</span></span
			>
		</li>
	{/each}
</ul>

<style>
	.deck {
		position: relative;
		border-radius: 9% / 6.43%;
		box-shadow: 0 0 0 1px var(--deck-edge);
	}
	/* The cards underneath, offset down and right: one for two cards, two for more. */
	.deck.layer-1 {
		box-shadow:
			0 0 0 1px var(--deck-edge),
			1.5px 1.5px 0 -0.5px var(--deck),
			1.5px 1.5px 0 0.5px var(--deck-edge);
	}
	.deck.layer-2 {
		box-shadow:
			0 0 0 1px var(--deck-edge),
			1.5px 1.5px 0 -0.5px var(--deck),
			1.5px 1.5px 0 0.5px var(--deck-edge),
			3px 3px 0 -0.5px var(--deck),
			3px 3px 0 0.5px var(--deck-edge);
	}
	/* The back's frame and hatching. */
	.back::before {
		content: '';
		position: absolute;
		inset: 2.5px;
		border-radius: 2px;
		border: 1px solid var(--back-line);
		background: repeating-linear-gradient(
			45deg,
			transparent 0 3px,
			color-mix(in oklab, var(--back-line) 45%, transparent) 3px 4px
		);
	}
	/* A solid patch behind the count, so the hatching doesn't run through the digits. */
	.back > span {
		padding: 1px 2px;
		border-radius: 2px;
		background: var(--back);
	}
</style>
