<script lang="ts">
	import Drama from '@lucide/svelte/icons/drama';
	import Eraser from '@lucide/svelte/icons/eraser';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import type { GameClient } from '#lib/game/client.js';
	import { previewPlay } from '#lib/game/rules.js';
	import type { Runner } from '#lib/game/selection.js';
	import type { Card, PlayerView } from '#lib/types.js';
	import PlayPreview from './PlayPreview.svelte';

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

	let confirmJester = $state(false);

	const preview = $derived(previewPlay(view, selected));
	const canPlay = $derived(connected && !busy && !!preview && preview.check.ok);
	const canYield = $derived(connected && !busy && !!view.turn?.canYield);
</script>

<div class="flex flex-col gap-2">
	{#if preview && view.enemy}
		<PlayPreview {preview} enemy={view.enemy} />
	{/if}
	<div class="flex gap-2">
		<Button
			class="h-12 flex-1 text-base"
			disabled={!canPlay}
			onclick={() => run(() => client.play(selected))}
		>
			Play
		</Button>
		<Button
			variant="secondary"
			class="h-12 flex-1 text-base"
			disabled={!canYield}
			title={view.turn?.canYield ? 'Skip to taking damage' : 'Everyone else yielded last turn'}
			onclick={() => run(() => client.yieldTurn())}
		>
			Yield
		</Button>
		<Button
			variant="ghost"
			size="icon"
			class="size-12"
			disabled={selected.length === 0 || busy}
			aria-label="Clear selection"
			title="Clear selection"
			onclick={onclear}
		>
			<Eraser />
		</Button>
	</div>
	{#if !view.turn?.canYield}
		<p class="text-center text-xs text-muted-foreground">Everyone else yielded last turn.</p>
	{/if}
	{#if view.soloJestersLeft > 0}
		<Button
			variant="outline"
			class="h-10 border-jester/60 text-jester"
			disabled={!connected || busy}
			onclick={() => (confirmJester = true)}
		>
			<Drama />
			Use Jester ({view.soloJestersLeft} left)
		</Button>
	{/if}
</div>

<AlertDialog.Root bind:open={confirmJester}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Use a Jester power?</AlertDialog.Title>
			<AlertDialog.Description>
				Discard your hand and draw up to {view.rules.maxHandSize} cards.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action
				onclick={() => {
					confirmJester = false;
					run(() => client.useSoloJester());
				}}
			>
				Use Jester
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
