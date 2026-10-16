<script lang="ts">
	import { fly } from 'svelte/transition';
	import { prefersReducedMotion } from 'svelte/motion';
	import Flag from '@lucide/svelte/icons/flag';
	import Undo2 from '@lucide/svelte/icons/undo-2';
	import Zap from '@lucide/svelte/icons/zap';
	import { Button } from '#lib/components/ui/button/index.js';
	import { targetCount, turnStatus } from '#lib/game/commands.js';
	import { cn } from '#lib/utils.js';
	import type { PlayerView } from '#lib/types.js';
	import CommandIcon from './CommandIcon.svelte';

	let {
		view,
		isMyTurn,
		disabled = false,
		pending = false,
		onundo,
		onendturn,
		ongiveup
	}: {
		view: PlayerView;
		isMyTurn: boolean;
		disabled?: boolean;
		pending?: boolean;
		onundo?: () => void;
		onendturn?: () => void;
		/** Shown instead of End turn while the player is stuck (opens the give-up confirmation). */
		ongiveup?: () => void;
	} = $props();

	const turn = $derived(view.turn);
	const active = $derived(view.players.find((p) => p.id === turn?.activePlayerId));
	const status = $derived(turnStatus(view, isMyTurn));
	const dots = $derived(
		turn ? Array.from({ length: Math.max(targetCount(turn), turn.cardsPlayed) }, (_, i) => i) : []
	);
</script>

<!--
	No "Your turn" title: the green bar and the player strip already say whose turn it is. Next to
	the buttons: one status line (count, exactly 3, STOP or stuck) and badges for 3! and Single pile while
	in force. Nothing wraps, so the panel keeps its height and the piles stay put.
-->
{#if turn && status}
	{#key turn.activePlayerId + ':' + turn.number}
		<div
			in:fly={{ y: 8, duration: prefersReducedMotion.current ? 0 : 200 }}
			class={cn(
				'flex min-h-11 items-center gap-2 rounded-xl px-3 py-1.5',
				isMyTurn ? 'bg-emerald-500/10 ring-1 ring-emerald-500/40' : 'bg-muted/60'
			)}
			aria-live="polite"
		>
			<div class="flex min-w-0 flex-1 flex-col gap-1">
				<p
					class={cn(
						'flex items-center gap-1.5 text-sm font-medium whitespace-nowrap',
						status.line.tone === 'warn' && 'text-destructive'
					)}
				>
					<span class="sr-only"
						>{isMyTurn ? 'Your turn' : `${active?.name ?? 'Someone'}'s turn`}:</span
					>
					{#if status.line.dots}
						<span class="flex shrink-0 gap-0.5" aria-hidden="true">
							{#each dots as i (i)}
								<span
									class={cn(
										'size-2 rounded-full',
										i < turn.cardsPlayed ? 'bg-emerald-500' : 'bg-muted-foreground/30'
									)}
								></span>
							{/each}
						</span>
					{/if}
					<span class="truncate tabular-nums">{status.line.text}</span>
				</p>
				{#if status.singlePile || status.exactly !== null}
					<div class="flex min-w-0 items-center gap-1">
						{#if status.exactly !== null}
							<!-- icon only, so it fits beside the Single pile tag; the line counts toward it -->
							<p
								class="flex shrink-0 items-center gap-0.5 rounded-full bg-amber-500/20 px-1.5 py-0.5 text-amber-800 dark:text-amber-200"
								title={`3!: exactly ${status.exactly} cards this turn`}
							>
								<Zap class="size-3.5" aria-hidden="true" />
								<CommandIcon command="exactlyThree" class="size-3.5" />
								<span class="sr-only">Exactly {status.exactly} cards this turn</span>
							</p>
						{/if}
						{#if status.singlePile}
							<p
								class={cn(
									'flex min-w-0 items-center gap-1 rounded-full px-1.5 py-0.5 text-xs leading-none font-semibold whitespace-nowrap',
									status.singlePile.bound
										? 'bg-emerald-500/15 text-emerald-800 dark:text-emerald-200'
										: 'bg-violet-500/15 text-violet-800 dark:text-violet-200'
								)}
								title={status.singlePile.title}
							>
								<CommandIcon command="singlePile" class="size-3.5" />
								<!-- always short, so it fits beside the 3! badge on any phone; the colour says bound or not -->
								<span class="sr-only">{status.singlePile.label}</span>
								<span aria-hidden="true">One pile</span>
							</p>
						{/if}
					</div>
				{/if}
			</div>
			{#if isMyTurn}
				<Button
					variant="outline"
					size="lg"
					class="h-12 px-3 text-base"
					disabled={disabled || pending || turn.cardsPlayed === 0}
					title="Take back the last card"
					onclick={onundo}
				>
					<Undo2 class="size-5" aria-hidden="true" />
					<span class="max-sm:sr-only">Undo</span>
				</Button>
				{#if turn.stuck}
					<Button
						variant="outline"
						size="lg"
						class="h-12 min-w-24 px-4 text-base"
						disabled={disabled || pending}
						onclick={ongiveup}
					>
						<Flag class="size-5 max-[24rem]:hidden" aria-hidden="true" />
						Give up
					</Button>
				{:else}
					<Button
						size="lg"
						class="h-12 min-w-24 px-4 text-base"
						disabled={disabled || pending || !turn.canEndTurn}
						onclick={onendturn}
					>
						{#if status.drawOnlyOne}
							<span class="flex flex-col items-center leading-tight">
								End turn
								<span class="flex items-center gap-0.5 text-[0.7rem] font-medium opacity-85">
									<CommandIcon command="drawOnlyOne" class="size-3" />
									draw only 1
								</span>
							</span>
						{:else}
							End turn
						{/if}
					</Button>
				{/if}
			{/if}
		</div>
	{/key}
{/if}
