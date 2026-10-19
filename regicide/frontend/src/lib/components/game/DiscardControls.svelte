<script lang="ts">
	import { Button } from '#lib/components/ui/button/index.js';
	import type { GameClient } from '#lib/game/client.js';
	import { discardProblemText } from '#lib/game/format.js';
	import { checkDiscard, handValue } from '#lib/game/rules.js';
	import type { Runner } from '#lib/game/selection.js';
	import { cn } from '#lib/utils.js';
	import type { Card, PlayerView } from '#lib/types.js';
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

	const required = $derived(view.turn?.damageToCover ?? 0);
	const check = $derived(checkDiscard(selected, required));
	const handShort = $derived(handValue(view.hand) < required && view.soloJestersLeft > 0);
</script>

<TurnBanner
	title={handShort ? `Can't cover ${required}` : `Discard ${required} or more`}
	mine
	{turnKey}
>
	{#if handShort}
		<span class="text-xs font-semibold text-jester">Use a Jester</span>
	{:else if !check.ok && check.problem === 'superfluous'}
		<span class="truncate text-xs text-muted-foreground">
			{discardProblemText(check, required, selected)}
		</span>
	{:else}
		<span
			class={cn(
				'text-xs tabular-nums',
				check.ok ? 'font-bold text-foreground' : 'text-muted-foreground'
			)}
		>
			Selected {check.sum} of {required}
		</span>
	{/if}
</TurnBanner>
<div class="flex h-11 gap-2">
	<SoloJesterButton {client} {view} {busy} {connected} {run} />
	<Button
		class="h-11 flex-1 text-base"
		disabled={!connected || busy || !check.ok}
		onclick={() => run(() => client.discard(selected))}
	>
		Discard
	</Button>
</div>
