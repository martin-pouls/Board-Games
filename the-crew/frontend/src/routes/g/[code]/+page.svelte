<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Card from '#lib/components/ui/card/index.js';
	import { Skeleton } from '#lib/components/ui/skeleton/index.js';
	import ConnectionBanner from '@board-games/shared/ConnectionBanner.svelte';
	import Lobby from '#lib/components/game/Lobby.svelte';
	import Table from '#lib/components/game/Table.svelte';
	import { GameConnection } from '#lib/game/connection.svelte.js';
	import { getMissions, leaveGame } from '#lib/api.js';
	import { terms } from '#lib/edition.js';
	import { errorText } from '#lib/errors.js';
	import { missionLabelOf } from '#lib/game/format.js';
	import type { MissionDef } from '#lib/types.js';
	import { ApiError } from '#lib/api.js';
	import { sessions } from '#lib/session.js';
	import { toast } from 'svelte-sonner';

	const code = $derived((page.params.code ?? '').toUpperCase());

	let conn = $state<GameConnection | null>(null);
	let notFound = $state(false);

	// One connection per code. The route component is reused when navigating from
	// one game to the next (rematch), so this must react to `code`, not just mount.
	$effect(() => {
		const c = code;
		notFound = false;
		const session = sessions.get(c);
		if (!session) {
			conn = null;
			goto(resolve('/join/[code]', { code: c }), { replaceState: true });
			return;
		}
		const next = new GameConnection(c, session.token);
		conn = next;
		next.connect();
		return () => next.close();
	});

	$effect(() => {
		if (!conn?.fatal) return;
		const c = conn.code;
		sessions.clear(c);
		if (conn.fatal === 'unauthorized') {
			goto(resolve('/join/[code]', { code: c }), { replaceState: true });
		} else {
			notFound = true;
		}
	});

	const view = $derived(conn?.view ?? null);

	// The mission table names the side quests ("Q1") in the logbook of an ended game.
	let missions = $state<MissionDef[] | null>(null);
	$effect(() => {
		if (view?.phase !== 'ended' || missions) return;
		getMissions().then(
			(c) => (missions = c.missions),
			() => {}
		);
	});

	async function leave() {
		const c = conn;
		const session = sessions.get(code);
		if (!c || !session) return;
		c.close();
		try {
			await leaveGame(code, session.token);
		} catch (e) {
			toast.error(errorText(e instanceof ApiError ? e.code : 'internal'));
			return;
		}
		sessions.clear(code);
		await goto(resolve('/'));
	}
</script>

<svelte:head>
	<title>{code} · {terms.title}</title>
</svelte:head>

{#if conn?.status === 'reconnecting'}
	<ConnectionBanner />
{/if}

{#if notFound}
	<div class="mx-auto flex w-full max-w-md flex-1 flex-col justify-center">
		<Card.Root>
			<Card.Header>
				<Card.Title class="text-2xl">Game not found</Card.Title>
				<Card.Description>
					The game <span class="font-mono tracking-widest">{code}</span> doesn't exist anymore.
				</Card.Description>
			</Card.Header>
			<Card.Content>
				<Button class="h-12 w-full text-base" href={resolve('/')}>Back to home</Button>
			</Card.Content>
		</Card.Root>
	</div>
{:else if conn && view}
	{#if view.phase === 'lobby'}
		<Lobby client={conn} {view} onleave={leave} />
	{:else if view.phase === 'ended'}
		{@const host = view.players.find((p) => p.id === view.hostId)}
		{@const total = view.logbook.reduce((n, e) => n + e.total, 0)}
		<div class="mx-auto flex w-full max-w-md flex-1 flex-col justify-center">
			<Card.Root>
				<Card.Header>
					<Card.Title class="text-2xl">Game ended</Card.Title>
					<Card.Description>
						{host?.isYou ? 'You' : (host?.name ?? 'The host')} ended this game.
					</Card.Description>
				</Card.Header>
				<Card.Content class="flex flex-col gap-4">
					{#if view.logbook.length > 0}
						<section aria-labelledby="logbook-title" class="flex flex-col gap-2">
							<h2 id="logbook-title" class="flex items-baseline font-semibold">
								Logbook
								<span class="ml-auto text-sm font-normal text-muted-foreground">
									Total: <span class="font-semibold text-foreground">{total}</span>
								</span>
							</h2>
							<ul class="flex flex-col divide-y divide-border text-sm">
								{#each view.logbook as e (e.mission)}
									<li class="flex items-baseline gap-2 py-1.5">
										<span class="font-medium">Mission {missionLabelOf(e.mission, missions)}</span>
										<span class="text-muted-foreground">
											{e.attempts}
											{e.attempts === 1 ? 'attempt' : 'attempts'}{e.distress
												? ' · distress signal'
												: ''}
										</span>
										<span class="ml-auto">
											{#if e.completed}
												<span class="font-semibold">{e.total}</span>
											{:else}
												<span class="text-muted-foreground">not completed</span>
											{/if}
										</span>
									</li>
								{/each}
							</ul>
						</section>
					{/if}
					<Button class="h-12 w-full text-base" href={resolve('/')}>Back to home</Button>
				</Card.Content>
			</Card.Root>
		</div>
	{:else if view.attempt}
		<Table client={conn} {view} attempt={view.attempt} />
	{/if}
{:else}
	<div class="mx-auto flex w-full max-w-md flex-1 flex-col gap-3" aria-busy="true">
		<Skeleton class="h-6 w-1/3" />
		<Skeleton class="h-8 w-full" />
		<Skeleton class="h-64 w-full rounded-2xl" />
		<div class="flex gap-1">
			<Skeleton class="h-24 flex-1" />
			<Skeleton class="h-24 flex-1" />
			<Skeleton class="h-24 flex-1" />
			<Skeleton class="h-24 flex-1" />
		</div>
	</div>
{/if}
