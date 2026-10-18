<script lang="ts">
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import { COMMANDS } from '#lib/game/commands.js';
	import type { CommandName } from '#lib/types.js';
	import CommandBadge from './CommandBadge.svelte';

	let {
		open = $bindable(false),
		extreme = false
	}: {
		open?: boolean;
		/** The game uses the Extreme command cards: list them too. */
		extreme?: boolean;
	} = $props();

	const commands = Object.entries(COMMANDS) as [CommandName, (typeof COMMANDS)[CommandName]][];
	const lightning = commands.filter(([, c]) => c.kind === 'lightning');
	const infinity = commands.filter(([, c]) => c.kind === 'infinity');
</script>

{#snippet commandList(list: typeof commands)}
	<ul class="flex flex-col gap-1">
		{#each list as [name, c] (name)}
			<li class="flex items-start gap-2">
				<CommandBadge command={name} class="-mt-0.5" />
				<span><b>{c.label}:</b> {c.description}</span>
			</li>
		{/each}
	</ul>
{/snippet}

<Dialog.Root bind:open>
	<Dialog.Content class="max-h-[85dvh] overflow-y-auto">
		<Dialog.Header>
			<Dialog.Title>Rules at a glance</Dialog.Title>
			<Dialog.Description>
				Play all 98 cards together. Never reveal or ask for concrete numbers.
			</Dialog.Description>
		</Dialog.Header>
		<div class="flex flex-col gap-3 text-sm leading-snug">
			<section>
				<h3 class="mb-1 font-semibold">Piles</h3>
				<ul class="list-disc pl-5">
					<li><b>Up (1 → 99):</b> each card must be higher than the top card.</li>
					<li><b>Down (100 → 2):</b> each card must be lower than the top card.</li>
					<li>
						<b>Backwards trick:</b> exactly 10 lower on an up pile, or exactly 10 higher on a down pile.
						As often as you like.
					</li>
				</ul>
			</section>
			<section>
				<h3 class="mb-1 font-semibold">Your turn</h3>
				<ul class="list-disc pl-5">
					<li>Play at least 2 cards, on any piles, then as many more as you like.</li>
					<li>End your turn to draw back up to your hand size.</li>
					<li>Once the draw pile is empty: at least 1 card; empty hands are skipped.</li>
					<li>Until you end your turn, you can undo this turn's cards.</li>
					<li>A green <b>+1</b> on a pile counts the cards played there this turn.</li>
				</ul>
			</section>
			<section>
				<h3 class="mb-1 font-semibold">End of the game</h3>
				<p>
					All cards played: you win. If the player on turn can neither play a card nor end the turn,
					the game is lost; anyone can also give up at any time. The score is the cards left in
					hands and the draw pile: lower is better, fewer than 10 is excellent.
				</p>
			</section>
			<section>
				<h3 class="mb-1 font-semibold">Pro</h3>
				<p>
					At least 3 cards per turn (still 1 once the draw pile is empty). Optionally one card less
					in hand.
				</p>
			</section>
			{#if extreme}
				<section>
					<h3 class="mb-1 font-semibold">Lightning commands</h3>
					<p class="mb-1 text-muted-foreground">Only for the player who played it, this turn.</p>
					{@render commandList(lightning)}
				</section>
				<section>
					<h3 class="mb-1 font-semibold">Infinity commands</h3>
					<p class="mb-1 text-muted-foreground">
						For everyone, as long as the card is on top of a pile.
					</p>
					{@render commandList(infinity)}
				</section>
			{/if}
		</div>
	</Dialog.Content>
</Dialog.Root>
