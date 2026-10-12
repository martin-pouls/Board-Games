<script lang="ts">
	import { Button } from '#lib/components/ui/button/index.js';
	import type { GameClient } from '#lib/game/client.js';
	import { discardProblemText } from '#lib/game/format.js';
	import { checkDiscard, handValue } from '#lib/game/rules.js';
	import type { Runner } from '#lib/game/selection.js';
	import { cn } from '#lib/utils.js';
	import type { Card, PlayerView } from '#lib/types.js';
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

	const required = $derived(view.turn?.damageToCover ?? 0);
	const check = $derived(checkDiscard(selected, required));
	const fill = $derived(required > 0 ? Math.min(100, (check.sum / required) * 100) : 100);
	const handShort = $derived(handValue(view.hand) < required && view.soloJestersLeft > 0);
	const label = $derived.by(() => {
		if (handShort) return "Can't cover: use a Jester";
		if (!check.ok && check.problem === 'superfluous') {
			return discardProblemText(check, required, selected);
		}
		return `Take ${required}`;
	});
</script>

<div class="flex items-center gap-2">
	<div class="flex min-w-0 flex-1 flex-col gap-1" aria-live="polite">
		<div class="flex items-baseline justify-between gap-2 text-sm">
			<span class={cn('truncate', handShort && 'text-jester')} title={label}>{label}</span>
			<span class={cn('tabular-nums', check.ok && 'font-semibold text-primary')}>
				{check.sum} / {required}
			</span>
		</div>
		<div class="h-1.5 overflow-hidden rounded-full bg-muted" aria-hidden="true">
			<div
				class={cn('h-full transition-[width]', check.ok ? 'bg-primary' : 'bg-health')}
				style="width: {fill}%"
			></div>
		</div>
	</div>
	<SoloJesterButton {client} {view} {busy} {connected} {run} />
	<Button
		class="h-9 px-4 text-base"
		disabled={!connected || busy || !check.ok}
		onclick={() => run(() => client.discard(selected))}
	>
		Discard
	</Button>
</div>
