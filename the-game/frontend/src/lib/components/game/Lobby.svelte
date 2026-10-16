<script lang="ts">
	import HeaderButton, { headerRow } from '@board-games/shared/HeaderButton.svelte';
	import HeaderThemeToggle from '@board-games/shared/HeaderThemeToggle.svelte';
	import LobbyActions from '@board-games/shared/LobbyActions.svelte';
	import { toast } from 'svelte-sonner';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Crown from '@lucide/svelte/icons/crown';
	import House from '@lucide/svelte/icons/house';
	import LoaderCircle from '@lucide/svelte/icons/loader-circle';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Card from '#lib/components/ui/card/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import { Switch } from '#lib/components/ui/switch/index.js';
	import { errorText } from '#lib/errors.js';
	import { ActionError } from '#lib/game/connection.svelte.js';
	import type { GameClient } from '#lib/game/client.js';
	import { plural } from '#lib/game/format.js';
	import { lobbyOptions } from '#lib/game/lobbyOptions.js';
	import { cn } from '#lib/utils.js';
	import type { GameOptions, PlayerView } from '#lib/types.js';
	import ShareInvite from '@board-games/shared/ShareInvite.svelte';

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
	const solo = $derived(view.players.length === 1);
	const canEdit = $derived(isHost && connected && !busy);

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

	/** Remembered on this device as the defaults for the next game this player hosts. */
	const setOptions = (o: Partial<GameOptions>) =>
		run(async () => {
			const next = { ...view.options, ...o };
			await client.setOptions(next);
			lobbyOptions.save(next);
		});
	// Reduced hand size only exists together with Pro.
	const setPro = (pro: boolean) => setOptions(pro ? { pro } : { pro, reducedHandSize: false });
	const setReduced = (reducedHandSize: boolean) => setOptions({ reducedHandSize });
	// Extreme is independent of Pro and Reduced hand size.
	const extreme = $derived(view.options.variant === 'extreme');
	const setExtreme = (on: boolean) => setOptions({ variant: on ? 'extreme' : 'classic' });
</script>

<div class="mx-auto flex w-full max-w-md flex-1 flex-col gap-3">
	<header class={headerRow}>
		<HeaderButton icon={House} label="Home" href={resolve('/')} />
		<h1 class="text-lg font-semibold">Lobby</h1>
		<div class="ml-auto flex items-center gap-1.5">
			<LobbyActions alone={solo} {isHost} {connected} {busy} onend={endGame} onleave={leave} />
			<HeaderThemeToggle />
		</div>
	</header>

	<Card.Root>
		<Card.Content>
			<ShareInvite
				code={view.gameId}
				link={`${globalThis.location?.origin ?? ''}${resolve('/join/[code]', { code: view.gameId })}`}
				title="The Game"
			/>
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Header>
			<Card.Title>Players ({view.players.length}/{view.rules.maxPlayers})</Card.Title>
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
							<Crown class="ml-auto size-4 shrink-0 text-amber-500" aria-label="Host" />
						{/if}
					</li>
				{/each}
			</ul>
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Header>
			<Card.Title>Options</Card.Title>
		</Card.Header>
		<Card.Content class="flex flex-col gap-4">
			<div class="flex flex-col gap-1.5">
				<div class="flex items-center justify-between gap-3">
					<Label for="opt-extreme" class="text-base">Extreme (command cards)</Label>
					<Switch id="opt-extreme" disabled={!canEdit} bind:checked={() => extreme, setExtreme} />
				</div>
				{#if extreme}
					<p class="text-sm text-muted-foreground">
						28 cards carry commands: STOP, Skull, 3!, No communication, No backwards trick, Single
						pile, Draw only one.
					</p>
				{/if}
			</div>
			<div class="flex items-center justify-between gap-3">
				<Label for="opt-pro" class="text-base">Pro (min 3 cards per turn)</Label>
				<Switch id="opt-pro" disabled={!canEdit} bind:checked={() => view.options.pro, setPro} />
			</div>
			<div class="flex items-center justify-between gap-3">
				<Label for="opt-reduced" class={cn('text-base', !view.options.pro && 'opacity-50')}>
					Reduced hand size (−1)
				</Label>
				<Switch
					id="opt-reduced"
					disabled={!canEdit || !view.options.pro}
					bind:checked={() => view.options.reducedHandSize, setReduced}
				/>
			</div>
			<p class="text-sm text-muted-foreground">
				Hand size {view.rules.handSize} · min {view.rules.minCardsPerTurn} cards per turn{extreme
					? ' · Extreme'
					: ''}
			</p>
		</Card.Content>
	</Card.Root>

	<div
		class="sticky bottom-0 -mx-3 mt-auto px-3 pt-1 pb-[max(0.5rem,env(safe-area-inset-bottom))] before:absolute before:inset-0 before:-z-10 before:backdrop-blur"
	>
		{#if isHost}
			<Button
				size="lg"
				class="h-12 w-full text-base"
				disabled={!connected || busy || view.players.length < view.rules.minPlayers}
				onclick={() => run(() => client.startGame())}
			>
				Start game ({plural(view.players.length, 'player')})
			</Button>
		{:else}
			<p class="flex h-12 items-center justify-center gap-2 text-muted-foreground">
				<LoaderCircle class="size-4 animate-spin" />
				Waiting for {host?.name ?? 'the host'} to start…
			</p>
		{/if}
	</div>
</div>
