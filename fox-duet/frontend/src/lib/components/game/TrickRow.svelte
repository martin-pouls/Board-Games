<script lang="ts">
	import Crown from '@lucide/svelte/icons/crown';
	import { getCardMotion, type Offstage } from '@board-games/shared/cardMotion.js';
	import { cardKey, SUIT_NAME, who } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import type { Card, PlayerView } from '#lib/types.js';
	import PlayingCard from './PlayingCard.svelte';

	let { view }: { view: PlayerView } = $props();

	const motion = getCardMotion();
	const { send, receive } = motion;

	interface Slot {
		card: Card | null;
		playerId: string;
	}

	const leaderId = $derived(view.trick?.lead ? view.trick.leaderId : view.lastTrick?.leaderId);
	const followerId = $derived(view.players.find((p) => p.id !== leaderId)?.id ?? '');
	const live = $derived(!!view.trick?.lead);
	const slots = $derived<Slot[]>(
		live
			? [
					{ card: view.trick!.lead, playerId: view.trick!.leaderId },
					{ card: view.trick!.follow, playerId: followerId }
				]
			: view.lastTrick
				? [
						{ card: view.lastTrick.lead, playerId: view.lastTrick.leaderId },
						{ card: view.lastTrick.follow, playerId: followerId }
					]
				: [
						{ card: null, playerId: view.trick?.leaderId ?? '' },
						{ card: null, playerId: '' }
					]
	);
	const winnerId = $derived(live ? view.trick?.winnerId : view.lastTrick?.winnerId);

	/** The viewer sits at the bottom, the teammate at the top. */
	const side = (playerId: string): Offstage =>
		view.players.find((p) => p.id === playerId)?.isYou ? 'bottom' : 'top';
</script>

<section class="flex items-end justify-center gap-3" aria-label="Table">
	<div class="flex flex-col items-center gap-1">
		<div class="relative h-24 w-[4.25rem]">
			{#if view.decree}
				{#key cardKey(view.decree)}
					<div
						class="absolute inset-0"
						in:receive={{ key: cardKey(view.decree), offstage: 'top' }}
						out:send={{ key: cardKey(view.decree), offstage: 'none' }}
					>
						<PlayingCard card={view.decree} class="ring-2 ring-fox/70" />
					</div>
				{/key}
			{/if}
		</div>
		<span class="text-xs font-semibold text-fox">
			Decree{view.decree ? ` · ${SUIT_NAME[view.decree.suit]} trump` : ''}
		</span>
	</div>

	<div class="mx-1 h-20 w-px self-center bg-border"></div>

	{#each slots as slot, i (i)}
		<div class="flex flex-col items-center gap-1">
			<div
				class={cn(
					'relative h-24 w-[4.25rem] rounded-lg border-2 border-dashed border-foreground/20',
					!live && view.lastTrick && 'opacity-80'
				)}
			>
				{#if slot.card}
					{#key cardKey(slot.card)}
						<div
							class="absolute -inset-0.5"
							in:receive={{ key: cardKey(slot.card), offstage: side(slot.playerId) }}
							out:send={{
								key: cardKey(slot.card),
								offstage: winnerId ? side(winnerId) : 'none'
							}}
						>
							<PlayingCard card={slot.card} />
						</div>
					{/key}
				{/if}
				{#if slot.card && winnerId === slot.playerId}
					<span
						class="absolute -top-2.5 -right-2.5 grid size-6 place-items-center rounded-full bg-fox text-primary-foreground shadow"
						title="Wins the trick"
					>
						<Crown class="size-3.5" />
					</span>
				{/if}
			</div>
			<span class="max-w-20 truncate text-xs text-muted-foreground">
				{slot.playerId ? who(view, slot.playerId) : ''}
				{i === 0 && slot.playerId ? '· lead' : ''}
			</span>
		</div>
	{/each}
</section>
