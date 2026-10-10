<script lang="ts">
	import LoaderCircle from '@lucide/svelte/icons/loader-circle';
	import type { GameClient } from '#lib/game/client.js';
	import type { Runner } from '#lib/game/selection.js';
	import type { Card, PlayerView } from '#lib/types.js';
	import ChooseNextPanel from './ChooseNextPanel.svelte';
	import DiscardControls from './DiscardControls.svelte';
	import PlayControls from './PlayControls.svelte';
	import SignalPanel from './SignalPanel.svelte';

	let {
		client,
		view,
		selected,
		busy,
		connected,
		run,
		onclear
	}: {
		client: GameClient;
		view: PlayerView;
		selected: Card[];
		busy: boolean;
		connected: boolean;
		run: Runner;
		onclear: () => void;
	} = $props();

	const step = $derived(view.turn?.step ?? '');
	const mine = $derived(view.turn?.activePlayerId === view.you);
	const activeName = $derived(
		view.players.find((p) => p.id === view.turn?.activePlayerId)?.name ?? 'Someone'
	);
	const waiting = $derived(
		step === 'discard'
			? `${activeName} must take ${view.turn?.damageToCover ?? 0} damage`
			: `${activeName} is choosing cards…`
	);
</script>

<div class="rounded-xl border bg-card/90 p-3">
	{#if step === 'play' && mine}
		<PlayControls {client} {view} {selected} {busy} {connected} {run} {onclear} />
	{:else if step === 'discard' && mine}
		<DiscardControls {client} {view} {selected} {busy} {connected} {run} {onclear} />
	{:else if step === 'chooseNext' && mine}
		<ChooseNextPanel {client} {view} {busy} {connected} {run} />
	{:else if step === 'chooseNext'}
		<SignalPanel {client} {view} {busy} {connected} {run} />
	{:else}
		<p class="flex min-h-12 items-center justify-center gap-2 text-muted-foreground">
			<LoaderCircle class="size-4 animate-spin" />
			{waiting}
		</p>
	{/if}
</div>
