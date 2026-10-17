<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Card from '#lib/components/ui/card/index.js';
	import { Skeleton } from '#lib/components/ui/skeleton/index.js';
	import GameScreen from '#lib/components/game/GameScreen.svelte';
	import ConnectionBanner from '@board-games/shared/ConnectionBanner.svelte';
	import { GameConnection } from '#lib/game/connection.svelte.js';
	import { ApiError, leaveGame } from '#lib/api.js';
	import { errorText } from '#lib/errors.js';
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
	<title>{code} · The Game</title>
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
	<GameScreen client={conn} {view} onleave={leave} />
{:else}
	<div class="mx-auto flex w-full max-w-md flex-1 flex-col gap-3" aria-busy="true">
		<Skeleton class="h-6 w-1/3" />
		<Skeleton class="h-8 w-full" />
		<div class="grid flex-1 grid-cols-2 grid-rows-2 gap-2">
			<Skeleton class="h-full rounded-2xl" />
			<Skeleton class="h-full rounded-2xl" />
			<Skeleton class="h-full rounded-2xl" />
			<Skeleton class="h-full rounded-2xl" />
		</div>
		<Skeleton class="h-16 w-full" />
	</div>
{/if}
