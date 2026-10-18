<script lang="ts">
	import { fly } from 'svelte/transition';
	import { prefersReducedMotion } from 'svelte/motion';
	import ArrowUp from '@lucide/svelte/icons/arrow-up';
	import ArrowDown from '@lucide/svelte/icons/arrow-down';
	import Flag from '@lucide/svelte/icons/flag';
	import Undo2 from '@lucide/svelte/icons/undo-2';
	import { Button } from '#lib/components/ui/button/index.js';
	import { commandInfo, targetCount, turnStatus } from '#lib/game/commands.js';
	import { cn } from '#lib/utils.js';
	import type { PlayerView } from '#lib/types.js';
	import CommandBadge from './CommandBadge.svelte';
	import CommandIcon from './CommandIcon.svelte';
	import GameCard from './GameCard.svelte';

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

	/** Other players' plays shown in the bar; older ones collapse into "+N". */
	const MAX_PLAYS = 4;

	const turn = $derived(view.turn);
	const active = $derived(view.players.find((p) => p.id === turn?.activePlayerId));
	const name = $derived(active?.name ?? 'Someone');
	const status = $derived(turnStatus(view, isMyTurn));
	const dots = $derived(
		turn ? Array.from({ length: Math.max(targetCount(turn), turn.cardsPlayed) }, (_, i) => i) : []
	);
	const plays = $derived(turn?.plays.slice(-MAX_PLAYS) ?? []);
	const hiddenPlays = $derived((turn?.plays.length ?? 0) - plays.length);
	const direction = (pile: number) => view.piles.find((p) => p.index === pile)?.direction;
</script>

<!--
	Two fixed rows, so the piles never move. The bar: whose turn, then your count (dots) or the
	cards the other player put down, and the rules in force as their printed badges (each rule is
	explained in the rules dialog). Below it the turn's buttons, or who we're waiting for.
-->
{#if turn && status}
	<div class="flex shrink-0 flex-col gap-2">
		{#key turn.activePlayerId + ':' + turn.number}
			<div
				in:fly={{ y: 8, duration: prefersReducedMotion.current ? 0 : 200 }}
				class={cn(
					'flex h-12 items-center gap-2.5 overflow-hidden rounded-xl pr-2.5 pl-3 whitespace-nowrap',
					isMyTurn ? 'bg-go-soft shadow-[inset_0_0_0_1.5px_var(--go)]' : 'bg-muted'
				)}
				aria-live="polite"
			>
				<b class="max-w-32 shrink-0 truncate font-bold">
					{isMyTurn ? 'Your turn' : `${name}'s turn`}
				</b>
				{#if isMyTurn || status.line.tone === 'warn'}
					<p
						class={cn(
							'flex min-w-0 items-center gap-1.5 text-sm tabular-nums',
							status.line.tone === 'warn' ? 'font-medium text-destructive' : 'text-muted-foreground'
						)}
					>
						{#if status.line.dots}
							<span class="flex shrink-0 gap-[3px]" aria-hidden="true">
								{#each dots as i (i)}
									<span
										class={cn(
											'size-2 rounded-full',
											i < turn.cardsPlayed ? 'bg-go' : 'bg-muted-foreground/30'
										)}
									></span>
								{/each}
							</span>
						{/if}
						<span class="truncate">{status.line.text}</span>
					</p>
				{:else}
					<ul class="flex min-w-0 items-center gap-1.5" aria-label={`${name}'s cards this turn`}>
						{#if hiddenPlays > 0}
							<li class="text-sm text-muted-foreground tabular-nums">+{hiddenPlays}</li>
						{/if}
						{#each plays as play, i (i)}
							{@const dir = direction(play.pile)}
							<li class="relative mr-1.5 shrink-0">
								<GameCard value={play.card.value} command={play.card.command} mode="mini" />
								<span
									class={cn(
										'absolute -right-1.5 -bottom-1 grid size-[15px] place-items-center rounded-full bg-muted',
										dir === 'up' ? 'text-up-label' : 'text-down-label'
									)}
									aria-hidden="true"
								>
									{#if dir === 'up'}<ArrowUp class="size-3" strokeWidth={2.6} />{:else}<ArrowDown
											class="size-3"
											strokeWidth={2.6}
										/>{/if}
								</span>
								<span class="sr-only"
									>{play.card.value}{play.card.command
										? ` ${commandInfo(play.card.command).label}`
										: ''} on an {dir === 'up' ? 'ascending' : 'descending'} pile</span
								>
							</li>
						{/each}
					</ul>
				{/if}
				{#if status.rules.length > 0}
					<ul class="ml-auto flex shrink-0 items-center gap-1" aria-label="Rules in force">
						{#each status.rules as rule (rule)}
							{@const info = commandInfo(rule)}
							<li
								class="flex"
								title={rule === 'singlePile' && status.singlePile
									? status.singlePile.title
									: `${info.label}: ${info.description}`}
							>
								<CommandBadge command={rule} />
								<span class="sr-only">{info.label}</span>
							</li>
						{/each}
					</ul>
				{/if}
			</div>
		{/key}

		<div class="flex h-12 items-stretch gap-2">
			{#if isMyTurn}
				<Button
					variant="outline"
					size="lg"
					class="h-12 w-24 text-base"
					disabled={disabled || pending || turn.cardsPlayed === 0}
					title="Take back the last card"
					onclick={onundo}
				>
					<Undo2 class="size-5" aria-hidden="true" />
					Undo
				</Button>
				{#if turn.stuck}
					<Button
						variant="outline"
						size="lg"
						class="h-12 flex-1 text-base"
						disabled={disabled || pending}
						onclick={ongiveup}
					>
						<Flag class="size-5" aria-hidden="true" />
						Give up
					</Button>
				{:else}
					<Button
						size="lg"
						class="h-12 flex-1 text-base"
						disabled={disabled || pending || !turn.canEndTurn}
						onclick={onendturn}
					>
						{#if status.drawOnlyOne}
							<span class="flex flex-col items-center leading-tight">
								End turn
								<span class="flex items-center gap-0.5 text-xs font-medium opacity-85">
									<CommandIcon command="drawOnlyOne" class="size-3" />
									draw only 1
								</span>
							</span>
						{:else}
							End turn
						{/if}
					</Button>
				{/if}
			{:else}
				<p class="flex flex-1 items-center justify-center text-sm text-muted-foreground">
					<span class="truncate">Waiting for {name} to play.</span>
				</p>
			{/if}
		</div>
	</div>
{/if}
