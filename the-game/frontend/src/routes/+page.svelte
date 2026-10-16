<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import HomePage from '@board-games/shared/HomePage.svelte';
	import { createGame, getSummary, joinGame } from '#lib/api.js';
	import { errorText } from '#lib/errors.js';
	import { sessions } from '#lib/session.js';
	import * as Card from '#lib/components/ui/card/index.js';
</script>

<HomePage
	heroClass="relative overflow-hidden shadow-lg dark:shadow-black/30"
	heroHeaderClass="grid-cols-[1fr_auto]"
	statusLabels={{ choosingFirstPlayer: 'In progress' }}
	gameHref={(code) => resolve('/g/[code]', { code })}
	{goto}
	{createGame}
	{joinGame}
	{getSummary}
	{sessions}
	{errorText}
>
	{#snippet heroTop()}
		<div
			class="absolute inset-x-0 top-0 h-1 bg-linear-to-r from-sky-500 from-50% to-amber-500 to-50%"
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
			<span
				class="pile-chip -rotate-6 border-sky-300 bg-sky-50 text-sky-700 dark:border-sky-800 dark:bg-sky-950/60 dark:text-sky-300"
				>1</span
			>
			<span
				class="pile-chip rotate-6 border-amber-300 bg-amber-50 text-amber-700 dark:border-amber-800 dark:bg-amber-950/60 dark:text-amber-300"
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
		border-width: 2px;
		border-radius: 0.5rem;
		font-weight: 700;
		font-size: 0.95rem;
		font-variant-numeric: tabular-nums;
		box-shadow: 0 2px 6px rgb(0 0 0 / 0.12);
	}
</style>
