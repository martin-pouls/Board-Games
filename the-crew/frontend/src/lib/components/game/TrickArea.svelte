<script lang="ts">
	import Crown from '@lucide/svelte/icons/crown';
	import History from '@lucide/svelte/icons/history';
	import { getCardMotion, type Offstage } from '@board-games/shared/cardMotion.js';
	import { cardKey } from '#lib/game/cards.js';
	import { seatName } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import type { Attempt, Card } from '#lib/types.js';
	import PlayingCard from './PlayingCard.svelte';

	let { attempt }: { attempt: Attempt } = $props();

	const motion = getCardMotion();
	const { send, receive } = motion;

	let peek = $state(false);

	interface Slot {
		seat: number;
		card: Card | null;
	}

	const n = $derived(attempt.members.length);
	const side3 = $derived(n <= 3);
	const live = $derived(attempt.trick.length > 0);
	// Between tricks (and in the debrief) the last trick stays on the table; during a trick the
	// peek button shows it instead of the trick in progress.
	const showLast = $derived(!!attempt.lastTrick && (!live || peek));
	const leader = $derived(showLast ? attempt.lastTrick!.leader : attempt.trickLeader);
	const plays = $derived(showLast ? attempt.lastTrick!.plays : attempt.trick);
	const winner = $derived(showLast ? attempt.lastTrick!.winner : -1);
	const slots = $derived<Slot[]>(
		Array.from({ length: n }, (_, i) => {
			const seat = (leader + i) % n;
			return { seat, card: plays.find((p) => p.seat === seat)?.card ?? null };
		})
	);

	/** The viewer acts now (also for JARVIS): their empty slot is marked. */
	const mine = $derived(attempt.step === 'play' && attempt.actors.includes(attempt.seat));

	// A trick in progress hides the peek again once a card is played.
	const playCount = $derived(attempt.trick.length);
	$effect(() => {
		void playCount;
		peek = false;
	});

	/** The viewer sits at the bottom; everyone else, JARVIS included, comes from the top. */
	const side = (seat: number): Offstage => (seat === attempt.seat ? 'bottom' : 'top');
	const dummySeat = $derived(attempt.members.find((m) => m.dummy)?.seat ?? -1);
	// JARVIS's cards fly from its stacks (matched by key); a human's from the edge.
	const from = (seat: number): Offstage => (seat === dummySeat ? 'none' : side(seat));
</script>

<!-- Up to three cards leave room beside the trick for its label; with more, it sits on top. -->
<section
	class={cn('grid items-start gap-1', side3 ? 'grid-cols-[3.375rem_1fr_3.375rem]' : 'grid-cols-1')}
	aria-label="Trick"
>
	<div
		class={cn(
			'flex text-xs text-muted-foreground',
			side3 ? 'flex-col items-start gap-1.5 pt-0.5' : 'items-center justify-between'
		)}
	>
		<span class="font-display leading-tight font-bold tracking-wide uppercase">
			{#if showLast}
				Last trick
			{:else}
				Trick {Math.min(attempt.trickNumber, attempt.totalTricks)}/{attempt.totalTricks}
			{/if}
		</span>
		{#if live && attempt.lastTrick}
			<button
				type="button"
				class={cn(
					'flex w-max items-center gap-[3px] rounded-md border border-border px-[5px] py-px hover:text-foreground',
					peek && 'bg-muted text-foreground'
				)}
				aria-pressed={peek}
				title="Show the last trick"
				onclick={() => (peek = !peek)}
			>
				<History class="size-3" />Last
			</button>
		{:else if showLast}
			<span>{attempt.trickNumber - 1}/{attempt.totalTricks}</span>
		{/if}
	</div>
	<ol class="flex w-full justify-center gap-1.5">
		{#each slots as slot, i (slot.seat)}
			{@const yours = !showLast && !slot.card && slot.seat === attempt.turn && mine}
			<li class="flex min-w-0 flex-col items-center gap-0.5">
				<div
					class={cn(
						'relative h-[4.625rem] w-[3.25rem] rounded-md border-[1.5px] border-dashed border-border',
						showLast && 'border-transparent',
						yours && 'border-solid border-primary ring-3 ring-primary/22'
					)}
				>
					{#if slot.card}
						{@const c = slot.card}
						{#key cardKey(c)}
							<div
								class={cn('absolute -inset-[1.5px]', showLast && 'opacity-80')}
								in:receive={{ key: cardKey(c), offstage: from(slot.seat) }}
								out:send={{ key: cardKey(c), offstage: winner >= 0 ? side(winner) : 'none' }}
							>
								<PlayingCard card={c} size="trick" />
							</div>
						{/key}
					{/if}
				</div>
				<span
					class={cn(
						'flex max-w-[3.75rem] items-center gap-0.5 truncate text-xs leading-tight',
						slot.seat === attempt.seat ? 'font-bold text-foreground' : 'text-muted-foreground'
					)}
				>
					{#if slot.card && slot.seat === winner}
						<Crown class="size-3 shrink-0 fill-current text-captain" aria-label="Won the trick" />
					{/if}
					{#if i === 0}<span class="sr-only">Lead: </span><span aria-hidden="true">▸</span>{/if}
					<span class="truncate">{seatName(attempt.members, slot.seat, attempt.seat)}</span>
				</span>
			</li>
		{/each}
	</ol>
</section>
