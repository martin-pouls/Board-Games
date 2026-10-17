<script lang="ts">
	import { resolve } from '$app/paths';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Card from '#lib/components/ui/card/index.js';
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
									<span class="font-medium">Mission {e.mission}</span>
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
	<Table {client} {view} attempt={view.attempt} />
{/if}
