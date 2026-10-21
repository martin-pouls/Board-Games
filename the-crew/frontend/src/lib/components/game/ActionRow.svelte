<!--
	The button row between your row and your hand while cards are played: one 44 px row. The main
	part says what happens now: a clear button once a card is picked ("Play 3"), an outlined hint
	while it is your turn ("Lead a card", "Pick a blue card"), plain text while someone else plays
	("Bob is playing"). Communicate sits beside it at every trick start until you have used it.
	Tapping it starts the communicate flow: the hand offers the cards that can be shown, the picked
	card decides the statement ("Show 2 as lowest"), and ✕ cancels.
-->
<script lang="ts">
	import X from '@lucide/svelte/icons/x';
	import { Button } from '#lib/components/ui/button/index.js';
	import { terms } from '#lib/edition.js';
	import { cardLabel, sameCard, SUIT_NAMES } from '#lib/game/cards.js';
	import { seatName } from '#lib/game/format.js';
	import type { Attempt, Card, CommOption } from '#lib/types.js';
	import CardName from './CardName.svelte';
	import EditionGlyph from './EditionGlyph.svelte';

	let {
		attempt,
		selected,
		communicating,
		shown,
		disabled,
		onplay,
		oncommunicate,
		oncancel,
		onshow
	}: {
		attempt: Attempt;
		/** The card picked in the hand or among the dummy's stacks. */
		selected: Card | null;
		/** The communicate flow is open. */
		communicating: boolean;
		/** The card picked to show. */
		shown: Card | null;
		/** Busy or disconnected. */
		disabled: boolean;
		onplay: (c: Card) => void;
		oncommunicate: () => void;
		oncancel: () => void;
		onshow: (o: CommOption) => void;
	} = $props();

	const name = (s: number) => seatName(attempt.members, s, attempt.seat);
	const mine = $derived(attempt.actors.includes(attempt.seat));
	const turn = $derived(attempt.members[attempt.turn]);
	const forDummy = $derived(mine && !!turn?.dummy);
	const leading = $derived(attempt.trick.length === 0);
	const led = $derived(attempt.trick[0]?.card.suit);
	/** Following: every playable card is of the led colour. */
	const mustFollow = $derived(!!led && attempt.playable.every((c) => c.suit === led));
	const forText = $derived(forDummy ? ` for ${terms.dummy}` : '');

	const canCommunicate = $derived(attempt.actions.communicate.length > 0);
	const option = $derived(
		shown ? attempt.actions.communicate.find((o) => sameCard(o.card, shown)) : undefined
	);
	const AS = { highest: ' as highest', only: ' as only card', lowest: ' as lowest', '': '' };

	/** Someone else acts: who, and whether they lead. */
	const waiting = $derived.by(() => {
		if (!turn) return '';
		const actor = attempt.actors[0];
		const who = turn.dummy && actor !== undefined ? name(actor) : name(turn.seat);
		const dummy = turn.dummy ? ` for ${terms.dummy}` : '';
		return leading ? `${who} leads${dummy}` : `${who} is playing${dummy}`;
	});
	const hint = $derived(
		leading
			? `Lead a card${forText}`
			: mustFollow
				? `Pick a ${SUIT_NAMES[led!]} card${forText}`
				: `Pick a card${forText}`
	);
</script>

<div class="flex h-11 shrink-0 gap-2">
	{#if communicating}
		<Button
			variant="outline"
			class="size-11 rounded-[0.6875rem] border-border bg-panel dark:bg-panel"
			aria-label="Cancel"
			title="Cancel"
			onclick={oncancel}
		>
			<X class="size-5" />
		</Button>
		{#if option}
			<Button
				class="h-11 min-w-0 flex-1 gap-1.5 rounded-[0.6875rem] text-base font-semibold"
				{disabled}
				aria-label={`Show ${cardLabel(option.card)}${AS[option.position]}`}
				onclick={() => onshow(option)}
			>
				Show <CardName card={option.card} />{AS[option.position]}
			</Button>
		{:else}
			{@render idle('Pick a card to show')}
		{/if}
	{:else if mine && selected}
		<Button
			class="h-11 min-w-0 flex-1 gap-1.5 rounded-[0.6875rem] text-base font-semibold"
			{disabled}
			aria-label={`Play ${cardLabel(selected)}${forText}`}
			onclick={() => onplay(selected)}
		>
			Play <CardName card={selected} />{forText}
		</Button>
	{:else if mine}
		{@render idle(hint)}
	{:else}
		<p class="flex min-w-0 flex-1 items-center justify-center text-sm text-muted-foreground">
			{waiting}
		</p>
	{/if}
	{#if canCommunicate && !communicating}
		<Button
			variant="outline"
			class="h-11 gap-1.5 rounded-[0.6875rem] border-border bg-panel pr-3.5 pl-[11px] text-sm font-semibold dark:bg-panel"
			{disabled}
			onclick={oncommunicate}
		>
			<EditionGlyph kind="comm" class="size-5" />Communicate
		</Button>
	{/if}
</div>

{#snippet idle(text: string)}
	<p
		class="flex min-w-0 flex-1 items-center justify-center rounded-[0.6875rem] border-[1.5px] border-primary bg-primary/8 text-base font-semibold"
	>
		{text}
	</p>
{/snippet}
