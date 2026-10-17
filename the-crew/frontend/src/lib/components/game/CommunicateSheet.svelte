<script lang="ts">
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import { terms } from '#lib/edition.js';
	import { cardKey, cardLabel, sameCard } from '#lib/game/cards.js';
	import { positionText } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import type { CommOption } from '#lib/types.js';
	import PlayingCard from './PlayingCard.svelte';

	let {
		open = $bindable(false),
		options,
		deadZone,
		tokensLeft = -1,
		disabled,
		oncommunicate
	}: {
		open?: boolean;
		/** The cards the viewer may communicate now, each with its true statement. */
		options: CommOption[];
		deadZone: boolean;
		/** Rapture of the deep: the crew's shared tokens left; -1 otherwise (once per attempt). */
		tokensLeft?: number;
		disabled: boolean;
		/** Resolves to false when the server refused; the sheet then stays open. */
		oncommunicate: (o: CommOption) => Promise<boolean>;
	} = $props();

	let chosen = $state<CommOption | null>(null);
	$effect(() => {
		if (!open) chosen = null;
	});

	async function confirm() {
		if (!chosen) return;
		if (await oncommunicate(chosen)) open = false;
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="max-h-[85dvh] overflow-y-auto">
		<Dialog.Header>
			<Dialog.Title>Communicate a card</Dialog.Title>
			<Dialog.Description>
				{tokensLeft >= 0
					? `Rapture of the deep: takes one of the crew's ${terms.commToken}s (${tokensLeft} left); you may communicate again while any are left.`
					: 'Once per attempt.'}
				The card goes face up in front of you until you play it.
				{deadZone
					? 'Dead zone: nobody learns whether it is your highest, only or lowest card of its colour.'
					: 'Everyone sees whether it is your highest, only or lowest card of its colour.'}
			</Dialog.Description>
		</Dialog.Header>
		{#if options.length}
			<ul class="flex flex-wrap justify-center gap-2 py-1">
				{#each options as o (cardKey(o.card))}
					{@const on = sameCard(o.card, chosen?.card)}
					<li>
						<button
							type="button"
							class={cn(
								'flex w-[4.5rem] flex-col items-center gap-1 rounded-lg p-1 text-center text-xs leading-tight focus-visible:ring-3 focus-visible:ring-ring focus-visible:outline-none',
								on ? 'bg-signal/15 ring-2 ring-signal' : 'hover:bg-muted'
							)}
							aria-pressed={on}
							onclick={() => (chosen = o)}
						>
							<PlayingCard card={o.card} size="sm" />
							<span>{o.position || '?'}</span>
						</button>
					</li>
				{/each}
			</ul>
		{:else}
			<p class="text-sm text-muted-foreground">
				None of your cards can be communicated: only your highest, lowest or only card of a colour,
				never a {terms.trump}.
			</p>
		{/if}
		{#if chosen}
			<p class="text-center text-sm">
				<b>{cardLabel(chosen.card)}</b>: {positionText(chosen.position)}
			</p>
		{/if}
		<Dialog.Footer>
			<Button class="h-11 w-full" disabled={disabled || !chosen} onclick={confirm}>
				{chosen ? 'Communicate' : 'Pick a card'}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
