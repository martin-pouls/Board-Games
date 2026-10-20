<script lang="ts">
	import { page } from '$app/state';
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
	import GameArt from '#lib/components/GameArt.svelte';
	import { gameHref, games, type Game } from '#lib/games.js';
	import { siteName } from '#lib/site.js';

	/** Each tile borrows its game's colours: tile, text and ring; the art carries the rest. */
	const look: Record<Game['id'], { tile: string; name?: string; muted: string }> = {
		'the-game': {
			tile: 'bg-the-game-tile text-the-game-fg ring-the-game-ring',
			muted: 'text-the-game-muted'
		},
		regicide: {
			tile: 'bg-regicide-tile text-regicide-fg ring-regicide-ring',
			muted: 'text-regicide-muted'
		},
		'fox-duet': {
			tile: 'bg-fox-tile text-fox-fg ring-fox-ring',
			muted: 'text-fox-muted'
		},
		'the-crew': {
			tile: 'bg-crew-tile text-foreground ring-crew-accent/30',
			name: 'text-crew-text',
			muted: 'text-muted-foreground'
		},
		'deep-sea': {
			tile: 'bg-deep-tile text-foreground ring-deep-accent/30',
			name: 'text-deep-text',
			muted: 'text-muted-foreground'
		}
	};

	/**
	 * Desktop tiles have one fixed size: two per row from md, three from lg, each spanning two
	 * columns of a half-tile grid, so a short last row can be centred.
	 */
	function centre(i: number): string[] {
		const n = games.length;
		return [
			n % 2 === 1 && i === n - 1 ? 'md:col-start-2' : '',
			n % 3 === 2 && i === n - 2
				? 'lg:col-start-2'
				: n % 3 === 1 && i === n - 1
					? 'lg:col-start-3'
					: 'lg:col-start-auto'
		];
	}
</script>

<svelte:head>
	<title>{siteName}</title>
	<meta
		name="description"
		content="Cooperative board and card games to play live with friends in the browser."
	/>
</svelte:head>

<header class="mb-6 text-center md:mt-2 md:mb-10">
	<div class="flex items-center justify-center gap-2.5">
		<!-- A meeple: board and card games in general. -->
		<svg
			viewBox="0 0 24 24"
			class="size-7.5 shrink-0 text-accent-piece md:size-10"
			aria-hidden="true"
		>
			<path
				d="M12 2a3 3 0 0 1 3 3c0 .9-.4 1.7-1 2.2 2.6.5 6.5 1.6 6.5 3.8 0 1.2-1.6 1.7-3.5 1.9L20 21h-5.5L12 17l-2.5 4H4l2.9-8.1C5.1 12.7 3.5 12.2 3.5 11c0-2.2 3.9-3.3 6.5-3.8-.6-.5-1-1.3-1-2.2a3 3 0 0 1 3-3z"
				fill="currentColor"
			/>
		</svg>
		<h1 class="text-3xl font-extrabold tracking-tight md:text-5xl">{siteName}</h1>
	</div>
	<p class="mx-auto mt-2 max-w-[34ch] text-balance text-muted-foreground md:max-w-none">
		Cooperative board and card games to play live with friends, right in the browser.
	</p>
</header>

<!-- Phone: a shelf of rows. Desktop: fixed-size tiles (148 px half-columns, a tile spans two). -->
<ul
	class="flex flex-col gap-2.5 md:grid md:auto-rows-fr md:grid-cols-[repeat(4,9.25rem)] md:justify-center md:gap-4 lg:grid-cols-[repeat(6,9.25rem)]"
>
	{#each games as game, i (game.id)}
		<li class={['md:col-span-2', ...centre(i)]}>
			<!-- Each game is a separate app (same host in prod, own port in dev): rel="external" makes it a full page load. -->
			<a
				href={gameHref(game, page.url, import.meta.env.DEV)}
				rel="external"
				class={[
					'flex h-full items-center gap-3.5 rounded-2xl py-2 pr-3 pl-2 shadow-md ring-1 shadow-tile-shadow transition',
					'hover:-translate-y-0.5 hover:shadow-lg focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-ring active:translate-y-0',
					'md:flex-col md:items-stretch md:gap-0 md:overflow-hidden md:p-0 md:hover:-translate-y-1',
					look[game.id].tile
				]}
			>
				<GameArt game={game.id} class="size-22 rounded-xl md:h-33 md:w-full md:rounded-none" />
				<div class="flex min-w-0 flex-1 flex-col gap-0.5 md:px-4.5 md:pt-4 md:pb-4.5">
					<h2
						class={[
							'text-lg leading-tight font-extrabold text-balance md:text-xl',
							look[game.id].name
						]}
					>
						{game.name}
					</h2>
					<span class={['text-sm md:mb-1.5', look[game.id].muted]}>{game.players}</span>
					<p class="text-sm opacity-90 md:hidden">{game.tagline}</p>
					<p class="hidden text-sm opacity-90 md:block">{game.description}</p>
				</div>
				<ChevronRight class="size-5 shrink-0 md:hidden {look[game.id].muted}" aria-hidden="true" />
			</a>
		</li>
	{/each}
</ul>
