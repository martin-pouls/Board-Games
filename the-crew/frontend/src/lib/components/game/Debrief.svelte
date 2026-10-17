<script lang="ts">
	import PartyPopper from '@lucide/svelte/icons/party-popper';
	import RotateCcw from '@lucide/svelte/icons/rotate-ccw';
	import House from '@lucide/svelte/icons/house';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import type { GameClient } from '#lib/game/client.js';
	import { cardKey } from '#lib/game/cards.js';
	import { resultText, seatName } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import type { Attempt, PlayerView } from '#lib/types.js';
	import EditionGlyph from './EditionGlyph.svelte';
	import PlayingCard from './PlayingCard.svelte';

	let {
		client,
		view,
		attempt,
		busy,
		connected,
		run
	}: {
		client: GameClient;
		view: PlayerView;
		attempt: Attempt;
		busy: boolean;
		connected: boolean;
		run: (action: () => Promise<void>) => Promise<boolean>;
	} = $props();

	let confirmEnd = $state(false);
	async function endGame() {
		if (await run(() => client.endGame())) confirmEnd = false;
	}

	const result = $derived(attempt.result);
	const won = $derived(!!result?.success);
	const isHost = $derived(view.you === view.hostId);
	const host = $derived(view.players.find((p) => p.id === view.hostId));
	const entry = $derived(view.logbook.find((e) => e.mission === view.mission));
	const disabled = $derived(busy || !connected);
	const lastMission = $derived(view.mission >= 50);
	const reason = $derived(
		result
			? resultText(result, attempt.members, attempt.seat, attempt.tasks, attempt.objectives)
			: ''
	);
	// Seat order, starting with the viewer.
	const members = $derived(
		[...attempt.members].sort(
			(a, b) =>
				((a.seat - attempt.seat + attempt.members.length) % attempt.members.length) -
				((b.seat - attempt.seat + attempt.members.length) % attempt.members.length)
		)
	);
</script>

<section class="flex flex-col gap-2.5 rounded-xl bg-muted/60 px-3 py-2.5" aria-label="Debrief">
	<div class="flex items-start gap-2">
		<span
			class={cn(
				'grid size-9 shrink-0 place-items-center rounded-full text-white',
				won ? 'bg-success' : 'bg-danger'
			)}
		>
			{#if won}<PartyPopper class="size-5" />{:else}<RotateCcw class="size-5" />{/if}
		</span>
		<div class="min-w-0">
			<h2 class="font-display text-lg leading-tight font-bold">
				{won ? `Mission ${view.mission} accomplished` : `Attempt ${attempt.number} failed`}
			</h2>
			<p class="text-sm leading-snug">{reason}</p>
			{#if entry}
				<p class="text-xs text-muted-foreground">
					Logbook: mission {entry.mission}, {entry.attempts}
					{entry.attempts === 1 ? 'attempt' : 'attempts'}{entry.distress
						? ' + 1 for the distress signal'
						: ''}{entry.completed ? ' · completed' : ''}
				</p>
			{/if}
		</div>
	</div>

	<ul class="flex flex-col gap-1">
		{#each members as m (m.seat)}
			{#if m.cards?.length}
				<li class="flex items-center gap-2">
					<span class="w-14 shrink-0 truncate text-xs text-muted-foreground">
						{seatName(attempt.members, m.seat, attempt.seat)}
					</span>
					<span class="flex flex-wrap gap-0.5">
						{#each m.cards as c (cardKey(c))}
							<PlayingCard card={c} size="xs" />
						{/each}
					</span>
				</li>
			{/if}
		{/each}
	</ul>

	{#if isHost}
		<div class="grid grid-cols-2 gap-2">
			{#if won}
				<Button
					class="h-11"
					disabled={disabled || lastMission}
					onclick={() => run(() => client.nextMission())}
				>
					<EditionGlyph kind="mission" />{lastMission
						? 'Last mission'
						: `Mission ${view.mission + 1}`}
				</Button>
			{:else}
				<Button class="h-11" {disabled} onclick={() => run(() => client.nextAttempt())}>
					<RotateCcw />Try again
				</Button>
			{/if}
			<Button
				class="h-11"
				variant="secondary"
				{disabled}
				onclick={() => run(() => client.backToLobby())}
			>
				<House />Lobby
			</Button>
		</div>
		<button
			type="button"
			class="self-center text-xs text-muted-foreground underline-offset-2 hover:text-destructive hover:underline disabled:opacity-50"
			{disabled}
			onclick={() => (confirmEnd = true)}
		>
			End the game for everyone
		</button>
	{:else}
		<p class="text-center text-sm text-muted-foreground">
			Waiting for {host?.name ?? 'the host'} to continue.
		</p>
	{/if}
</section>

<AlertDialog.Root bind:open={confirmEnd}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>End the game?</AlertDialog.Title>
			<AlertDialog.Description>
				The session closes for everyone and can no longer be played. Its logbook stays viewable
				until the game is deleted after 30 days.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={busy}>Keep playing</AlertDialog.Cancel>
			<AlertDialog.Action
				class="bg-destructive text-white hover:bg-destructive/90"
				{disabled}
				onclick={endGame}
			>
				End game
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
