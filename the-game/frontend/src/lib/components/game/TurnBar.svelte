<script lang="ts">
	import { fly } from 'svelte/transition';
	import { prefersReducedMotion } from 'svelte/motion';
	import { Button } from '#lib/components/ui/button/index.js';
	import { cn } from '#lib/utils.js';
	import type { PlayerView } from '#lib/types.js';

	let {
		view,
		isMyTurn,
		disabled = false,
		pending = false,
		onendturn
	}: {
		view: PlayerView;
		isMyTurn: boolean;
		disabled?: boolean;
		pending?: boolean;
		onendturn?: () => void;
	} = $props();

	const turn = $derived(view.turn);
	const active = $derived(view.players.find((p) => p.id === turn?.activePlayerId));
	const dots = $derived(
		turn
			? Array.from({ length: Math.max(turn.requiredMinimum, turn.cardsPlayed) }, (_, i) => i)
			: []
	);
</script>

{#if turn}
	{#key turn.activePlayerId + ':' + turn.number}
		<div
			in:fly={{ y: 8, duration: prefersReducedMotion.current ? 0 : 200 }}
			class={cn(
				'flex min-h-14 items-center gap-3 rounded-xl px-3 py-1.5',
				isMyTurn ? 'bg-emerald-500/10 ring-1 ring-emerald-500/40' : 'bg-muted/60'
			)}
			aria-live="polite"
		>
			<div class="min-w-0 flex-1">
				{#if isMyTurn}
					<p class="font-semibold">Your turn</p>
					<p class="flex items-center gap-1.5 text-sm text-muted-foreground">
						<span class="flex gap-0.5" aria-hidden="true">
							{#each dots as i (i)}
								<span
									class={cn(
										'size-2 rounded-full',
										i < turn.cardsPlayed ? 'bg-emerald-500' : 'bg-muted-foreground/30'
									)}
								></span>
							{/each}
						</span>
						<span class="tabular-nums">{turn.cardsPlayed}/{turn.requiredMinimum} played</span>
					</p>
				{:else}
					<p class="truncate font-semibold">{active?.name ?? 'Someone'}'s turn</p>
					<p class="text-sm text-muted-foreground tabular-nums">{turn.cardsPlayed} played</p>
				{/if}
			</div>
			{#if isMyTurn}
				<Button
					size="lg"
					class="h-12 min-w-32 px-6 text-base"
					disabled={disabled || pending || !turn.canEndTurn}
					onclick={onendturn}
				>
					End turn
				</Button>
			{/if}
		</div>
	{/key}
{/if}
