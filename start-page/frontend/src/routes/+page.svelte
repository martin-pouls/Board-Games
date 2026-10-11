<script lang="ts">
	import { page } from '$app/state';
	import { gameHref, games } from '#lib/games.js';
	import { Badge } from '#lib/components/ui/badge/index.js';
	import * as Card from '#lib/components/ui/card/index.js';
	import { siteName } from '#lib/site.js';
</script>

<svelte:head>
	<title>{siteName}</title>
	<meta
		name="description"
		content="Cooperative board and card games to play live with friends in the browser."
	/>
</svelte:head>

<header class="mb-10 text-center sm:mb-14">
	<!-- A die, a meeple and a card: board and card games in general. -->
	<div class="mb-4 flex items-end justify-center gap-4 text-muted-foreground" aria-hidden="true">
		<svg viewBox="0 0 24 24" class="size-7 -rotate-12">
			<rect
				x="3"
				y="3"
				width="18"
				height="18"
				rx="4"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
			/>
			<circle cx="8.5" cy="8.5" r="1.6" fill="currentColor" />
			<circle cx="12" cy="12" r="1.6" fill="currentColor" />
			<circle cx="15.5" cy="15.5" r="1.6" fill="currentColor" />
		</svg>
		<svg viewBox="0 0 24 24" class="size-9 text-accent-piece">
			<path
				d="M12 2a3 3 0 0 1 3 3c0 .9-.4 1.7-1 2.2 2.6.5 6.5 1.6 6.5 3.8 0 1.2-1.6 1.7-3.5 1.9L20 21h-5.5L12 17l-2.5 4H4l2.9-8.1C5.1 12.7 3.5 12.2 3.5 11c0-2.2 3.9-3.3 6.5-3.8-.6-.5-1-1.3-1-2.2a3 3 0 0 1 3-3z"
				fill="currentColor"
			/>
		</svg>
		<svg viewBox="0 0 24 24" class="size-7 rotate-12">
			<rect
				x="5"
				y="2"
				width="14"
				height="20"
				rx="2.5"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
			/>
			<path d="M12 8.5 15 12l-3 3.5L9 12z" fill="currentColor" />
		</svg>
	</div>
	<h1 class="text-4xl font-bold tracking-tight sm:text-5xl">{siteName}</h1>
	<p class="mt-3 text-muted-foreground">
		Cooperative board and card games to play live with friends, right in the browser.
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
						game.id === 'regicide' && 'bg-castle ring-gold/30 group-hover:ring-gold/80',
						game.id === 'fox-duet' &&
							'bg-fox-tile ring-fox-orange/30 group-hover:ring-fox-orange/80'
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
					{:else if game.id === 'fox-duet'}
						<div
							class="relative flex h-36 items-center justify-center bg-linear-to-br from-fox-moss via-fox-moss to-fox-tile"
							aria-hidden="true"
						>
							<svg
								viewBox="0 0 32 32"
								class="size-20 text-fox-orange transition group-hover:scale-105"
							>
								<path
									d="M4 5l7 5h10l7-5-1 11-6 11h-10l-6-11z"
									fill="currentColor"
									stroke="currentColor"
									stroke-width="1"
									stroke-linejoin="round"
								/>
								<path d="M12 27l4-6 4 6z" fill="white" fill-opacity="0.85" />
							</svg>
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
									game.id === 'regicide' && 'font-display tracking-wide text-gold',
									game.id === 'fox-duet' && 'text-fox-text'
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
								game.id === 'the-game'
									? 'text-the-game-text'
									: game.id === 'fox-duet'
										? 'text-fox-text'
										: 'text-gold'
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
