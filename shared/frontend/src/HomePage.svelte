<!--
	A game's home screen: host a game, join one by code, continue the games remembered in this
	browser and browse the finished ones. The game's routes/+page.svelte renders it with the
	hero card's header (name, tagline, any decoration) and the app's collaborators: navigation
	(goto, built hrefs), the REST calls, the session store and its error texts.
-->
<script lang="ts" generics="Code extends string">
	import { onMount, type Snippet } from 'svelte';
	import ChevronDown from '@lucide/svelte/icons/chevron-down';
	import { toast } from 'svelte-sonner';
	import { ApiError } from './api.js';
	import type { Goto, LobbySession, LobbySummary } from './lobby.js';
	import { CODE_LENGTH, normalizeCode } from './protocol.js';
	import { loadRecentGames, partitionGames, statusLabel, type RecentGame } from './recent.js';
	import type { SessionStore } from './session.js';
	import { Badge } from './ui/badge/index.js';
	import { Button } from './ui/button/index.js';
	import * as Card from './ui/card/index.js';
	import { Input } from './ui/input/index.js';
	import { Label } from './ui/label/index.js';
	import { Separator } from './ui/separator/index.js';
	import { Skeleton } from './ui/skeleton/index.js';

	let {
		hero,
		heroTop,
		heroClass,
		heroHeaderClass,
		statusLabels,
		gameHref,
		goto,
		createGame,
		joinGame,
		getSummary,
		sessions,
		errorText
	}: {
		/** The content of the hero card's header: Card.Title, Card.Description and any decoration. */
		hero: Snippet;
		/** Decoration inside the hero card, before its header (e.g. an absolutely positioned stripe). */
		heroTop?: Snippet;
		/** Extra classes for the hero card. */
		heroClass?: string;
		/** Extra classes for the hero card's header (e.g. a grid with a column for decoration). */
		heroHeaderClass?: string;
		/** Labels for the game's own phases in the lists (see recent.ts statusLabel). */
		statusLabels?: Record<string, string>;
		/** The game page of a code (resolve('/g/[code]', { code })). */
		gameHref: (code: string) => string;
		goto: Goto;
		createGame: (name: string) => Promise<LobbySession>;
		joinGame: (code: string, name: string) => Promise<LobbySession>;
		getSummary: (code: string) => Promise<LobbySummary>;
		sessions: SessionStore;
		/** The game's message for an error code; it also gets 'invalid_name', 'game_not_found' and 'network'. */
		errorText: (code: Code) => string;
	} = $props();

	// The codes the page raises itself; every game has them.
	const codeOf = (e: unknown) => (e instanceof ApiError ? e.code : 'network') as Code;
	const invalidName = 'invalid_name' as Code;
	const gameNotFound = 'game_not_found' as Code;

	let nickname = $state('');
	let code = $state('');
	let busy = $state(false);
	/** Number of remembered games; null until localStorage has been read. */
	let storedCount = $state<number | null>(null);
	let recent = $state<RecentGame[] | null>(null);
	let historyOpen = $state(false);

	const lists = $derived(partitionGames(recent ?? []));

	onMount(() => {
		nickname = sessions.getNickname();
		const stored = sessions.list().map((s) => ({ code: s.code, name: s.session.name }));
		storedCount = stored.length;
		if (stored.length === 0) return;
		void loadRecentGames(stored, getSummary, (code) => sessions.clear(code)).then((games) => {
			recent = games;
		});
	});

	const dateFormat = new Intl.DateTimeFormat(undefined, {
		day: 'numeric',
		month: 'short',
		hour: '2-digit',
		minute: '2-digit'
	});
	const when = (ms: number) => (ms > 0 ? dateFormat.format(ms) : '');
	const badgeVariant = (g: RecentGame) =>
		g.phase === 'won' ? 'default' : g.phase === 'lost' ? 'destructive' : 'secondary';

	const name = $derived(nickname.trim());

	async function enter(res: LobbySession) {
		sessions.setNickname(name);
		sessions.save(res.gameId, { token: res.token, playerId: res.playerId, name });
		await goto(gameHref(res.gameId));
	}

	async function run(action: () => Promise<LobbySession>) {
		if (busy) return;
		if (!name) {
			toast.error(errorText(invalidName));
			return;
		}
		busy = true;
		try {
			await enter(await action());
		} catch (e) {
			toast.error(errorText(codeOf(e)));
		} finally {
			busy = false;
		}
	}

	const host = () => run(() => createGame(name));
	const join = () => {
		if (code.length !== CODE_LENGTH) {
			toast.error(errorText(gameNotFound));
			return;
		}
		return run(() => joinGame(code, name));
	};
