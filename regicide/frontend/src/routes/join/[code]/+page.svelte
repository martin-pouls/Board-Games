<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { ApiError, getSummary, joinGame } from '#lib/api.js';
	import { errorText } from '#lib/errors.js';
	import { sessions } from '#lib/session.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Card from '#lib/components/ui/card/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import { Skeleton } from '#lib/components/ui/skeleton/index.js';
	import type { GameSummary } from '#lib/types.js';

	const code = $derived((page.params.code ?? '').toUpperCase());

	let summary = $state<GameSummary | null>(null);
	let loadError = $state<string | null>(null);
	let nickname = $state('');
	let busy = $state(false);

	const name = $derived(nickname.trim());

	onMount(async () => {
		if (sessions.get(code)) {
			await goto(resolve('/g/[code]', { code }), { replaceState: true });
			return;
		}
		nickname = sessions.getNickname();
		try {
			summary = await getSummary(code);
		} catch (e) {
			loadError = errorText(e instanceof ApiError ? e.code : 'network');
		}
	});

	const blocked = $derived(
		summary && !summary.joinable
			? summary.phase !== 'lobby'
				? 'This game has already started.'
				: 'This game is full.'
			: null
	);

	async function join() {
		if (busy) return;
		if (!name) {
			toast.error(errorText('invalid_name'));
			return;
		}
		busy = true;
		try {
			const res = await joinGame(code, name);
			sessions.setNickname(name);
			sessions.save(res.gameId, { token: res.token, playerId: res.playerId, name });
			await goto(resolve('/g/[code]', { code: res.gameId }));
		} catch (e) {
			toast.error(errorText(e instanceof ApiError ? e.code : 'network'));
		} finally {
			busy = false;
		}
	}
</script>

<div class="mx-auto flex w-full max-w-md flex-1 flex-col justify-center">
	<Card.Root>
		<Card.Header>
			<Card.Title class="text-2xl">Regicide</Card.Title>
			<Card.Description class="font-mono tracking-widest">{code}</Card.Description>
		</Card.Header>
		<Card.Content class="flex flex-col gap-4">
			{#if loadError}
				<p class="text-destructive">{loadError}</p>
				<Button variant="outline" class="h-12 text-base" href={resolve('/')}>Home</Button>
			{:else if !summary}
				<Skeleton class="h-6 w-2/3" />
				<Skeleton class="h-12 w-full" />
			{:else if blocked}
				<p>{blocked}</p>
				<Button variant="outline" class="h-12 text-base" href={resolve('/')}>Home</Button>
			{:else}
				<p class="text-lg">
					Join {summary.hostName}'s game ({summary.playerCount}/{summary.maxPlayers})
				</p>
				<form
					class="flex flex-col gap-4"
					onsubmit={(e) => {
						e.preventDefault();
						join();
					}}
				>
					<div class="flex flex-col gap-2">
						<Label for="nickname">Your name</Label>
						<Input
							id="nickname"
							class="h-12 text-base"
							bind:value={nickname}
							maxlength={20}
							autocomplete="nickname"
							placeholder="Nickname"
						/>
					</div>
					<Button type="submit" size="lg" class="h-12 text-base" disabled={busy}>Join</Button>
				</form>
			{/if}
		</Card.Content>
	</Card.Root>
</div>
