<script lang="ts">
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

	const headerAction =
		'flex items-center gap-1 rounded-md p-1 text-sm text-muted-foreground hover:text-destructive disabled:opacity-50';

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
	<header class="flex items-center gap-2">
		<a
			href={resolve('/')}
			class="-ml-1 rounded-md p-1 text-muted-foreground hover:text-foreground"
			aria-label="Home"
		>
			<House class="size-5" />
		</a>
		<h1 class="text-lg font-semibold">Lobby</h1>
		<div class="-mr-1 ml-auto flex items-center gap-1">
			{#if alone}
				<!-- A sole player cannot leave; closing affects nobody else, so no confirmation. -->
				<button
					type="button"
					class={headerAction}
					disabled={!connected || busy}
					title="Close lobby"
					onclick={endGame}
				>
					<X class="size-4" />
					Close lobby
				</button>
			{:else}
				<button
					type="button"
					class={headerAction}
					disabled={busy}
					title="Leave"
					onclick={() => (confirmLeave = true)}
				>
					<DoorOpen class="size-4" />
					Leave
				</button>
				{#if isHost}
					<button
						type="button"
						class={headerAction}
						disabled={!connected || busy}
						title="End game"
						onclick={() => (confirmEnd = true)}
					>
						<Flag class="size-4" />
						End game
					</button>
				{/if}
			{/if}
		</div>
	</header>

	<Card.Root>
		<Card.Content>
			<ShareInvite
				code={view.gameId}
				link={`${globalThis.location?.origin ?? ''}${resolve('/join/[code]', { code: view.gameId })}`}
				title="The Crew"
			/>
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Header>
			<Card.Title>Players ({view.players.length}/5)</Card.Title>
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
			{#if view.players.length === 2}
				<p class="mt-3 text-sm text-muted-foreground">JARVIS joins as the third crew member</p>
			{/if}
		</Card.Content>
	</Card.Root>

	<div
		class="sticky bottom-0 -mx-3 mt-auto bg-background/95 px-3 pt-1 pb-[max(0.5rem,env(safe-area-inset-bottom))] backdrop-blur"
	>
		{#if isHost}
			<!-- Placeholder: starting a mission arrives with the game flow (implementation plan WP5/WP9). -->
			<Button size="lg" class="h-12 w-full text-base" disabled>
				{view.players.length < 2 ? 'Waiting for a crew member…' : 'Start mission'}
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
