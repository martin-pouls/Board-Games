<script lang="ts">
	import { Button } from '#lib/components/ui/button/index.js';
	import type { GameClient } from '#lib/game/client.js';
	import { previewPlay } from '#lib/game/rules.js';
	import type { Runner } from '#lib/game/selection.js';
	import type { Card, PlayerView } from '#lib/types.js';
	import PlayPreview from './PlayPreview.svelte';
	import SoloJesterButton from './SoloJesterButton.svelte';

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

	const preview = $derived(previewPlay(view, selected));
	const canPlay = $derived(connected && !busy && !!preview && preview.check.ok);
	const canYield = $derived(connected && !busy && !!view.turn?.canYield);
</script>

<div class="flex items-center gap-2">
	<div class="min-w-0 flex-1" aria-live="polite">
		{#if preview && selected.length > 0}
			<PlayPreview {preview} />
		{:else}
			<p class="text-sm text-muted-foreground">
				{view.turn?.canYield ? 'Your turn' : 'Your turn: no yielding'}
			</p>
		{/if}
	</div>
	<SoloJesterButton {client} {view} {busy} {connected} {run} />
	<Button
		class="h-9 px-4 text-base"
		disabled={!canPlay}
		onclick={() => run(() => client.play(selected))}
	>
		Play
	</Button>
	<Button
		variant="secondary"
		class="h-9 px-4 text-base"
		disabled={!canYield}
		title={view.turn?.canYield ? 'Skip to taking damage' : 'Everyone else yielded last turn'}
		onclick={() => run(() => client.yieldTurn())}
	>
		Yield
	</Button>
</div>
