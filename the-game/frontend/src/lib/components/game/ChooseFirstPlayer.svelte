<script lang="ts">
	import { toast } from 'svelte-sonner';
	import LoaderCircle from '@lucide/svelte/icons/loader-circle';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import { errorText } from '#lib/errors.js';
	import { ActionError } from '#lib/game/connection.svelte.js';
	import type { GameClient } from '#lib/game/client.js';
	import type { PlayerSummary, PlayerView } from '#lib/types.js';
	import Board from './Board.svelte';

	let { client, view }: { client: GameClient; view: PlayerView } = $props();

	let candidate = $state<PlayerSummary | null>(null);
	let open = $state(false);
	let busy = $state(false);

	const isHost = $derived(view.hostId === view.you);
	const host = $derived(view.players.find((p) => p.id === view.hostId));
	const players = $derived([...view.players].sort((a, b) => a.seat - b.seat));
	const connected = $derived(client.status === 'open');

	function pick(p: PlayerSummary) {
		candidate = p;
		open = true;
	}

	async function confirm() {
		if (!candidate || busy) return;
		busy = true;
		try {
			await client.chooseFirstPlayer(candidate.id);
			open = false;
		} catch (e) {
			toast.error(errorText(e instanceof ActionError ? e.code : 'internal'));
		} finally {
			busy = false;
		}
	}
</script>

<Board {client} {view} readonly>
	{#snippet controls()}
		{#if isHost}
			<div class="flex flex-col gap-2 rounded-xl bg-muted/60 p-3">
				<p class="text-sm font-medium">Look at your cards, discuss, then choose who starts.</p>
				<div class="grid grid-cols-2 gap-2">
					{#each players as p (p.id)}
						<Button
							variant="outline"
							class="h-11 min-w-0 text-base"
							disabled={!connected || busy}
							onclick={() => pick(p)}
						>
							<span class="truncate">{p.isYou ? `${p.name} (you)` : p.name}</span>
						</Button>
					{/each}
				</div>
			</div>
		{:else}
			<p
				class="flex min-h-14 items-center justify-center gap-2 rounded-xl bg-muted/60 px-3 text-center text-sm text-muted-foreground"
			>
				<LoaderCircle class="size-4 shrink-0 animate-spin" />
				Look at your cards and discuss. Waiting for {host?.name ?? 'the host'} to choose who starts.
			</p>
		{/if}
	{/snippet}
</Board>

<AlertDialog.Root bind:open>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>
				{candidate?.isYou ? 'You start?' : `${candidate?.name ?? ''} starts?`}
			</AlertDialog.Title>
			<AlertDialog.Description>Play then continues in seat order.</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={busy}>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action disabled={busy || !connected} onclick={confirm}>Start</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
