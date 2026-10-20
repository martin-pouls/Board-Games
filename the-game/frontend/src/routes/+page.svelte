<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import HomePage from '@board-games/shared/HomePage.svelte';
	import { createGame, getSummary, joinGame } from '#lib/api.js';
	import { errorText } from '#lib/errors.js';
	import { lobbyOptions } from '#lib/game/lobbyOptions.js';
	import { sessions } from '#lib/session.js';
	import * as Card from '#lib/components/ui/card/index.js';
	import StartCard from '#lib/components/game/StartCard.svelte';
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
		<!-- The two printed start cards: ascending from 1, descending from 100, fanned outwards. -->
		<div
			class="col-start-2 row-span-2 row-start-1 flex items-center gap-1.5 self-center pr-1 [--w:2.625rem]"
			aria-hidden="true"
		>
			<StartCard dir="up" class="-rotate-5" />
			<StartCard dir="down" class="rotate-5" />
		</div>
	{/snippet}
</HomePage>
