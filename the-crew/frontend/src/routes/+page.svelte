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
	import HomeBand from '#lib/components/HomeBand.svelte';
</script>

<HomePage
	heroClass="overflow-hidden pt-0 shadow-lg dark:shadow-black/30"
	gameHref={(code) => resolve('/g/[code]', { code })}
	{goto}
	createGame={(name) => createGame(name, lobbyOptions.load() ?? undefined)}
	{joinGame}
	{getSummary}
	{sessions}
	{errorText}
	statusLabels={{ debrief: 'Debrief' }}
>
	{#snippet heroTop()}
		<!-- A strip of the start-page tile's scene along the card's top edge. -->
		<HomeBand />
	{/snippet}
	{#snippet hero()}
		<Card.Title class="font-display text-[2rem] leading-[1.05] font-bold text-balance">
			{terms.title}
		</Card.Title>
		<Card.Description>{terms.tagline}</Card.Description>
	{/snippet}
</HomePage>
