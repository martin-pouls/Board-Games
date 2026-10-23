<script lang="ts">
	import GameScreen from '#lib/components/game/GameScreen.svelte';
	import { GameConnection } from '#lib/game/connection.svelte.js';
	import type { PageProps } from './$types';

	// Dev only: one fixture view on the game screen, through a connection that never connects.
	let { data }: PageProps = $props();

	const client = $derived(new GameConnection(data.view.gameId, 'fixture').showFixture(data.view));
</script>

<svelte:head>
	<title>{data.name} · Fixture</title>
</svelte:head>

{#key data.name}
	<GameScreen {client} view={data.view} />
{/key}
