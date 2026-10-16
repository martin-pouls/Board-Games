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
	gameHref={(code) => resolve('/g/[code]', { code })}
	{goto}
	{createGame}
	{joinGame}
	{getSummary}
	{sessions}
	{errorText}
>
	{#snippet hero()}
		<Card.Title class="text-3xl">Fox Duet</Card.Title>
		<Card.Description>Two friends, one forest path. Collect every gem together.</Card.Description>
	{/snippet}
</HomePage>
