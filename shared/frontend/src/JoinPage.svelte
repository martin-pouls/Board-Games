<!--
	The page behind an invite link: shows who is hosting, asks for a nickname and joins the game.
	A player who already has a session for the code goes straight to the game. The game's
	routes/join/[code]/+page.svelte renders it with its title and the app's collaborators:
	navigation (goto, built hrefs), the REST calls, the session store and its error texts.
-->
<script lang="ts" generics="Code extends string">
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { ApiError } from './api.js';
	import type { Goto, LobbySession, LobbySummary } from './lobby.js';
	import type { SessionStore } from './session.js';
	import { Button } from './ui/button/index.js';
	import * as Card from './ui/card/index.js';
	import { Input } from './ui/input/index.js';
	import { Label } from './ui/label/index.js';
	import { Skeleton } from './ui/skeleton/index.js';

	let {
		title,
		code,
		homeHref,
		gameHref,
		goto,
		getSummary,
		joinGame,
		sessions,
		errorText
	}: {
		/** The game's name. */
		title: string;
		/** The join code from the URL. */
		code: string;
		/** The game's home page. */
		homeHref: string;
		/** The game page of a code (resolve('/g/[code]', { code })). */
		gameHref: (code: string) => string;
		goto: Goto;
		getSummary: (code: string) => Promise<LobbySummary>;
		joinGame: (code: string, name: string) => Promise<LobbySession>;
		sessions: SessionStore;
		/** The game's message for an error code; it also gets the codes 'invalid_name' and 'network'. */
		errorText: (code: Code) => string;
	} = $props();

	// The two codes the page raises itself; every game has them.
	const codeOf = (e: unknown) => (e instanceof ApiError ? e.code : 'network') as Code;
	const invalidName = 'invalid_name' as Code;

	let summary = $state<LobbySummary | null>(null);
	let loadError = $state<string | null>(null);
	let nickname = $state('');
	let busy = $state(false);

	const name = $derived(nickname.trim());

	onMount(async () => {
		if (sessions.get(code)) {
			await goto(gameHref(code), { replaceState: true });
			return;
		}
		nickname = sessions.getNickname();
		try {
			summary = await getSummary(code);
		} catch (e) {
			loadError = errorText(codeOf(e));
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
			toast.error(errorText(invalidName));
			return;
		}
		busy = true;
		try {
			const res = await joinGame(code, name);
			sessions.setNickname(name);
			sessions.save(res.gameId, { token: res.token, playerId: res.playerId, name });
			await goto(gameHref(res.gameId));
		} catch (e) {
			toast.error(errorText(codeOf(e)));
		} finally {
			busy = false;
		}
	}
</script>

<div class="mx-auto flex w-full max-w-md flex-1 flex-col justify-center">
	<Card.Root>
		<Card.Header>
			<Card.Title class="text-2xl">{title}</Card.Title>
			<Card.Description class="font-mono tracking-widest">{code}</Card.Description>
		</Card.Header>
		<Card.Content class="flex flex-col gap-4">
			{#if loadError}
				<p class="text-destructive">{loadError}</p>
				<Button variant="outline" class="h-12 text-base" href={homeHref}>Home</Button>
			{:else if !summary}
				<Skeleton class="h-6 w-2/3" />
				<Skeleton class="h-12 w-full" />
			{:else if blocked}
				<p>{blocked}</p>
				<Button variant="outline" class="h-12 text-base" href={homeHref}>Home</Button>
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
