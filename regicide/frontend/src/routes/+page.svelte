<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import HomePage from '@board-games/shared/HomePage.svelte';
	import { createGame, getSummary, joinGame } from '#lib/api.js';
	import { errorText } from '#lib/errors.js';
	import { sessions } from '#lib/session.js';
	import * as Card from '#lib/components/ui/card/index.js';
	import PlayingCard from '#lib/components/game/PlayingCard.svelte';
</script>

<HomePage
	heroClass="shadow-lg dark:shadow-black/30"
	heroHeaderClass="grid-cols-[1fr_auto]"
	gameHref={(code) => resolve('/g/[code]', { code })}
	{goto}
	{createGame}
	{joinGame}
	{getSummary}
	{sessions}
	{errorText}
>
	{#snippet hero()}
		<Card.Title class="text-3xl">Regicide</Card.Title>
		<Card.Description>Storm the castle together.</Card.Description>
		<!-- The enemy and the one card that can stop it: a King and the Jester. Offset, not rotated:
		     Firefox renders rotated cards blurry. -->
		<div
			class="col-start-2 row-span-2 row-start-1 flex items-start self-center pr-1"
			aria-hidden="true"
		>
			<PlayingCard card={{ suit: 'spades', rank: 13 }} size="sm" />
			<span class="mt-2.5 -ml-4"><PlayingCard card={{ suit: '', rank: 0 }} size="sm" /></span>
		</div>
	{/snippet}
</HomePage>
