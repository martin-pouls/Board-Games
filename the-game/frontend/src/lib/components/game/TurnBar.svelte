<script lang="ts">
	import { fly } from 'svelte/transition';
	import { prefersReducedMotion } from 'svelte/motion';
	import Undo2 from '@lucide/svelte/icons/undo-2';
	import { Button } from '#lib/components/ui/button/index.js';
	import { commandInfo, targetCount, turnHints } from '#lib/game/commands.js';
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
		/** Shown instead of End turn while the player is stuck. */
		ongiveup?: () => void;
	} = $props();

	const turn = $derived(view.turn);
	const active = $derived(view.players.find((p) => p.id === turn?.activePlayerId));
	const target = $derived(turn ? targetCount(turn) : 0);
	const exactly = $derived(turn?.constraints.exactCount != null);
	const dots = $derived(
		turn ? Array.from({ length: Math.max(target, turn.cardsPlayed) }, (_, i) => i) : []
	);
	const hints = $derived(turnHints(view, isMyTurn));
</script>

{#if turn}
	{#key turn.activePlayerId + ':' + turn.number}
		<div
			in:fly={{ y: 8, duration: prefersReducedMotion.current ? 0 : 200 }}
			class={cn(
				'flex flex-col gap-1 rounded-xl px-3 py-1.5',
				isMyTurn ? 'bg-emerald-500/10 ring-1 ring-emerald-500/40' : 'bg-muted/60'
			)}
			aria-live="polite"
		>
			<div class="flex min-h-11 items-center gap-2">
				<div class="min-w-0 flex-1">
					{#if isMyTurn}
						<p class="font-semibold">Your turn</p>
						<p class="flex items-center gap-1.5 text-sm text-muted-foreground">
							<span class="flex gap-0.5" aria-hidden="true">
								{#each dots as i (i)}
									<span
										class={cn(
											'size-2 rounded-full',
											i < turn.cardsPlayed ? 'bg-emerald-500' : 'bg-muted-foreground/30'
										)}
									></span>
								{/each}
							</span>
							<span class="tabular-nums"
								>{turn.cardsPlayed}/{target} played{#if exactly}&nbsp;· exactly{/if}</span
							>
						</p>
						{#if turn.stuck}
							<p class="text-sm font-medium text-destructive">Stuck: undo a card or give up</p>
						{/if}
					{:else}
						<p class="truncate font-semibold">{active?.name ?? 'Someone'}'s turn</p>
						<p class="text-sm text-muted-foreground tabular-nums">{turn.cardsPlayed} played</p>
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
							size="lg"
							class="h-12 min-w-28 bg-destructive px-5 text-base text-white hover:bg-destructive/90"
							disabled={disabled || pending}
							onclick={ongiveup}
						>
							Give up
						</Button>
					{:else}
						<Button
							size="lg"
							class="h-12 min-w-28 px-5 text-base"
							disabled={disabled || pending || !turn.canEndTurn}
							onclick={onendturn}
						>
							End turn
						</Button>
					{/if}
				{/if}
			</div>
			{#if hints.length > 0}
				<ul class="flex flex-wrap gap-1 pb-0.5" aria-label="Commands this turn">
					{#each hints as h (h.command)}
						<li
							class={cn(
								'flex max-w-full min-w-0 items-center gap-1 rounded-full px-2 py-0.5 text-xs leading-tight font-medium',
								commandInfo(h.command).kind === 'lightning'
									? 'bg-red-500/15 text-red-800 dark:text-red-200'
									: 'bg-violet-500/15 text-violet-800 dark:text-violet-200'
							)}
						>
							<CommandIcon command={h.command} class="size-3.5" />
							<span class="min-w-0">{h.text}</span>
						</li>
					{/each}
				</ul>
			{/if}
		</div>
	{/key}
{/if}
