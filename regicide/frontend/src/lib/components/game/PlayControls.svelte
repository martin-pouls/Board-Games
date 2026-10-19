<script lang="ts">
	import { Button } from '#lib/components/ui/button/index.js';
	import type { GameClient } from '#lib/game/client.js';
	import { previewPlay } from '#lib/game/rules.js';
	import type { Runner } from '#lib/game/selection.js';
	import type { Card, PlayerView } from '#lib/types.js';
	import PlayPreview from './PlayPreview.svelte';
	import SoloJesterButton from './SoloJesterButton.svelte';
	import TurnBanner from './TurnBanner.svelte';

	let {
		client,
		view,
		selected,
		busy,
		connected,
		run,
		turnKey
	}: {
		client: GameClient;
		view: PlayerView;
		selected: Card[];
		busy: boolean;
		connected: boolean;
		run: Runner;
		turnKey: string;
	} = $props();

	const preview = $derived(previewPlay(view, selected));
	const canPlay = $derived(connected && !busy && !!preview && preview.check.ok);
	const canYield = $derived(connected && !busy && !!view.turn?.canYield);
	const showPreview = $derived(!!preview && selected.length > 0);
	// Five badges don't fit next to "Your turn"; the outline still says whose turn it is.
	const badges = $derived(
		preview?.check.ok
			? preview.effects.length +
					(preview.defeats ? 1 : 0) +
					(preview.check.kind === 'jester' ? 1 : 0)
			: 0
	);
</script>

<TurnBanner title="Your turn" mine {turnKey} hideTitle={showPreview && badges >= 5}>
	{#if showPreview && preview}
		<PlayPreview {preview} />
	{:else}
		<span class="text-xs text-muted-foreground">
			{view.turn?.canYield ? 'Select cards to play' : "Select cards: you can't yield"}
		</span>
	{/if}
</TurnBanner>
<div class="flex h-11 gap-2">
	<SoloJesterButton {client} {view} {busy} {connected} {run} />
	<Button
		variant="secondary"
		class="h-11 px-5 text-base"
		disabled={!canYield}
		title={view.turn?.canYield ? 'Skip to taking damage' : 'Everyone else yielded last turn'}
		onclick={() => run(() => client.yieldTurn())}
	>
		Yield
	</Button>
	<Button
		class="h-11 flex-1 text-base"
		disabled={!canPlay}
		onclick={() => run(() => client.play(selected))}
	>
		Play
	</Button>
</div>
