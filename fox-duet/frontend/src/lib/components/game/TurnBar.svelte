<!--
	The action area, as in The Game and Regicide: a turn bar (whose turn in bold, on your turn what
	to do; outlined in fox orange when it's yours), then one row with the buttons for the decision,
	or who we're waiting for. Both rows keep their height, so nothing above them moves.
-->
<script lang="ts" module>
	import { cn } from '#lib/utils.js';

	/** The classes of the turn bar; the game-over controls use them too. */
	export const barClass = (mine: boolean) =>
		cn(
			'flex h-10 shrink-0 items-center gap-2 overflow-hidden rounded-xl border px-3 whitespace-nowrap',
			mine
				? 'border-primary bg-[color-mix(in_oklab,var(--primary)_10%,var(--card))] shadow-[inset_0_0_0_1px_var(--primary)]'
				: 'bg-card'
		);
</script>

<script lang="ts">
	import { fly } from 'svelte/transition';
	import { prefersReducedMotion } from 'svelte/motion';
	import { Button } from '#lib/components/ui/button/index.js';
	import type { GameClient } from '#lib/game/client.js';
	import { endLabel, moveLabel, seats, towardMe, turnStatus } from '#lib/game/format.js';
	import type { Card, MoveOption, PlayerView } from '#lib/types.js';
	import SuitIcon from './SuitIcon.svelte';

	let {
		client,
		view,
		selected,
		busy,
		connected,
		run,
		highlight = $bindable(null)
	}: {
		client: GameClient;
		view: PlayerView;
		/** The card picked in the hand. */
		selected: Card | null;
		busy: boolean;
		connected: boolean;
		run: (action: () => Promise<void>) => Promise<void>;
		/** The move option under the pointer, shared with the board. */
		highlight?: number | null;
	} = $props();

	const pending = $derived(view.pending);
	const status = $derived(turnStatus(view));
	const { me, mate } = $derived(seats(view));
	const disabled = $derived(busy || !connected);
	const turnKey = $derived(
		`${pending?.step}:${pending?.playerIds.join(',')}:${view.trick?.number}`
	);
	/** Three move options get short labels, so they fit side by side. */
	const compact = $derived((pending?.moveOptions.length ?? 0) > 2);

	function moveSteps(o: MoveOption): number {
		return towardMe(o.delta, me.end);
	}
</script>

{#snippet cardName(c: Card)}
	{c.rank}<SuitIcon suit={c.suit} class="size-3.5" />
{/snippet}

{#snippet moveOutcome(o: MoveOption)}
	{#if o.lost}
		<span class="text-xs">lost</span>
	{:else if o.beyond}
		<svg viewBox="0 0 24 24" class="size-3.5 fill-current" aria-hidden="true">
			<path d="M12 2 5 11h3.5L4 17.5h6.5V22h3v-4.5H20L15.5 11H19Z" />
		</svg>
	{:else if o.gem}
		<svg viewBox="0 0 24 24" class="h-3.5 w-3 fill-gem" aria-hidden="true">
			<path d="M12 2 21 10 12 22 3 10Z" />
		</svg>
	{/if}
{/snippet}

{#if status && pending}
	<div class="flex shrink-0 flex-col gap-1.5">
		{#key turnKey}
			<div
				in:fly={{ y: 8, duration: prefersReducedMotion.current ? 0 : 200 }}
				class={barClass(status.mine)}
				aria-live="polite"
			>
				<b class="shrink-0 text-sm font-extrabold">{status.title}</b>
				{#if status.hint}
					<span class="ml-auto min-w-0 truncate text-xs text-muted-foreground">{status.hint}</span>
				{/if}
			</div>
		{/key}

		<div class="flex h-11 gap-2">
			{#if !status.mine}
				<p class="flex flex-1 items-center justify-center text-sm text-muted-foreground">
					<span class="truncate">{status.waiting}</span>
				</p>
			{:else if pending.step === 'lead' || pending.step === 'follow'}
				<Button
					class="h-11 flex-1 text-base"
					disabled={disabled || !selected}
					onclick={() => selected && run(() => client.playCard(selected))}
				>
					{#if selected}Play {@render cardName(selected)}{:else}Pick a card to play{/if}
				</Button>
			{:else if pending.step === 'foxesChoose'}
				<Button
					class="h-11 flex-1"
					{disabled}
					onclick={() => run(() => client.chooseFoxesPlayer(view.you))}
				>
					Me
				</Button>
				<Button
					class="h-11 min-w-0 flex-1"
					variant="secondary"
					disabled={disabled || !mate}
					onclick={() => mate && run(() => client.chooseFoxesPlayer(mate.id))}
				>
					<span class="truncate">{mate?.name ?? 'Teammate'}</span>
				</Button>
			{:else if pending.step === 'foxesSwap'}
				<Button
					class="h-11 flex-1"
					disabled={disabled || !selected}
					onclick={() => selected && run(() => client.swapDecree(selected))}
				>
					{#if selected}Swap {@render cardName(selected)}{:else}Pick a card{/if}
				</Button>
				<Button
					class="h-11 flex-1"
					variant="secondary"
					{disabled}
					onclick={() => run(() => client.keepDecree())}
				>
					Keep decree
				</Button>
			{:else if pending.step === 'gift'}
				<Button
					class="h-11 min-w-0 flex-1 text-base"
					disabled={disabled || !selected}
					onclick={() => selected && run(() => client.giveGift(selected))}
				>
					{#if selected}
						Give {@render cardName(selected)}
						<span class="truncate">to {mate?.name ?? 'your teammate'}</span>
					{:else}
						Pick a card to give
					{/if}
				</Button>
			{:else if pending.step === 'move'}
				{#each pending.moveOptions as o (o.delta)}
					{@const label = moveLabel(view, o)}
					{@const steps = moveSteps(o)}
					{@const n = Math.abs(steps)}
					<Button
						class={cn(
							'h-11 min-w-0 flex-1 gap-1.5 px-2',
							highlight === o.delta && 'ring-3 ring-fox'
						)}
						variant={o.lost ? 'destructive' : 'secondary'}
						aria-label="{label.title}: {label.result}"
						title="{label.title}: {label.result}"
						{disabled}
						onclick={() => run(() => client.chooseMove(o.delta))}
						onpointerenter={() => (highlight = o.delta)}
						onpointerleave={() => (highlight = null)}
						onfocus={() => (highlight = o.delta)}
						onblur={() => (highlight = null)}
					>
						{#if steps === 0}
							Stay
						{:else if steps > 0}
							← {compact ? n : `${n} ${n === 1 ? 'space' : 'spaces'}`}
						{:else}
							{compact ? n : `${n} ${n === 1 ? 'space' : 'spaces'}`} →
						{/if}
						{@render moveOutcome(o)}
					</Button>
				{/each}
			{:else if pending.step === 'forestEnd'}
				{#each pending.ends as end (end)}
					<Button
						class="h-11 min-w-0 flex-1"
						variant="secondary"
						{disabled}
						onclick={() => run(() => client.chooseForestEnd(end))}
					>
						<span class="truncate">Cover {endLabel(view, end)}</span>
					</Button>
				{/each}
			{/if}
		</div>
	</div>
{/if}
