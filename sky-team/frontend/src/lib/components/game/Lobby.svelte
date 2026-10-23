<script lang="ts">
	import HeaderButton, { headerRow } from '@board-games/shared/HeaderButton.svelte';
	import HeaderThemeToggle from '@board-games/shared/HeaderThemeToggle.svelte';
	import LobbyActions from '@board-games/shared/LobbyActions.svelte';
	import ShareInvite from '@board-games/shared/ShareInvite.svelte';
	import { toast } from 'svelte-sonner';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Crown from '@lucide/svelte/icons/crown';
	import House from '@lucide/svelte/icons/house';
	import * as Card from '#lib/components/ui/card/index.js';
	import { errorText } from '#lib/errors.js';
	import { ActionError } from '#lib/game/connection.svelte.js';
	import type { GameClient } from '#lib/game/client.js';
	import { cn } from '#lib/utils.js';
	import type { PlayerView } from '#lib/types.js';

	let {
		client,
		view,
		onleave
	}: { client: GameClient; view: PlayerView; onleave?: () => Promise<void> | void } = $props();

	let busy = $state(false);

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

	/** Closes the lobby for everyone (host only); it moves to everyone's history. */
	function endGame() {
		return run(async () => {
			await client.endGame();
			await goto(resolve('/'));
		});
	}

	async function leave() {
		if (onleave) await onleave();
		else await goto(resolve('/'));
	}
</script>

<div class="mx-auto flex w-full max-w-md flex-1 flex-col gap-3">
	<header class={headerRow}>
		<HeaderButton icon={House} label="Home" href={resolve('/')} />
		<h1 class="text-lg font-semibold">Lobby</h1>
		<div class="ml-auto flex items-center gap-1.5">
			<LobbyActions {alone} {isHost} {connected} {busy} onend={endGame} onleave={leave} />
			<HeaderThemeToggle />
		</div>
	</header>

	<Card.Root>
		<Card.Content>
			<ShareInvite
				code={view.gameId}
				link={`${globalThis.location?.origin ?? ''}${resolve('/join/[code]', { code: view.gameId })}`}
				title="Sky Team"
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
								isOnline ? 'bg-online' : 'bg-muted-foreground/40'
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

	<!-- Placeholder: starting a game, the scenario and the roles come with the rules. -->
	<p class="mt-auto py-3 text-center text-muted-foreground">
		{#if alone}
			Waiting for a teammate…
		{:else}
			Waiting for {isHost ? 'the game to be built' : (host?.name ?? 'the host')}…
		{/if}
	</p>
</div>
