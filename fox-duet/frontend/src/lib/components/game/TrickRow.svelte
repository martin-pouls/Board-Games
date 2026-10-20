<script lang="ts">
	import Crown from '@lucide/svelte/icons/crown';
	import { getCardMotion, type Offstage } from '@board-games/shared/cardMotion.js';
	import { cardKey, SUIT_NAME } from '#lib/game/format.js';
	import type { Card, PlayerView } from '#lib/types.js';
	import PlayingCard from './PlayingCard.svelte';
	import SuitIcon from './SuitIcon.svelte';

	let { view }: { view: PlayerView } = $props();

	const motion = getCardMotion();
	const { send, receive } = motion;

	interface Slot {
		card: Card | null;
		playerId: string;
	}

	const leaderId = $derived(
		view.trick?.lead || !view.lastTrick ? view.trick?.leaderId : view.lastTrick.leaderId
	);
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
						{ card: null, playerId: view.trick ? followerId : '' }
					]
	);
	const winnerId = $derived(live ? view.trick?.winnerId : view.lastTrick?.winnerId);

	const name = (playerId: string) => {
		const p = view.players.find((x) => x.id === playerId);
		return p?.isYou ? 'You' : (p?.name ?? '');
	};

	/** The viewer sits at the bottom, the teammate at the top. */
	const side = (playerId: string): Offstage =>
		view.players.find((p) => p.id === playerId)?.isYou ? 'bottom' : 'top';
</script>

<!--
	The decree card (small, with the trump suit and the trick count), then the trick: lead and
	follow, or the completed trick until the next card is led, with a crown on the winner.
-->
<section class="flex items-end justify-center gap-3" aria-label="Table">
	<div class="flex flex-col items-center gap-1 text-xs text-muted-foreground">
		<div class="relative h-[75.6px] w-[54px]">
			{#if view.decree}
				{#key cardKey(view.decree)}
					<div
						class="absolute inset-0"
						in:receive={{ key: cardKey(view.decree), offstage: 'top' }}
						out:send={{ key: cardKey(view.decree), offstage: 'none' }}
					>
						<PlayingCard card={view.decree} size="decree" />
					</div>
				{/key}
			{/if}
		</div>
		{#if view.decree}
			<span
				class="flex items-center gap-1 font-extrabold text-foreground"
				title="{SUIT_NAME[view.decree.suit]} are trump"
			>
				Trump <SuitIcon suit={view.decree.suit} class="size-3.5" />
				<span class="sr-only">{SUIT_NAME[view.decree.suit]}</span>
			</span>
		{/if}
		{#if view.trick?.number}
			<span>Trick <b class="font-extrabold text-foreground">{view.trick.number}</b>/11</span>
		{/if}
	</div>

	<div class="mt-2 mb-5 w-px self-stretch bg-border"></div>

	{#each slots as slot, i (i)}
		{@const won = !!slot.card && winnerId === slot.playerId}
		<div class="flex flex-col items-center gap-1">
			<div
				class="relative h-[117.6px] w-[84px] rounded-[9px] border-2 border-dashed border-foreground/20 min-[390px]:h-[126px] min-[390px]:w-[90px]"
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
				{#if won}
					<span
						class="absolute top-1 right-1 z-10 grid size-6 place-items-center rounded-full bg-fox text-primary-foreground shadow"
						title="Wins the trick"
					>
						<Crown class="size-3.5" />
					</span>
				{/if}
			</div>
			<span class="max-w-24 truncate text-xs text-muted-foreground">
				{#if slot.playerId}
					<b class="font-extrabold text-foreground">{name(slot.playerId)}</b>
					{#if won}won{:else if i === 0 && slot.card}led{:else if i === 0}{name(slot.playerId) ===
						'You'
							? 'lead'
							: 'leads'}{/if}
				{/if}
			</span>
		</div>
	{/each}
</section>
