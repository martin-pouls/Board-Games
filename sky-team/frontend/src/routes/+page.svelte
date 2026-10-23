<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import HomePage from '@board-games/shared/HomePage.svelte';
	import { createGame, getSummary, joinGame } from '#lib/api.js';
	import { errorText } from '#lib/errors.js';
	import { sessions } from '#lib/session.js';
	import * as Card from '#lib/components/ui/card/index.js';
	import HomeBand from '#lib/components/HomeBand.svelte';
</script>

<HomePage
	heroClass="overflow-hidden pt-0 shadow-lg dark:shadow-black/30"
	gameHref={(code) => resolve('/g/[code]', { code })}
	{goto}
	{createGame}
	{joinGame}
	{getSummary}
	{sessions}
	{errorText}
>
	{#snippet heroTop()}
		<!-- A band along the card's top edge. -->
		<HomeBand />
	{/snippet}
	{#snippet hero()}
		<Card.Title class="font-display text-4xl">Sky Team</Card.Title>
		<Card.Description>Two pilots, one cockpit. Land the plane together.</Card.Description>
	{/snippet}
</HomePage>
