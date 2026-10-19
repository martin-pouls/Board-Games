<!--
	The action area above the hand, split like The Game's: the turn bar, then the turn's buttons,
	or who we're waiting for (plain text, no spinner).
-->
<script lang="ts">
	import type { GameClient } from '#lib/game/client.js';
	import type { Runner } from '#lib/game/selection.js';
	import type { Card, PlayerView } from '#lib/types.js';
	import ChooseNextPanel from './ChooseNextPanel.svelte';
	import DiscardControls from './DiscardControls.svelte';
	import PlayControls from './PlayControls.svelte';
	import SignalPanel from './SignalPanel.svelte';
	import TurnBanner from './TurnBanner.svelte';

	let {
		client,
		view,
		selected,
		busy,
		connected,
		run
	}: {
		client: GameClient;
		view: PlayerView;
		selected: Card[];
		busy: boolean;
		connected: boolean;
		run: Runner;
	} = $props();

	const step = $derived(view.turn?.step ?? '');
	const mine = $derived(view.turn?.activePlayerId === view.you);
	const activeName = $derived(
		view.players.find((p) => p.id === view.turn?.activePlayerId)?.name ?? 'Someone'
	);
	const turnKey = $derived(`${view.turn?.activePlayerId}:${view.turn?.number}`);
</script>

<div class="flex shrink-0 flex-col gap-1.5">
	{#if step === 'play' && mine}
		<PlayControls {client} {view} {selected} {busy} {connected} {run} {turnKey} />
	{:else if step === 'discard' && mine}
		<DiscardControls {client} {view} {selected} {busy} {connected} {run} {turnKey} />
	{:else if step === 'chooseNext' && mine}
		<ChooseNextPanel {client} {view} {busy} {connected} {run} {turnKey} />
	{:else if step === 'chooseNext'}
		<SignalPanel {client} {view} {busy} {connected} {run} {turnKey} />
	{:else}
		<TurnBanner title={`${activeName}'s turn`} mine={false} {turnKey}>
			{#if step === 'discard'}
				<span class="text-xs text-muted-foreground tabular-nums">
					Takes {view.turn?.damageToCover ?? 0} damage
				</span>
			{/if}
		</TurnBanner>
		<p class="flex h-11 items-center justify-center text-sm text-muted-foreground">
			<span class="truncate">
				Waiting for {activeName} to {step === 'discard' ? 'discard' : 'play'}.
			</span>
		</p>
	{/if}
</div>
