<script lang="ts">
	import Hand from '@lucide/svelte/icons/hand';
	import ThumbsUp from '@lucide/svelte/icons/thumbs-up';
	import { Button } from '#lib/components/ui/button/index.js';
	import type { GameClient } from '#lib/game/client.js';
	import type { Runner } from '#lib/game/selection.js';
	import type { JesterSignal, PlayerView } from '#lib/types.js';

	let {
		client,
		view,
		busy,
		connected,
		run
	}: { client: GameClient; view: PlayerView; busy: boolean; connected: boolean; run: Runner } =
		$props();

	const chooser = $derived(view.players.find((p) => p.id === view.turn?.activePlayerId));
	const current = $derived<JesterSignal>(view.players.find((p) => p.isYou)?.signal ?? 'none');

	const options: { value: JesterSignal; label: string }[] = [
		{ value: 'goodPlay', label: 'I have a good play' },
		{ value: 'ratherNot', label: "I'd rather not go next" },
		{ value: 'none', label: 'No signal' }
	];
</script>

<div class="flex flex-col gap-2">
	<p class="text-sm text-muted-foreground">
		{chooser?.name ?? 'Someone'} is choosing who goes next. You may signal:
	</p>
	<div class="flex flex-col gap-1.5" role="group" aria-label="Signal">
		{#each options as o (o.value)}
			<Button
				variant={current === o.value ? 'default' : 'secondary'}
				class="h-11 justify-start gap-2"
				aria-pressed={current === o.value}
				disabled={!connected || busy}
				onclick={() => run(() => client.setSignal(o.value))}
			>
				{#if o.value === 'goodPlay'}<ThumbsUp />{:else if o.value === 'ratherNot'}<Hand />{/if}
				{o.label}
			</Button>
		{/each}
	</div>
</div>
