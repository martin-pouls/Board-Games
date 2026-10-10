<script lang="ts">
	import Drama from '@lucide/svelte/icons/drama';
	import Eraser from '@lucide/svelte/icons/eraser';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import type { GameClient } from '#lib/game/client.js';
	import { discardProblemText } from '#lib/game/format.js';
	import { checkDiscard, handValue } from '#lib/game/rules.js';
	import type { Runner } from '#lib/game/selection.js';
	import { cn } from '#lib/utils.js';
	import type { Card, PlayerView } from '#lib/types.js';

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

	const required = $derived(view.turn?.damageToCover ?? 0);
	const check = $derived(checkDiscard(selected, required));
	const sum = $derived(check.sum);
	const fill = $derived(required > 0 ? Math.min(100, (sum / required) * 100) : 100);
	const handShort = $derived(handValue(view.hand) < required);
	const status = $derived(
		selected.length === 0
			? `Select cards worth at least ${required}.`
			: discardProblemText(check, required, selected)
	);
</script>

<div class="flex flex-col gap-2">
	<div class="flex items-baseline justify-between">
		<h2 class="text-base font-semibold">Take {required} damage</h2>
		<span class={cn('text-sm tabular-nums', check.ok && 'font-semibold text-primary')}>
			{sum} / {required}
		</span>
	</div>
	<div class="h-2 overflow-hidden rounded-full bg-muted" aria-hidden="true">
		<div
			class={cn('h-full transition-[width]', check.ok ? 'bg-primary' : 'bg-health')}
			style="width: {fill}%"
		></div>
	</div>
	<p class="min-h-5 text-sm text-muted-foreground" aria-live="polite">{status}</p>
	{#if handShort && view.soloJestersLeft > 0}
		<p class="rounded-md bg-jester/15 px-2 py-1 text-sm text-jester">
			Your hand can't cover this: use a Jester.
		</p>
	{/if}
	<div class="flex gap-2">
		<Button
			class="h-12 flex-1 text-base"
			disabled={!connected || busy || !check.ok}
			onclick={() => run(() => client.discard(selected))}
		>
			Discard
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
