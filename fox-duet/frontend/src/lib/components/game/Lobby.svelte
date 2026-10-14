<script lang="ts">
	import HeaderButton, { headerRow } from '@board-games/shared/HeaderButton.svelte';
	import HeaderThemeToggle from '@board-games/shared/HeaderThemeToggle.svelte';
	import { toast } from 'svelte-sonner';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Crown from '@lucide/svelte/icons/crown';
	import DoorOpen from '@lucide/svelte/icons/door-open';
	import Flag from '@lucide/svelte/icons/flag';
	import House from '@lucide/svelte/icons/house';
	import LoaderCircle from '@lucide/svelte/icons/loader-circle';
	import X from '@lucide/svelte/icons/x';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Card from '#lib/components/ui/card/index.js';
	import { errorText } from '#lib/errors.js';
	import { ActionError } from '#lib/game/connection.svelte.js';

	import type { GameClient } from '#lib/game/client.js';
	import { cn } from '#lib/utils.js';
	import type { PlayerView } from '#lib/types.js';
	import ShareInvite from '@board-games/shared/ShareInvite.svelte';

	let {
		client,
		view,
		onleave
	}: { client: GameClient; view: PlayerView; onleave?: () => Promise<void> | void } = $props();

	let busy = $state(false);
	let confirmLeave = $state(false);
	let confirmEnd = $state(false);

	const isHost = $derived(view.hostId === view.you);
	const host = $derived(view.players.find((p) => p.id === view.hostId));
	const connected = $derived(client.status === 'open');
	const players = $derived([...view.players].sort((a, b) => a.seat - b.seat));
	/** The host is still alone in the lobby. */
	const alone = $derived(view.players.length === 1);
	let level = $state(1);

	const LEVELS = [
		{ level: 1, text: 'Side A: 12 gems, 4 forest tokens. Learn the forest.' },
		{ level: 2, text: 'Side B: 13 gems, 3 forest tokens, a shorter path.' },
		{ level: 3, text: 'Side B with diamonds: 16 gems, 3 forest tokens.' }
	];

	async function run(action: () => Promise<void>) {
		if (busy) return;
		busy = true;
		try {
			await action();
		} catch (e) {
			toast.error(errorText(e instanceof ActionError ? e.code : 'internal'));
		} finally {
			busy = false;
		}
	}

	/** Closes the lobby for everyone (host only); it moves to everyone's history. */
	function endGame() {
		return run(async () => {
			await client.endGame();
			confirmEnd = false;
			await goto(resolve('/'));
		});
	}

	async function leave() {
		if (busy) return;
		busy = true;
		try {
			if (onleave) await onleave();
			else await goto(resolve('/'));
		} finally {
			busy = false;
			confirmLeave = false;
		}
	}
</script>

<div class="mx-auto flex w-full max-w-md flex-1 flex-col gap-3">
	<header class={headerRow}>
		<HeaderButton icon={House} label="Home" href={resolve('/')} />
		<h1 class="text-lg font-semibold">Lobby</h1>
		<div class="ml-auto flex items-center gap-1.5">
			{#if alone}
				<!-- A sole player cannot leave; closing affects nobody else, so no confirmation. -->
				<HeaderButton
					icon={X}
					label="Close lobby"
					showLabel
					destructive
					disabled={!connected || busy}
					onclick={endGame}
				/>
			{:else}
				<HeaderButton
					icon={DoorOpen}
					label="Leave"
					showLabel
					destructive
					disabled={busy}
					onclick={() => (confirmLeave = true)}
				/>
				{#if isHost}
					<HeaderButton
						icon={Flag}
						label="End game"
						showLabel
						destructive
						disabled={!connected || busy}
						onclick={() => (confirmEnd = true)}
					/>
				{/if}
			{/if}
			<HeaderThemeToggle />
		</div>
	</header>

	<Card.Root>
		<Card.Content>
			<ShareInvite
				code={view.gameId}
				link={`${globalThis.location?.origin ?? ''}${resolve('/join/[code]', { code: view.gameId })}`}
				title="Fox Duet"
			/>
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Header>
			<Card.Title>Players ({view.players.length}/2)</Card.Title>
		</Card.Header>
		<Card.Content>
			<ul class="flex flex-col gap-2">
				{#each players as p (p.id)}
					{@const isOnline = client.online.includes(p.id)}
					<li class="flex items-center gap-2">
						<span
							class={cn(
								'size-2.5 shrink-0 rounded-full',
								isOnline ? 'bg-emerald-500' : 'bg-muted-foreground/40'
							)}
							title={isOnline ? 'Online' : 'Offline'}
						></span>
						<span class="truncate text-base">{p.name}</span>
						{#if p.isYou}<span class="text-sm text-muted-foreground">(you)</span>{/if}
						{#if p.isHost}
							<Crown class="ml-auto size-4 shrink-0 text-primary" aria-label="Host" />
						{/if}
					</li>
				{/each}
			</ul>
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Header>
			<Card.Title>Difficulty</Card.Title>
		</Card.Header>
		<Card.Content class="flex flex-col gap-2" role="radiogroup" aria-label="Difficulty level">
			{#each LEVELS as l (l.level)}
				<button
					type="button"
					role="radio"
					aria-checked={level === l.level}
					disabled={!isHost}
					class={cn(
						'flex flex-col items-start rounded-lg border px-3 py-2 text-left transition-colors disabled:cursor-default',
						level === l.level ? 'border-fox bg-fox/10' : 'border-border',
						!isHost && level !== l.level && 'opacity-60'
					)}
					onclick={() => (level = l.level)}
				>
					<span class="font-semibold">Level {l.level}</span>
					<span class="text-sm text-muted-foreground">{l.text}</span>
				</button>
			{/each}
			{#if !isHost}
				<p class="text-sm text-muted-foreground">
					{host?.name ?? 'The host'} picks the level when starting.
				</p>
			{/if}
		</Card.Content>
	</Card.Root>

	<div
		class="sticky bottom-0 -mx-3 mt-auto px-3 pt-1 pb-[max(0.5rem,env(safe-area-inset-bottom))] before:absolute before:inset-0 before:-z-10 before:backdrop-blur"
	>
		{#if isHost}
			<Button
				size="lg"
				class="h-12 w-full text-base"
				disabled={!connected || busy || view.players.length < 2}
				onclick={() => run(() => client.startGame(level))}
			>
				{view.players.length < 2 ? 'Waiting for a teammate…' : `Start level ${level}`}
			</Button>
		{:else}
			<p class="flex h-12 items-center justify-center gap-2 text-muted-foreground">
				<LoaderCircle class="size-4 animate-spin" />
				Waiting for {host?.name ?? 'the host'} to start…
			</p>
		{/if}
	</div>
</div>

<AlertDialog.Root bind:open={confirmLeave}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Leave this game?</AlertDialog.Title>
			<AlertDialog.Description>
				You will be removed from the lobby. You can rejoin with the code while it is still open.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={busy}>Stay</AlertDialog.Cancel>
			<AlertDialog.Action
				class="bg-destructive text-white hover:bg-destructive/90"
				disabled={busy}
				onclick={leave}
			>
				Leave
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>

<AlertDialog.Root bind:open={confirmEnd}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>End this game for everyone?</AlertDialog.Title>
			<AlertDialog.Description>
				The lobby closes before the game starts and moves to everyone's history.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={busy}>Keep lobby</AlertDialog.Cancel>
			<AlertDialog.Action
				class="bg-destructive text-white hover:bg-destructive/90"
				disabled={busy || !connected}
				onclick={endGame}
			>
				End game
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
