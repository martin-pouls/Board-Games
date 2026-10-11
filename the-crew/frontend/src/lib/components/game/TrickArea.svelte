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

	// A trick in progress hides the peek again once it changes.
	$effect(() => {
		void attempt.trick.length;
		peek = false;
	});

	/** The viewer sits at the bottom; everyone else, JARVIS included, comes from the top. */
	const side = (seat: number): Offstage => (seat === attempt.seat ? 'bottom' : 'top');
	const dummySeat = $derived(attempt.members.find((m) => m.dummy)?.seat ?? -1);
	// JARVIS's cards fly from its stacks (matched by key); a human's from the edge.
	const from = (seat: number): Offstage => (seat === dummySeat ? 'none' : side(seat));
</script>

<section class="flex flex-col items-center gap-1" aria-label="Trick">
	<div class="flex w-full items-center justify-between text-xs text-muted-foreground">
		<span class="font-display font-semibold tracking-wide uppercase">
			{#if showLast && (peek || attempt.step === '')}
				Last trick
			{:else if showLast}
				Trick {attempt.trickNumber - 1}
			{:else}
				Trick {Math.min(attempt.trickNumber, attempt.totalTricks)}/{attempt.totalTricks}
			{/if}
		</span>
		{#if live && attempt.lastTrick}
			<button
				type="button"
				class={cn(
					'flex items-center gap-1 rounded-md px-1.5 py-0.5 hover:text-foreground',
					peek && 'bg-muted text-foreground'
				)}
				aria-pressed={peek}
				onclick={() => (peek = !peek)}
			>
				<History class="size-3.5" />Last trick
			</button>
		{/if}
	</div>
	<ol class="flex w-full justify-center gap-1.5">
		{#each slots as slot, i (slot.seat)}
			<li class="flex min-w-0 flex-col items-center gap-0.5">
				<div
					class={cn(
						'relative h-[5.25rem] w-[3.6rem] rounded-lg border-2 border-dashed border-foreground/15',
						showLast && 'border-transparent'
					)}
				>
					{#if slot.card}
						{@const c = slot.card}
						{#key cardKey(c)}
							<div
								class={cn('absolute -inset-0.5', showLast && slot.seat !== winner && 'opacity-70')}
								in:receive={{ key: cardKey(c), offstage: from(slot.seat) }}
								out:send={{ key: cardKey(c), offstage: winner >= 0 ? side(winner) : 'none' }}
							>
								<PlayingCard card={c} />
							</div>
						{/key}
					{/if}
					{#if slot.card && slot.seat === winner}
						<span
							class="absolute -top-2.5 -right-2.5 grid size-6 place-items-center rounded-full bg-signal text-primary-foreground shadow"
							title="Won the trick"
						>
							<Crown class="size-3.5" />
						</span>
					{/if}
				</div>
				<span
					class={cn(
						'max-w-[3.75rem] truncate text-[0.7rem] leading-tight',
						slot.seat === attempt.seat ? 'font-semibold' : 'text-muted-foreground'
					)}
				>
					{seatName(attempt.members, slot.seat, attempt.seat)}
				</span>
				{#if i === 0}
					<span class="text-[0.6rem] leading-none text-muted-foreground uppercase">lead</span>
				{/if}
			</li>
		{/each}
	</ol>
</section>
