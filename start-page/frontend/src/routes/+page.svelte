<script lang="ts">
	import { games } from '#lib/games.js';
</script>

<header class="mb-10 text-center sm:mb-14">
	<p class="mb-3 text-2xl tracking-[0.5em] text-chalk-muted" aria-hidden="true">
		♠<span class="text-red-400">♥</span>♣<span class="text-red-400">♦</span>
	</p>
	<h1 class="text-4xl font-bold tracking-tight sm:text-5xl">Card Games</h1>
	<p class="mt-3 text-chalk-muted">
		Cooperative card games to play live with friends, right in the browser.
	</p>
</header>

<ul class="grid gap-5 sm:grid-cols-2">
	{#each games as game (game.id)}
		<li>
			<!-- Each game is a separate app on the same host: rel="external" makes it a full page load. -->
			<a
				href={game.href}
				rel="external"
				class={[
					'group flex h-full flex-col overflow-hidden rounded-2xl border shadow-lg shadow-black/40 transition',
					'hover:-translate-y-1 hover:shadow-xl focus-visible:-translate-y-1 focus-visible:outline-2 focus-visible:outline-offset-4',
					game.id === 'the-game' &&
						'border-violet/40 bg-zinc-900 hover:border-violet focus-visible:outline-violet',
					game.id === 'regicide' &&
						'border-gold/25 bg-castle hover:border-gold/70 focus-visible:outline-gold'
				]}
			>
				{#if game.id === 'the-game'}
					<div
						class="relative flex h-36 items-center justify-center gap-3 bg-linear-to-br from-violet to-indigo-700"
						aria-hidden="true"
					>
						<span class="card-face -rotate-6 bg-white text-violet">1</span>
						<span class="card-face z-10 bg-white text-zinc-900 group-hover:-translate-y-1">57</span>
						<span class="card-face rotate-6 bg-white text-violet">100</span>
					</div>
				{:else}
					<div
						class="relative flex h-36 items-center justify-center bg-linear-to-br from-castle via-[oklch(0.24_0.06_305)] to-castle"
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

				<div class="flex flex-1 flex-col gap-3 p-5">
					<div class="flex items-baseline justify-between gap-3">
						<h2
							class={[
								'text-2xl font-bold',
								game.id === 'regicide' && 'font-display tracking-wide text-gold'
							]}
						>
							{game.name}
						</h2>
						<span class="shrink-0 text-sm text-chalk-muted">{game.players}</span>
					</div>
					<p class={['text-chalk/85', game.id === 'regicide' && 'text-parchment/85']}>
						{game.tagline}
					</p>
					<ul class="flex flex-wrap gap-2">
						{#each game.features as feature (feature)}
							<li class="rounded-full bg-white/8 px-2.5 py-0.5 text-xs text-chalk-muted">
								{feature}
							</li>
						{/each}
					</ul>
					<span
						class={[
							'mt-auto pt-2 font-semibold',
							game.id === 'the-game' ? 'text-violet-300' : 'text-gold'
						]}
					>
						Play {game.name}
						<span class="inline-block transition group-hover:translate-x-1">→</span>
					</span>
				</div>
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
