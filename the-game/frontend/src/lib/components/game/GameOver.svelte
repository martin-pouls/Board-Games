<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { fly } from 'svelte/transition';
	import { prefersReducedMotion } from 'svelte/motion';
	import { toast } from 'svelte-sonner';
	import Trophy from '@lucide/svelte/icons/trophy';
	import Eye from '@lucide/svelte/icons/eye';
	import RotateCcw from '@lucide/svelte/icons/rotate-ccw';
	import House from '@lucide/svelte/icons/house';
	import ChevronUp from '@lucide/svelte/icons/chevron-up';
	import { Button } from '#lib/components/ui/button/index.js';
	import { errorText } from '#lib/errors.js';
	import { ActionError } from '#lib/game/connection.svelte.js';
	import type { GameClient } from '#lib/game/client.js';
	import { endedLine, resultHeadline, stuckLine } from '#lib/game/format.js';
	import { sessions } from '#lib/session.js';
	import { cn } from '#lib/utils.js';
	import type { PlayerView } from '#lib/types.js';
	import { commandInfo } from '#lib/game/commands.js';
	import Board from './Board.svelte';
	import CommandIcon from './CommandIcon.svelte';

	let { client, view }: { client: GameClient; view: PlayerView } = $props();

	let hidden = $state(false);
	let busy = $state(false);
	/** Set after "Play again" so we follow the new game as soon as its id arrives. */
	let follow = $state(false);

	const result = $derived(view.result);
	const won = $derived(result?.outcome === 'won');
	const stuck = $derived(stuckLine(view) ?? endedLine(view));
	const players = $derived([...view.players].sort((a, b) => a.seat - b.seat));
	const connected = $derived(client.status === 'open');

	async function joinNext() {
		const next = view.nextGameId;
		if (!next) return;
		sessions.copy(view.gameId, next);
		await goto(resolve('/g/[code]', { code: next }));
	}

	async function playAgain() {
		if (busy) return;
		busy = true;
		try {
			await client.rematch();
			follow = true;
		} catch (e) {
			toast.error(errorText(e instanceof ActionError ? e.code : 'internal'));
			busy = false;
		}
	}

	// The broadcast view (carrying nextGameId) normally arrives before the ack.
	$effect(() => {
		if (follow && view.nextGameId) {
			follow = false;
			joinNext();
		}
	});
</script>

<Board {client} {view} readonly>
	{#snippet controls()}
		<div class="flex min-h-14 items-center gap-3 rounded-xl bg-muted/60 px-3 py-1.5">
			<p class="min-w-0 flex-1 leading-snug font-semibold">
				{result ? resultHeadline(result) : 'Game over'}
			</p>
			<Button size="lg" class="h-12 px-4 text-base" onclick={() => (hidden = false)}>
				<ChevronUp />
				Results
			</Button>
		</div>
	{/snippet}
</Board>

{#if result}
	{#if !hidden}
		<div
			class="fixed inset-0 z-40 flex items-end justify-center overflow-y-auto bg-background/70 p-3 pb-[max(0.75rem,env(safe-area-inset-bottom))] backdrop-blur-sm sm:items-center"
			role="dialog"
			aria-modal="true"
			aria-labelledby="game-over-title"
		>
			<div
				in:fly={{ y: 24, duration: prefersReducedMotion.current ? 0 : 250 }}
				class="flex w-full max-w-md flex-col gap-4 rounded-2xl border bg-card p-5 text-card-foreground shadow-xl"
			>
				<div class="flex flex-col items-center gap-1 text-center">
					{#if won || (result.excellent && result.outcome !== 'aborted')}
						<Trophy class={cn('size-10', won ? 'text-amber-500' : 'text-sky-500')} />
					{/if}
					<h2 id="game-over-title" class="text-2xl font-bold">{resultHeadline(result)}</h2>
					{#if stuck}<p class="text-muted-foreground">{stuck}</p>{/if}
				</div>

				<!-- Lost with score 0 (uncovered skull as the last card): nothing left to count. -->
				{#if !won && result.score > 0}
					<dl class="grid grid-cols-2 gap-2 text-center">
						<div class="rounded-lg bg-muted/60 p-2">
							<dt class="text-xs text-muted-foreground">In hands</dt>
							<dd class="text-xl font-semibold tabular-nums">{result.cardsLeftInHands}</dd>
						</div>
						<div class="rounded-lg bg-muted/60 p-2">
							<dt class="text-xs text-muted-foreground">Draw pile</dt>
							<dd class="text-xl font-semibold tabular-nums">{result.cardsLeftInDrawPile}</dd>
						</div>
					</dl>

					<ul class="flex flex-col gap-2" aria-label="Cards left in each hand">
						{#each players as p (p.id)}
							<li class="flex flex-col gap-1">
								<span class="truncate text-sm font-medium">
									{p.isYou ? 'You' : p.name}
								</span>
								<span class="flex flex-wrap gap-1">
									{#each p.revealedHand ?? [] as c (c.value)}
										<span
											class={cn(
												'flex h-8 w-7 flex-col items-center justify-center rounded border bg-background text-xs leading-none font-bold tabular-nums',
												c.command &&
													(commandInfo(c.command).kind === 'lightning'
														? 'border-red-400 dark:border-red-500'
														: 'border-violet-400 dark:border-violet-500')
											)}
										>
											{c.value}
											{#if c.command}
												<CommandIcon command={c.command} class="size-2.5 opacity-80" />
												<span class="sr-only">, {commandInfo(c.command).label}</span>
											{/if}
										</span>
									{:else}
										<span class="text-sm text-muted-foreground">No cards left</span>
									{/each}
								</span>
							</li>
						{/each}
					</ul>
				{/if}

				<div class="flex flex-col gap-2">
					{#if view.nextGameId}
						<Button size="lg" class="h-12 text-base" onclick={joinNext}>
							<RotateCcw />
							Join next game
						</Button>
					{:else}
						<Button
							size="lg"
							class="h-12 text-base"
							disabled={busy || !connected}
							onclick={playAgain}
						>
							<RotateCcw />
							Play again
						</Button>
					{/if}
					<div class="grid grid-cols-2 gap-2">
						<Button variant="outline" class="h-11" onclick={() => (hidden = true)}>
							<Eye />
							View board
						</Button>
						<Button variant="outline" class="h-11" href={resolve('/')}>
							<House />
							Home
						</Button>
					</div>
				</div>
			</div>
		</div>
	{/if}
{/if}
