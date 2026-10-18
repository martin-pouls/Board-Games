<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import HomePage from '@board-games/shared/HomePage.svelte';
	import { createGame, getSummary, joinGame } from '#lib/api.js';
	import { errorText } from '#lib/errors.js';
	import { lobbyOptions } from '#lib/game/lobbyOptions.js';
	import { sessions } from '#lib/session.js';
	import * as Card from '#lib/components/ui/card/index.js';
</script>

<HomePage
	heroClass="relative overflow-hidden shadow-lg dark:shadow-black/30"
	heroHeaderClass="grid-cols-[1fr_auto]"
	statusLabels={{ choosingFirstPlayer: 'In progress' }}
	gameHref={(code) => resolve('/g/[code]', { code })}
	{goto}
	createGame={(name) => createGame(name, lobbyOptions.load() ?? undefined)}
	{joinGame}
	{getSummary}
	{sessions}
	{errorText}
>
	{#snippet heroTop()}
		<div
			class="absolute inset-x-0 top-0 h-1 bg-linear-to-r from-up from-50% to-down to-50%"
			aria-hidden="true"
		></div>
	{/snippet}
	{#snippet hero()}
		<Card.Title class="text-3xl">The Game</Card.Title>
		<Card.Description>Play as a team, as long as you can.</Card.Description>
		<!-- The two kinds of piles: ascending from 1, descending from 100. -->
		<div
			class="col-start-2 row-span-2 row-start-1 flex items-center gap-1 self-center pr-1"
			aria-hidden="true"
		>
			<span class="pile-chip -rotate-6 border-up bg-up-soft font-display text-up-label">1</span>
			<span class="pile-chip rotate-6 border-down bg-down-soft font-display text-down-label"
				>100</span
			>
		</div>
	{/snippet}
</HomePage>

<style>
	.pile-chip {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 2.25rem;
		height: 3.25rem;
		border-width: 1.5px;
		border-radius: 0.5rem;
		font-size: 1.5rem;
		font-variant-numeric: tabular-nums;
	}
</style>
