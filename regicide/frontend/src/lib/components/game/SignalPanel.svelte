<script lang="ts">
	import Hand from '@lucide/svelte/icons/hand';
	import ThumbsUp from '@lucide/svelte/icons/thumbs-up';
	import { Button } from '#lib/components/ui/button/index.js';
	import type { GameClient } from '#lib/game/client.js';
	import type { Runner } from '#lib/game/selection.js';
	import type { JesterSignal, PlayerView } from '#lib/types.js';
	import TurnBanner from './TurnBanner.svelte';

	let {
		client,
		view,
		busy,
		connected,
		run,
		turnKey
	}: {
		client: GameClient;
		view: PlayerView;
		busy: boolean;
		connected: boolean;
		run: Runner;
		turnKey: string;
	} = $props();

	const chooser = $derived(view.players.find((p) => p.id === view.turn?.activePlayerId));
	const current = $derived<JesterSignal>(view.players.find((p) => p.isYou)?.signal ?? 'none');

	const options: { value: JesterSignal; label: string; title: string }[] = [
		{ value: 'goodPlay', label: 'Good play', title: 'I have a good play' },
		{ value: 'ratherNot', label: 'Rather not', title: "I'd rather not go next" }
	];

	/** Tapping the active signal again withdraws it. */
	const toggle = (s: JesterSignal) => run(() => client.setSignal(current === s ? 'none' : s));
</script>

<TurnBanner title={`${chooser?.name ?? 'Someone'}'s turn`} mine={false} {turnKey}>
	<span class="text-xs text-muted-foreground">Picking who goes next</span>
</TurnBanner>
<div class="flex h-11 gap-2" role="group" aria-label="Signal">
	{#each options as o (o.value)}
		<Button
			variant={current === o.value ? 'default' : 'secondary'}
			class="h-11 flex-1 gap-1.5 text-base"
			aria-pressed={current === o.value}
			title={o.title}
			disabled={!connected || busy}
			onclick={() => toggle(o.value)}
		>
			{#if o.value === 'goodPlay'}<ThumbsUp />{:else}<Hand />{/if}
			{o.label}
		</Button>
	{/each}
</div>
