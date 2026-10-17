<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import HomePage from '@board-games/shared/HomePage.svelte';
	import { createGame, getSummary, joinGame } from '#lib/api.js';
	import { terms } from '#lib/edition.js';
	import { errorText } from '#lib/errors.js';
	import { lobbyOptions } from '#lib/game/lobbyOptions.js';
	import { sessions } from '#lib/session.js';
	import * as Card from '#lib/components/ui/card/index.js';
</script>

<HomePage
	gameHref={(code) => resolve('/g/[code]', { code })}
	{goto}
	createGame={(name) => createGame(name, lobbyOptions.load() ?? undefined)}
	{joinGame}
	{getSummary}
	{sessions}
	{errorText}
	statusLabels={{ debrief: 'Debrief' }}
>
	{#snippet hero()}
		<Card.Title class="text-3xl">{terms.title}</Card.Title>
		<Card.Description>{terms.tagline}</Card.Description>
	{/snippet}
</HomePage>
