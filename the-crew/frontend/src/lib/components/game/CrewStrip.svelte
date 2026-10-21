<script lang="ts">
	import { getCardMotion } from '@board-games/shared/cardMotion.js';
	import { cardKey, sameCard } from '#lib/game/cards.js';
	import { terms } from '#lib/edition.js';
	import { cn } from '#lib/utils.js';
	import type { Attempt, Card, Member } from '#lib/types.js';
	import CardBack from './CardBack.svelte';
	import MemberRow from './MemberRow.svelte';
	import PlayingCard from './PlayingCard.svelte';
	import type { TaskPick } from './TaskList.svelte';

	let {
		attempt,
		online,
		jarvisEnabled = () => false,
		selected = null,
		onpick,
		taskPick
	}: {
		attempt: Attempt;
		/** Player IDs with a live connection. */
		online: string[];
		/** Whether the viewer can pick this dummy card now (the commander on the dummy's turn). */
		jarvisEnabled?: (c: Card) => boolean;
		selected?: Card | null;
		onpick?: (c: Card) => void;
		taskPick?: TaskPick;
	} = $props();

	const { send, receive } = getCardMotion();

	// The other members in seat order, starting left of the viewer.
	const others = $derived.by(() => {
		const n = attempt.members.length;
		const out: Member[] = [];
		for (let i = 1; i < n; i++) out.push(attempt.members[(attempt.seat + i) % n]);
		return out;
	});
	const humans = $derived(others.filter((m) => !m.dummy));
	const jarvis = $derived(others.find((m) => m.dummy));
	const turnSeat = $derived(
		attempt.step === 'play' || attempt.step === 'draft' ? attempt.turn : -1
	);
	const anyJarvisEnabled = $derived(!!jarvis?.stacks.some((s) => s.up && jarvisEnabled(s.up)));
</script>

<!-- Task tokens and radio tokens sit on the panel: their outline takes its colour. -->
<section class="flex flex-col gap-2 [--tab-ring:var(--panel)]" aria-label="Crew">
	{#if humans.length}
		<ul class={cn('grid gap-2', humans.length < 3 ? 'grid-cols-1' : 'grid-cols-2')}>
			{#each humans as m (m.seat)}
				<li
					class={cn(
						'rounded-xl bg-panel px-2 py-1.5 ring-2 ring-transparent transition-shadow',
						(m.mustAct || turnSeat === m.seat) && 'ring-signal/70'
					)}
				>
					<MemberRow
						{attempt}
						member={m}
						online={online.includes(m.playerId)}
						{taskPick}
						class={humans.length >= 3 ? 'flex-wrap' : ''}
					/>
				</li>
			{/each}
		</ul>
	{/if}

	{#if jarvis}
		<div
			class={cn(
				'flex flex-col gap-1.5 rounded-xl bg-panel px-2 py-1.5 ring-2 ring-transparent',
				(turnSeat === jarvis.seat || anyJarvisEnabled) && 'ring-signal/70'
			)}
		>
			<MemberRow {attempt} member={jarvis} {taskPick} />
			<!-- In the debrief its cards are listed with everyone's hand. -->
			{#if !attempt.result}
				<ul class="flex justify-between gap-[5px]" aria-label={`${terms.dummyPossessive} stacks`}>
					{#each jarvis.stacks as s, i (i)}
						{@const on = !!s.up && jarvisEnabled(s.up)}
						<!-- A hidden card under the face-up one shows as a dark edge below it. -->
						<li
							class={cn(
								'relative h-[3.75rem] w-[2.625rem] shrink-0 rounded-md',
								s.hidden && s.up && 'shadow-[0_3px_0_-1px_var(--rocket)]'
							)}
						>
							{#if s.up}
								{@const c = s.up}
								{#key cardKey(c)}
									<button
										type="button"
										class={cn(
											'absolute inset-0 rounded-md transition-[translate] duration-150 focus-visible:ring-3 focus-visible:ring-ring focus-visible:outline-none',
											sameCard(c, selected) &&
												'-translate-y-1.5 shadow-[0_3px_8px_var(--card-shadow)] ring-2 ring-primary'
										)}
										disabled={!on}
										aria-pressed={sameCard(c, selected)}
										onclick={() => onpick?.(c)}
										in:receive={{ key: cardKey(c), offstage: 'none' }}
										out:send={{ key: cardKey(c), offstage: 'none' }}
									>
										<PlayingCard card={c} size="stack" />
										{#if anyJarvisEnabled && !on}
											<span class="absolute inset-0 z-[5] rounded-md bg-background/58"></span>
										{/if}
									</button>
								{/key}
							{:else if s.hidden}
								<CardBack size="stack" label="Hidden card" />
							{:else}
								<span class="absolute inset-0 rounded-md border-[1.5px] border-dashed border-border"
								></span>
							{/if}
						</li>
					{/each}
				</ul>
			{/if}
		</div>
	{/if}
</section>
