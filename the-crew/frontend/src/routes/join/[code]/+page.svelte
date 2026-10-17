<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import JoinPage from '@board-games/shared/JoinPage.svelte';
	import { getSummary, joinGame } from '#lib/api.js';
	import { terms } from '#lib/edition.js';
	import { errorText } from '#lib/errors.js';
	import { sessions } from '#lib/session.js';

	const code = $derived((page.params.code ?? '').toUpperCase());
</script>

<JoinPage
	title={terms.title}
	{code}
	homeHref={resolve('/')}
	gameHref={(code) => resolve('/g/[code]', { code })}
	{goto}
	{getSummary}
	{joinGame}
	{sessions}
	{errorText}
/>
