<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { fly } from 'svelte/transition';
	import { prefersReducedMotion } from 'svelte/motion';
	import { toast } from 'svelte-sonner';
	import ChevronUp from '@lucide/svelte/icons/chevron-up';
	import Eye from '@lucide/svelte/icons/eye';
	import Gem from '@lucide/svelte/icons/gem';
	import House from '@lucide/svelte/icons/house';
	import RotateCcw from '@lucide/svelte/icons/rotate-ccw';
	import Trees from '@lucide/svelte/icons/trees';
	import { Button } from '#lib/components/ui/button/index.js';
	import { errorText } from '#lib/errors.js';
	import { ActionError } from '#lib/game/connection.svelte.js';
	import type { GameClient } from '#lib/game/client.js';
	import { lossText, resultHeadline } from '#lib/game/format.js';
	import { sessions } from '#lib/session.js';
	import type { PlayerView } from '#lib/types.js';
	import PlayingCard from './PlayingCard.svelte';
	import Table from './Table.svelte';

	let { client, view }: { client: GameClient; view: PlayerView } = $props();

	let hidden = $state(false);
	let busy = $state(false);
	/** Set after "Play again" so we follow the new game as soon as its id arrives. */
	let follow = $state(false);

	const result = $derived(view.result);
	const won = $derived(result?.outcome === 'won');
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

<Table {client} {view} readonly>
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
</Table>

{#if result && !hidden}
	<div
		class="fixed inset-0 z-40 flex items-end justify-center overflow-y-auto bg-background/70 p-3 pb-[max(0.75rem,env(safe-area-inset-bottom))] backdrop-blur-sm sm:items-center"
		role="dialog"
		aria-modal="true"
		aria-labelledby="game-over-title"
	>
		<div
			in:fly={{ y: 24, duration: prefersReducedMotion.current ? 0 : 250 }}
			class="flex max-h-full w-full max-w-md flex-col gap-4 overflow-y-auto rounded-2xl border bg-card p-5 text-card-foreground shadow-xl"
		>
			<div class="flex flex-col items-center gap-1 text-center">
				{#if won}
					<Gem class="size-10 text-gem" aria-hidden="true" />
					<h2 id="game-over-title" class="text-2xl font-bold">Every gem found!</h2>
					<p class="text-muted-foreground">Level {view.level}, won in round {result.round}.</p>
				{:else}
					<Trees class="size-10 text-forest" aria-hidden="true" />
					<h2 id="game-over-title" class="text-2xl font-bold">Defeat</h2>
					<p class="text-muted-foreground">{lossText(view)}</p>
				{/if}
			</div>

			{#if result.score}
				{@const s = result.score}
				<dl
					class="grid grid-cols-[1fr_auto] gap-x-4 gap-y-1 rounded-lg bg-muted/60 p-3 tabular-nums"
				>
					<dt>Level {view.level}</dt>
					<dd class="text-right">{s.level}</dd>
					<dt>Cards left in hand</dt>
					<dd class="text-right">+{s.cards}</dd>
					{#if s.earlyWin}
						<dt>Won before round 3</dt>
						<dd class="text-right">+{s.earlyWin}</dd>
					{/if}
					<dt>Forest tokens left</dt>
					<dd class="text-right">+{s.forestTokens}</dd>
					<dt class="border-t pt-1 font-semibold">Score</dt>
					<dd class="border-t pt-1 text-right text-xl font-bold text-fox">{s.total}</dd>
				</dl>
			{:else}
				<dl class="grid grid-cols-2 gap-2 text-center">
					<div class="rounded-lg bg-muted/60 p-2">
						<dt class="text-xs text-muted-foreground">Round</dt>
						<dd class="text-xl font-semibold tabular-nums">{result.round}/3</dd>
					</div>
					<div class="rounded-lg bg-muted/60 p-2">
						<dt class="text-xs text-muted-foreground">Gems left</dt>
						<dd class="text-xl font-semibold tabular-nums">{result.gemsLeft}</dd>
					</div>
				</dl>
			{/if}

			<ul class="flex flex-col gap-2" aria-label="Cards left in each hand">
				{#each players as p (p.id)}
					<li class="flex flex-col gap-1">
						<span class="truncate text-sm font-medium">{p.isYou ? 'You' : p.name}</span>
						<span class="flex flex-wrap gap-1">
							{#each p.revealedHand ?? [] as c, i (i)}
								<PlayingCard card={c} size="xs" />
							{:else}
								<span class="text-sm text-muted-foreground">No cards left</span>
							{/each}
						</span>
					</li>
				{/each}
			</ul>

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
						View table
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
