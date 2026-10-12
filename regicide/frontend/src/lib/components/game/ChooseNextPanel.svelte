<script lang="ts">
	import Hand from '@lucide/svelte/icons/hand';
	import Hourglass from '@lucide/svelte/icons/hourglass';
	import ThumbsUp from '@lucide/svelte/icons/thumbs-up';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import type { GameClient } from '#lib/game/client.js';
	import { canPlayerYield } from '#lib/game/rules.js';
	import type { Runner } from '#lib/game/selection.js';
	import { cn } from '#lib/utils.js';
	import type { PlayerSummary, PlayerView } from '#lib/types.js';

	let {
		client,
		view,
		busy,
		connected,
		run
	}: { client: GameClient; view: PlayerView; busy: boolean; connected: boolean; run: Runner } =
		$props();

	let target = $state<PlayerSummary | null>(null);
	let open = $state(false);

	const players = $derived([...view.players].sort((a, b) => a.seat - b.seat));
	const stuck = (p: PlayerSummary) => p.handCount === 0 && !canPlayerYield(view.players, p.id);

	function choose(p: PlayerSummary) {
		target = p;
		open = true;
	}
</script>

<div class="flex flex-wrap items-center gap-2">
	<p class="text-sm text-muted-foreground">Next:</p>
	<ul class="flex flex-1 flex-wrap justify-end gap-1.5">
		{#each players as p (p.id)}
			<li>
				<Button
					variant="secondary"
					class={cn('h-9 gap-1.5 px-3', stuck(p) && 'text-health')}
					title={stuck(p) ? 'No cards and cannot yield: the game would be lost' : undefined}
					disabled={!connected || busy}
					onclick={() => choose(p)}
				>
					<span class="max-w-[6.5rem] truncate">{p.isYou ? 'Me' : p.name}</span>
					<span class="text-muted-foreground tabular-nums">· {p.handCount}</span>
					{#if p.yieldedLastTurn}<Hourglass aria-label="Yielded" />{/if}
					{#if p.signal === 'goodPlay'}
						<ThumbsUp class="text-primary" aria-label="Has a good play" />
					{:else if p.signal === 'ratherNot'}
						<Hand class="text-health" aria-label="Would rather not go next" />
					{/if}
				</Button>
			</li>
		{/each}
	</ul>
</div>

<AlertDialog.Root bind:open>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>
				{target?.isYou ? 'You go next?' : `${target?.name} goes next?`}
			</AlertDialog.Title>
			<AlertDialog.Description>
				{#if target && stuck(target)}
					{target.isYou ? 'You have' : `${target.name} has`} no cards and cannot yield. Everyone loses.
				{:else}
					The Jester player picks who takes the next turn.
				{/if}
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action
				onclick={() => {
					const t = target;
					open = false;
					if (t) run(() => client.chooseNext(t.id));
				}}
			>
				Confirm
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
