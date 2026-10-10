<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { toast } from 'svelte-sonner';
	import { onMount } from 'svelte';
	import { ApiError, createGame, joinGame } from '#lib/api.js';
	import { errorText } from '#lib/errors.js';
	import { normalizeCode, CODE_LENGTH } from '#lib/game/protocol.js';
	import { sessions } from '#lib/session.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Card from '#lib/components/ui/card/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import { Separator } from '#lib/components/ui/separator/index.js';
	import type { SessionResponse } from '#lib/types.js';

	let nickname = $state('');
	let code = $state('');
	let busy = $state(false);
	let stored = $state<{ code: string; name: string }[]>([]);

	onMount(() => {
		nickname = sessions.getNickname();
		stored = sessions.list().map((s) => ({ code: s.code, name: s.session.name }));
	});

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
			<Card.Title class="text-3xl">The Game</Card.Title>
			<Card.Description>Play as a team, as long as you can.</Card.Description>
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

	{#if stored.length > 0}
		<Card.Root>
			<Card.Header>
				<Card.Title>Continue</Card.Title>
			</Card.Header>
			<Card.Content class="flex flex-col gap-2">
				{#each stored as s (s.code)}
					<Button
						variant="outline"
						class="h-12 justify-between text-base"
						href={resolve('/g/[code]', { code: s.code })}
					>
						<span class="font-mono tracking-widest">{s.code}</span>
						<span class="text-muted-foreground">{s.name}</span>
					</Button>
				{/each}
			</Card.Content>
		</Card.Root>
	{/if}
</div>
