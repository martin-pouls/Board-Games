<script lang="ts">
	import { resolve } from '$app/paths';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Card from '#lib/components/ui/card/index.js';
	import GameOver from '#lib/components/game/GameOver.svelte';
	import Lobby from '#lib/components/game/Lobby.svelte';
	import Table from '#lib/components/game/Table.svelte';
	import type { GameClient } from '#lib/game/client.js';
	import type { PlayerView } from '#lib/types.js';

	/** The game screen for one view: the live game route and the dev fixtures route render it. */
	let {
		client,
		view,
		onleave
	}: { client: GameClient; view: PlayerView; onleave?: () => Promise<void> | void } = $props();
</script>

{#if view.phase === 'lobby'}
	<Lobby {client} {view} {onleave} />
{:else if view.phase === 'playing'}
	<Table {client} {view} />
{:else if view.phase === 'ended'}
	{@const host = view.players.find((p) => p.id === view.hostId)}
	<div class="mx-auto flex w-full max-w-md flex-1 flex-col justify-center">
		<Card.Root>
			<Card.Header>
				<Card.Title class="text-2xl">Lobby closed</Card.Title>
				<Card.Description>
					{host?.isYou ? 'You' : (host?.name ?? 'The host')} closed this lobby before the game started.
				</Card.Description>
			</Card.Header>
			<Card.Content>
				<Button class="h-12 w-full text-base" href={resolve('/')}>Back to home</Button>
			</Card.Content>
		</Card.Root>
	</div>
{:else}
	<GameOver {client} {view} />
{/if}
