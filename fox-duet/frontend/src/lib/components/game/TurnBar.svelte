<script lang="ts">
	import LoaderCircle from '@lucide/svelte/icons/loader-circle';
	import { Button } from '#lib/components/ui/button/index.js';
	import type { GameClient } from '#lib/game/client.js';
	import { cardLabel, endLabel, moveLabel, seats, statusText } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import type { Card, PlayerView } from '#lib/types.js';

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
	const mine = $derived(!!pending?.playerIds.includes(view.you));
	const { mate } = $derived(seats(view));
	const disabled = $derived(busy || !connected);
	const status = $derived(statusText(view));
</script>

<div
	class="flex min-h-14 flex-col justify-center gap-1.5 rounded-xl bg-muted/60 px-3 py-2"
	role="status"
>
	<p
		class={cn(
			'flex items-center gap-2 leading-snug',
			mine ? 'font-semibold' : 'text-muted-foreground'
		)}
	>
		{#if !mine}<LoaderCircle class="size-4 shrink-0 animate-spin" />{/if}
		{status}
	</p>

	{#if mine && pending}
		{#if pending.step === 'lead' || pending.step === 'follow'}
			{#if selected}
				<Button
					class="h-11 text-base"
					{disabled}
					onclick={() => run(() => client.playCard(selected))}
				>
					Play {cardLabel(selected)}
				</Button>
			{/if}
		{:else if pending.step === 'foxesChoose'}
			<div class="grid grid-cols-2 gap-2">
				<Button
					class="h-11"
					{disabled}
					onclick={() => run(() => client.chooseFoxesPlayer(view.you))}
				>
					Me
				</Button>
				<Button
					class="h-11"
					variant="secondary"
					disabled={disabled || !mate}
					onclick={() => run(() => client.chooseFoxesPlayer(mate!.id))}
				>
					{mate?.name ?? 'Teammate'}
				</Button>
			</div>
		{:else if pending.step === 'foxesSwap'}
			<div class="grid grid-cols-2 gap-2">
				<Button
					class="h-11"
					disabled={disabled || !selected}
					onclick={() => selected && run(() => client.swapDecree(selected))}
				>
					{selected ? `Swap ${cardLabel(selected)}` : 'Pick a card'}
				</Button>
				<Button
					class="h-11"
					variant="secondary"
					{disabled}
					onclick={() => run(() => client.keepDecree())}
				>
					Keep decree
				</Button>
			</div>
		{:else if pending.step === 'gift'}
			<Button
				class="h-11 text-base"
				disabled={disabled || !selected}
				onclick={() => selected && run(() => client.giveGift(selected))}
			>
				{selected
					? `Give ${cardLabel(selected)} to ${mate?.name ?? 'your teammate'}`
					: 'Pick a card to give'}
			</Button>
		{:else if pending.step === 'move'}
			<div class="flex flex-col gap-1.5">
				{#each pending.moveOptions as o (o.delta)}
					{@const label = moveLabel(view, o)}
					<Button
						class={cn(
							'h-auto min-h-11 justify-between gap-2 py-1.5',
							highlight === o.delta && 'ring-3 ring-fox'
						)}
						variant={o.lost ? 'destructive' : 'secondary'}
						{disabled}
						onclick={() => run(() => client.chooseMove(o.delta))}
						onpointerenter={() => (highlight = o.delta)}
						onpointerleave={() => (highlight = null)}
						onfocus={() => (highlight = o.delta)}
						onblur={() => (highlight = null)}
					>
						<span class="font-semibold">{label.title}</span>
						<span class="text-xs font-normal opacity-80">{label.result}</span>
					</Button>
				{/each}
			</div>
		{:else if pending.step === 'forestEnd'}
			<div class="grid grid-cols-2 gap-2">
				{#each pending.ends as end (end)}
					<Button
						class="h-11"
						variant="secondary"
						{disabled}
						onclick={() => run(() => client.chooseForestEnd(end))}
					>
						Cover {endLabel(view, end)}
					</Button>
				{/each}
			</div>
		{/if}
	{/if}
</div>