</script>

<div class="mx-auto flex w-full max-w-md flex-1 flex-col justify-center gap-4">
	<Card.Root class={heroClass}>
		{@render heroTop?.()}
		<Card.Header class={heroHeaderClass}>
			{@render hero()}
		</Card.Header>
		<Card.Content class="flex flex-col gap-5">
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

			<Button size="lg" class="h-12 text-base" disabled={busy} onclick={host}>Host a game</Button>

			<div class="flex items-center gap-3">
				<Separator class="flex-1" />
				<span class="text-sm text-muted-foreground">or</span>
				<Separator class="flex-1" />
			</div>

			<form
				class="flex flex-col gap-2"
				onsubmit={(e) => {
					e.preventDefault();
					join();
				}}
			>
				<Label for="code">Game code</Label>
				<div class="flex gap-2">
					<Input
						id="code"
						class="h-12 text-center font-mono text-xl tracking-widest uppercase"
						value={code}
						oninput={(e) => (code = normalizeCode(e.currentTarget.value))}
						maxlength={CODE_LENGTH}
						autocapitalize="characters"
						autocomplete="off"
						spellcheck={false}
						placeholder="K7MQ2"
					/>
					<Button
						type="submit"
						variant="secondary"
						size="lg"
						class="h-12 px-6 text-base"
						disabled={busy}
					>
						Join
					</Button>
				</div>
			</form>
		</Card.Content>
	</Card.Root>

	{#snippet gameRow(g: RecentGame)}
		{@const label = statusLabel(g, statusLabels)}
		<Button
			variant="outline"
			class="h-auto min-h-14 justify-between gap-3 py-2 text-base"
			href={gameHref(g.code)}
		>
			<span class="flex min-w-0 flex-col items-start gap-0.5">
				<span class="font-mono tracking-widest">{g.code}</span>
				<span class="max-w-full truncate text-xs font-normal text-muted-foreground">
					{[g.name, when(g.createdAt)].filter(Boolean).join(' · ')}
				</span>
			</span>
			{#if label}
				<Badge variant={badgeVariant(g)}>{label}</Badge>
			{/if}
		</Button>
	{/snippet}

	{#if storedCount !== null && storedCount > 0 && recent === null}
		<Card.Root>
			<Card.Header>
				<Card.Title>Continue</Card.Title>
			</Card.Header>
			<Card.Content class="flex flex-col gap-2">
				{#each [0, 1].slice(0, storedCount) as i (i)}
					<Skeleton class="h-14 w-full" />
				{/each}
			</Card.Content>
		</Card.Root>
	{:else if recent !== null}
		{#if lists.continuable.length > 0}
			<Card.Root>
				<Card.Header>
					<Card.Title>Continue</Card.Title>
				</Card.Header>
				<Card.Content class="flex flex-col gap-2">
					{#each lists.continuable as g (g.code)}
						{@render gameRow(g)}
					{/each}
				</Card.Content>
			</Card.Root>
		{/if}

		{#if lists.history.length > 0}
			<Card.Root>
				<Card.Header>
					<button
						type="button"
						class="flex min-h-10 w-full items-center justify-between gap-2 text-left"
						aria-expanded={historyOpen}
						aria-controls="history-list"
						onclick={() => (historyOpen = !historyOpen)}
					>
						<Card.Title
							>History <span class="text-muted-foreground">({lists.history.length})</span
							></Card.Title
						>
						<ChevronDown
							class="size-5 shrink-0 text-muted-foreground transition-transform {historyOpen
								? 'rotate-180'
								: ''}"
						/>
					</button>
				</Card.Header>
				{#if historyOpen}
					<Card.Content id="history-list" class="flex flex-col gap-2">
						{#each lists.history as g (g.code)}
							{@render gameRow(g)}
						{/each}
					</Card.Content>
				{/if}
			</Card.Root>
		{/if}
	{/if}
</div>
