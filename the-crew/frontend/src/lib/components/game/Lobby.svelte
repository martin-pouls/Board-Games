<script lang="ts">
	import HeaderButton, { headerRow } from '@board-games/shared/HeaderButton.svelte';
	import HeaderThemeToggle from '@board-games/shared/HeaderThemeToggle.svelte';
	import LobbyActions from '@board-games/shared/LobbyActions.svelte';
	import { toast } from 'svelte-sonner';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Bot from '@lucide/svelte/icons/bot';
	import Crown from '@lucide/svelte/icons/crown';
	import House from '@lucide/svelte/icons/house';
	import LoaderCircle from '@lucide/svelte/icons/loader-circle';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Card from '#lib/components/ui/card/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import { Switch } from '#lib/components/ui/switch/index.js';
	import { getMissions } from '#lib/api.js';
	import { errorText } from '#lib/errors.js';
	import { ActionError } from '#lib/game/connection.svelte.js';

	import type { GameClient } from '#lib/game/client.js';
	import { cn } from '#lib/utils.js';
	import type { MissionDef, PlayerView } from '#lib/types.js';
	import ShareInvite from '@board-games/shared/ShareInvite.svelte';
	import MissionBriefing from './MissionBriefing.svelte';
	import MissionPicker from './MissionPicker.svelte';

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

	/** The mission table, for the grid markers; the lobby works without it. */
	let catalogue = $state<MissionDef[] | null>(null);
	let catalogueLoading = $state(true);

	function loadCatalogue() {
		catalogueLoading = true;
		getMissions().then(
			(c) => {
				catalogue = c.missions;
				catalogueLoading = false;
			},
			() => {
				catalogueLoading = false;
			}
		);
	}
	loadCatalogue();

	const def = $derived(
		view.missionDef ?? catalogue?.find((m) => m.number === view.mission) ?? null
	);
	const entry = $derived(view.logbook.find((e) => e.mission === view.mission));
	/** Attempts of all missions, plus one for each distress signal. */
	const logbookTotal = $derived(view.logbook.reduce((n, e) => n + e.total, 0));
	const canStart = $derived(connected && !busy && view.players.length >= 2);
	/** Only the host changes the options; everybody else sees them. */
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

	function setDistress(distress: boolean) {
		return run(() => client.setOptions({ ...view.options, distress }));
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
				<p class="mt-3 flex items-center gap-1.5 text-sm text-muted-foreground">
					<Bot class="size-4 shrink-0" />
					JARVIS joins as the third crew member
				</p>
			{/if}
		</Card.Content>
	</Card.Root>

	<Card.Root class="[--card-spacing:--spacing(3)]">
		<Card.Header>
			<Card.Title class="flex items-baseline gap-2">
				Mission
				<span
					class="ml-auto text-sm font-normal text-muted-foreground"
					title="Attempts, plus one for each mission where the distress signal was used"
				>
					Logbook total: <span class="font-semibold text-foreground">{logbookTotal}</span>
				</span>
			</Card.Title>
			{#if !isHost}
				<Card.Description>{host?.name ?? 'The host'} picks the mission.</Card.Description>
			{/if}
		</Card.Header>
		<Card.Content class="flex flex-col gap-3">
			<MissionPicker
				selected={view.mission}
				logbook={view.logbook}
				missions={catalogue}
				editable={isHost}
				disabled={!connected || busy}
				onselect={(n) => {
					if (n !== view.mission) run(() => client.selectMission(n));
				}}
			/>
			<hr class="border-border" />
			<MissionBriefing mission={view.mission} {def} {entry} loading={catalogueLoading} />
			{#if !catalogueLoading && !catalogue}
				<p class="text-xs text-muted-foreground">
					The mission table could not be loaded, so the grid shows no markers.
					<button type="button" class="underline" onclick={loadCatalogue}>Try again</button>
				</p>
			{/if}
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Header>
			<Card.Title>Options</Card.Title>
			{#if !isHost}
				<Card.Description>{host?.name ?? 'The host'} sets the options.</Card.Description>
			{/if}
		</Card.Header>
		<Card.Content class="flex flex-col gap-1.5">
			<div class="flex items-center justify-between gap-3">
				<Label for="opt-distress" class="text-base">Distress signal</Label>
				<Switch
					id="opt-distress"
					disabled={!canEdit}
					bind:checked={() => view.options.distress, setDistress}
				/>
			</div>
			<p class="text-sm text-muted-foreground">
				Lets the commander have everyone pass a card before the first trick.
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
				disabled={!canStart}
				onclick={() => run(() => client.startMission())}
			>
				{view.players.length < 2 ? 'Waiting for a crew member…' : `Start mission ${view.mission}`}
			</Button>
		{:else}
			<p class="flex h-12 items-center justify-center gap-2 text-muted-foreground">
				<LoaderCircle class="size-4 animate-spin" />
				Waiting for {host?.name ?? 'the host'} to start…
			</p>
		{/if}
	</div>
</div>
