<script lang="ts">
	import ArrowUp from '@lucide/svelte/icons/arrow-up';
	import ArrowDown from '@lucide/svelte/icons/arrow-down';
	import Skull from '@lucide/svelte/icons/skull';
	import { prefersReducedMotion } from 'svelte/motion';
	import { getCardMotion, type CardMotionParams } from '@board-games/shared/cardMotion.js';
	import { commandInfo } from '#lib/game/commands.js';
	import { pileTopKey } from '#lib/game/motion.js';
	import { cn } from '#lib/utils.js';
	import type { Pile } from '#lib/types.js';
	import GameCard from './GameCard.svelte';

	let {
		pile,
		highlight = false,
		dimmed = false,
		disabled = false,
		playsThisTurn = 0,
		bound = false,
		skull = false,
		onclick
	}: {
		pile: Pile;
		highlight?: boolean;
		dimmed?: boolean;
		disabled?: boolean;
		playsThisTurn?: number;
		/** Single pile: every further card this turn must go here. */
		bound?: boolean;
		/** A skull played this turn shows here and must be covered. */
		skull?: boolean;
		onclick?: () => void;
	} = $props();

	// Card motion: your own card flies in from your hand (same key there) and back on undo; other
	// players' cards come in from their hidden hands above (and go back up on their undo). A top
	// that gets covered stays underneath while the new card slides over it; one uncovered by an
	// undo fades back in place.
	const { send, receive } = getCardMotion();
	// The cards on the pile before the latest update (bookkeeping, not rendered).
	let lastCount = -1;
	// A keyed each of one, so a leaving top keeps its own key and count (a $derived would move on).
	const tops = $derived.by(() => {
		const count = pile.cardCount;
		const uncovered = count < lastCount;
		lastCount = count;
		return [
			{
				key: pileTopKey(pile),
				value: pile.topValue,
				command: pile.topCard?.command,
				active: pile.commandActive,
				start: !pile.topCard,
				count,
				uncovered
			}
		];
	});
	/** Longest card flight (the shared default): a covered top stays until the new one has landed. */
	const LANDING = 450;
	/**
	 * A leaving top flies only when an undo took it back. A covered one stays put underneath until
	 * the new card has slid over it.
	 */
	function leave(node: Element, params: CardMotionParams & { taken: boolean }) {
		if (params.taken) return send(node, params);
		return () => ({
			duration: prefersReducedMotion.current ? 0 : LANDING,
			css: () => 'z-index: 0'
		});
	}

	const up = $derived(pile.direction === 'up');
	const empty = $derived(pile.topCard === null);
	const command = $derived(pile.topCard?.command);
	const label = $derived(
		`${up ? 'Ascending' : 'Descending'} pile, ${empty ? 'empty' : `top card ${pile.topValue}`}` +
			(command
				? `, ${commandInfo(command).label}${pile.commandActive ? ' (active)' : ' (no effect any more)'}`
				: '') +
			(playsThisTurn > 0 ? `, ${playsThisTurn} played this turn` : '') +
			(skull ? ', skull must be covered' : '') +
			(bound ? ', all cards this turn go here' : '') +
			(highlight ? ', tap to play here' : '')
	);
</script>

<!--
	A mat in the pile's colour, the whole tap target. Top left a solid pill with the direction,
	top right a green pill with the cards played here this turn (the same green as the turn bar's
	dots). The board sizes the card (--w, --mt) so it clears both pills; the pile never resizes.
-->
<button
	type="button"
	{disabled}
	{onclick}
	aria-label={label}
	class={cn(
		'relative grid h-full min-h-0 w-full [touch-action:manipulation] place-items-center rounded-[14px] pt-(--mt) transition-[outline-color,box-shadow] duration-150 select-none [-webkit-tap-highlight-color:transparent]',
		up
			? 'bg-up-soft shadow-[inset_0_0_0_1.5px_var(--up)] [--pile-label:var(--up-label)] [--pile:var(--up)]'
			: 'bg-down-soft shadow-[inset_0_0_0_1.5px_var(--down)] [--pile-label:var(--down-label)] [--pile:var(--down)]',
		bound && 'outline-2 outline-offset-2 outline-go outline-dashed',
		highlight && 'outline-3 outline-offset-2 outline-go outline-solid',
		skull && 'shadow-[inset_0_0_0_2.5px_var(--destructive)]',
		disabled && 'cursor-default'
	)}
>
	<span
		class={cn(
			'absolute top-1.5 left-1.5 flex h-[22px] w-7 items-center justify-center rounded-full',
			up ? 'bg-up-tab text-up-tab-ink' : 'bg-down-tab text-down-tab-ink'
		)}
		aria-hidden="true"
	>
		{#if up}<ArrowUp class="size-4" strokeWidth={2.6} />{:else}<ArrowDown
				class="size-4"
				strokeWidth={2.6}
			/>{/if}
	</span>
	{#if playsThisTurn > 0}
		<span
			class="absolute top-1.5 right-1.5 flex h-[22px] min-w-7 items-center justify-center rounded-full bg-go-soft px-[7px] text-xs font-bold text-go-ink tabular-nums shadow-[inset_0_0_0_1px_var(--go)]"
			aria-hidden="true">+{playsThisTurn}</span
		>
	{/if}
	{#if skull}
		<span
			class="absolute -top-2 left-1/2 z-10 flex -translate-x-1/2 items-center gap-0.5 rounded-full bg-destructive px-1.5 py-0.5 text-xs leading-none font-bold text-white shadow-sm motion-safe:animate-pulse"
			aria-hidden="true"
		>
			<Skull class="size-3" />
			Cover!
		</span>
	{/if}

	<!-- One grid cell, so a top flying back to a hand doesn't push the uncovered one aside. -->
	<span class={cn('grid transition-opacity duration-150', dimmed && 'opacity-35')}>
		{#each tops as t (t.key)}
			<span
				class="relative z-1 [grid-area:1/1]"
				in:receive={{ key: t.key, offstage: t.uncovered ? 'none' : 'top' }}
				out:leave={{ key: t.key, offstage: 'top', taken: pile.cardCount < t.count }}
			>
				<GameCard
					value={t.value}
					command={t.command}
					inactive={!t.active}
					mode={t.start ? 'start' : 'face'}
				/>
			</span>
		{/each}
	</span>
</button>
