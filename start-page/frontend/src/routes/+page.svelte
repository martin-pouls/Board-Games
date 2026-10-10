<script lang="ts">
	import { page } from '$app/state';
	import { gameHref, games } from '#lib/games.js';
	import { Badge } from '#lib/components/ui/badge/index.js';
	import * as Card from '#lib/components/ui/card/index.js';
</script>

<header class="mb-10 text-center sm:mb-14">
	<p class="mb-3 text-2xl tracking-[0.5em] text-muted-foreground" aria-hidden="true">
		♠<span class="text-suit-red">♥</span>♣<span class="text-suit-red">♦</span>
	</p>
	<h1 class="text-4xl font-bold tracking-tight sm:text-5xl">Card Games</h1>
	<p class="mt-3 text-muted-foreground">
		Cooperative card games to play live with friends, right in the browser.
	</p>
</header>

<ul class="grid gap-5 sm:grid-cols-2">
	{#each games as game (game.id)}
		<li>
			<!-- Each game is a separate app (same host in prod, own port in dev): rel="external" makes it a full page load. -->
			<a
				href={gameHref(game, page.url, import.meta.env.DEV)}
				rel="external"
				class="group block h-full rounded-2xl transition hover:-translate-y-1 focus-visible:-translate-y-1 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-ring"
			>
				<Card.Root
					class={[
						'h-full gap-0 rounded-2xl py-0 shadow-lg transition-shadow group-hover:shadow-xl dark:shadow-black/40',
						game.id === 'the-game' &&
							'bg-the-game-tile ring-the-game-blue/30 group-hover:ring-the-game-blue/80',
						game.id === 'regicide' && 'bg-castle ring-gold/30 group-hover:ring-gold/80'
					]}
				>
					{#if game.id === 'the-game'}
						<div
							class="relative flex h-36 items-center justify-center gap-3 bg-linear-to-br from-the-game-blue via-blue-800 to-blue-950"
							aria-hidden="true"
						>
							<span class="card-face -rotate-6 bg-white text-the-game-blue">1</span>
							<span class="card-face z-10 bg-white text-zinc-900 group-hover:-translate-y-1"
								>57</span
							>
							<span class="card-face rotate-6 bg-white text-the-game-orange">100</span>
						</div>
					{:else}
						<div
							class="relative flex h-36 items-center justify-center bg-linear-to-br from-castle via-castle-banner to-castle"
							aria-hidden="true"
						>
							<svg viewBox="0 0 32 32" class="size-20 text-gold transition group-hover:scale-105">
								<path
									d="M7 22 5.5 10.5l5.5 4.5L16 8l5 7 5.5-4.5L25 22z"
									fill="currentColor"
									stroke="currentColor"
									stroke-width="1"
									stroke-linejoin="round"
								/>
								<rect x="7" y="23.5" width="18" height="2.5" rx="1" fill="currentColor" />
							</svg>
						</div>
					{/if}

					<Card.Content class="flex flex-1 flex-col gap-3 p-5 text-base">
						<div class="flex items-baseline justify-between gap-3">
							<h2
								class={[
									'text-2xl font-bold',
									game.id === 'regicide' && 'font-display tracking-wide text-gold'
								]}
							>
								{game.name}
							</h2>
							<span class="shrink-0 text-sm text-muted-foreground">{game.players}</span>
						</div>
						<p class={game.id === 'regicide' ? 'text-castle-text' : 'text-foreground/85'}>
							{game.tagline}
						</p>
						<ul class="flex flex-wrap gap-2">
							{#each game.features as feature (feature)}
								<li><Badge variant="secondary">{feature}</Badge></li>
							{/each}
						</ul>
						<span
							class={[
								'mt-auto pt-2 font-semibold',
								game.id === 'the-game' ? 'text-the-game-text' : 'text-gold'
							]}
						>
							Play {game.name}
							<span class="inline-block transition group-hover:translate-x-1">→</span>
						</span>
					</Card.Content>
				</Card.Root>
			</a>
		</li>
	{/each}
</ul>

<style>
	.card-face {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 3.5rem;
		height: 5rem;
		border-radius: 0.5rem;
		font-weight: 800;
		font-size: 1.25rem;
		box-shadow: 0 6px 14px rgb(0 0 0 / 0.35);
		transition: translate 150ms;
	}
</style>
