<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { toast } from 'svelte-sonner';
	import { onMount } from 'svelte';
	import { ApiError, createGame, getSummary, joinGame } from '#lib/api.js';
	import { errorText } from '#lib/errors.js';
	import { normalizeCode, CODE_LENGTH } from '@board-games/shared/protocol.js';
	import {
		loadRecentGames,
		partitionGames,
		statusLabel,
		type RecentGame
	} from '#lib/game/recent.js';
	import { sessions } from '#lib/session.js';
	import ChevronDown from '@lucide/svelte/icons/chevron-down';
	import { Badge } from '#lib/components/ui/badge/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Card from '#lib/components/ui/card/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import { Separator } from '#lib/components/ui/separator/index.js';
	import { Skeleton } from '#lib/components/ui/skeleton/index.js';
	import type { SessionResponse } from '#lib/types.js';

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

	const name = $derived(nickname.trim());

	async function enter(res: SessionResponse) {
		sessions.setNickname(name);
		sessions.save(res.gameId, { token: res.token, playerId: res.playerId, name });
		await goto(resolve('/g/[code]', { code: res.gameId }));
	}

	async function run(action: () => Promise<SessionResponse>) {
		if (busy) return;
		if (!name) {
			toast.error(errorText('invalid_name'));
			return;
		}
		busy = true;
		try {
			await enter(await action());
		} catch (e) {
			toast.error(errorText(e instanceof ApiError ? e.code : 'network'));
		} finally {
			busy = false;
		}
	}

	const host = () => run(() => createGame(name));
	const join = () => {
		if (code.length !== CODE_LENGTH) {
			toast.error(errorText('game_not_found'));
			return;
		}
		return run(() => joinGame(code, name));
	};
</script>

<div class="mx-auto flex w-full max-w-md flex-1 flex-col justify-center gap-4">
	<Card.Root>
		<Card.Header>
			<Card.Title class="text-3xl">The Crew</Card.Title>
			<Card.Description
				>Reach Planet Nine together: win every trick you promised, in silence.</Card.Description
			>
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
		<Button
			variant="outline"
			class="h-auto min-h-14 justify-between gap-3 py-2 text-base"
			href={resolve('/g/[code]', { code: g.code })}
		>
			<span class="flex min-w-0 flex-col items-start gap-0.5">
				<span class="font-mono tracking-widest">{g.code}</span>
				<span class="max-w-full truncate text-xs font-normal text-muted-foreground">
					{[g.name, when(g.createdAt)].filter(Boolean).join(' · ')}
				</span>
			</span>
			{#if statusLabel(g)}
				<Badge variant="secondary">{statusLabel(g)}</Badge>
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
