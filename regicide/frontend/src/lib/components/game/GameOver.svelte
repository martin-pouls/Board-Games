<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { fly } from 'svelte/transition';
	import { prefersReducedMotion } from 'svelte/motion';
	import { toast } from 'svelte-sonner';
	import ChevronUp from '@lucide/svelte/icons/chevron-up';
	import Eye from '@lucide/svelte/icons/eye';
	import House from '@lucide/svelte/icons/house';
	import Medal from '@lucide/svelte/icons/medal';
	import RotateCcw from '@lucide/svelte/icons/rotate-ccw';
	import Skull from '@lucide/svelte/icons/skull';
	import { Button } from '#lib/components/ui/button/index.js';
	import { errorText } from '#lib/errors.js';
	import { ActionError } from '#lib/game/connection.svelte.js';
	import type { GameClient } from '#lib/game/client.js';
	import { gradeLabel, lossText, resultHeadline } from '#lib/game/format.js';
	import { sessions } from '#lib/session.js';
	import { cn } from '#lib/utils.js';
	import type { Grade, PlayerView } from '#lib/types.js';
	import PlayingCard from './PlayingCard.svelte';
	import Table from './Table.svelte';

	let { client, view }: { client: GameClient; view: PlayerView } = $props();

	let hidden = $state(false);
	let busy = $state(false);
	/** Set after "Play again" so we follow the new game as soon as its id arrives. */
	let follow = $state(false);

	const result = $derived(view.result);
	const won = $derived(result?.outcome === 'won');
	const solo = $derived(view.players.length === 1);
	const players = $derived([...view.players].sort((a, b) => a.seat - b.seat));
	const connected = $derived(client.status === 'open');

	const MEDAL_COLOR: Record<Grade, string> = {
		gold: 'text-gold',
		silver: 'text-slate-500 dark:text-slate-300',
		bronze: 'text-amber-700'
	};
	const GRADE_NOTE: Record<Grade, string> = {
		gold: 'no Jester used',
		silver: 'one Jester used',
		bronze: 'two Jesters used'
	};

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
					{#if result.grade}
						<Medal class={cn('size-10', MEDAL_COLOR[result.grade])} aria-hidden="true" />
					{/if}
					<h2 id="game-over-title" class="text-2xl font-bold">The castle has fallen!</h2>
					{#if result.grade}
						<p class="text-muted-foreground">
							{gradeLabel(result.grade)}: {GRADE_NOTE[result.grade]}
						</p>
					{/if}
				{:else}
					<Skull class="size-10 text-health" aria-hidden="true" />
					<h2 id="game-over-title" class="text-2xl font-bold">Defeat</h2>
					<p class="text-muted-foreground">{lossText(view)}</p>
				{/if}
			</div>

			<dl class="grid grid-cols-2 gap-2 text-center">
				<div class="rounded-lg bg-muted/60 p-2">
					<dt class="text-xs text-muted-foreground">Enemies defeated</dt>
					<dd class="text-xl font-semibold tabular-nums">{result.enemiesDefeated}/12</dd>
				</div>
				{#if solo}
					<div class="rounded-lg bg-muted/60 p-2">
						<dt class="text-xs text-muted-foreground">Jesters used</dt>
						<dd class="text-xl font-semibold tabular-nums">
							{result.soloJestersUsed}/{view.rules.soloJesters}
						</dd>
					</div>
				{:else}
					<div class="rounded-lg bg-muted/60 p-2">
						<dt class="text-xs text-muted-foreground">Players</dt>
						<dd class="text-xl font-semibold tabular-nums">{players.length}</dd>
					</div>
				{/if}
			</dl>

			<ul class="flex flex-col gap-2" aria-label="Cards left in each hand">
				{#each players as p (p.id)}
					<li class="flex flex-col gap-1">
						<span class="truncate text-sm font-medium">{p.isYou ? 'You' : p.name}</span>
						<span class="flex flex-wrap gap-1">
							{#each p.revealedHand ?? [] as c, i (i)}
								<PlayingCard card={c} size="xs" />
							{:else}
								<span class="text-sm text-muted-foreground"> No cards left </span>
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
