<script lang="ts">
	import Hand from '@lucide/svelte/icons/hand';
	import Hourglass from '@lucide/svelte/icons/hourglass';
	import ThumbsUp from '@lucide/svelte/icons/thumbs-up';
	import { cn } from '#lib/utils.js';
	import type { PlayerSummary } from '#lib/types.js';

	let {
		players,
		online,
		class: className
	}: { players: PlayerSummary[]; online: string[]; class?: string } = $props();

	const ordered = $derived([...players].sort((a, b) => a.seat - b.seat));
	// Four chips only fit on one line without the online dots: offline players fade instead. The
	// active player's chip never shrinks; the others truncate their names.
	const compact = $derived(players.length >= 4);
</script>

<ul class={cn('flex min-w-0 gap-1.5', className)} aria-label="Players">
	{#each ordered as p (p.id)}
		{@const isOnline = online.includes(p.id)}
		<li
			class={cn(
				'flex min-w-0 items-center rounded-full border py-0.5 text-sm transition-colors',
				compact ? 'gap-1 px-1.5' : 'gap-1.5 px-2',
				p.isActive
					? 'shrink-0 border-primary bg-primary font-semibold text-primary-foreground shadow-sm'
					: 'bg-card',
				compact && !isOnline && 'opacity-60'
			)}
			aria-current={p.isActive ? 'true' : undefined}
			title={compact && !isOnline ? 'Offline' : undefined}
		>
			{#if compact}
				{#if !isOnline}<span class="sr-only">Offline:</span>{/if}
			{:else}
				<span
					class={cn(
						'size-2 shrink-0 rounded-full',
						isOnline ? 'bg-online' : 'bg-muted-foreground/40'
					)}
					title={isOnline ? 'Online' : 'Offline'}
				></span>
			{/if}
			<span class="max-w-[6.5rem] min-w-0 truncate">{p.name}</span>
			{#if p.yieldedLastTurn}<Hourglass class="size-3.5 shrink-0" aria-label="Yielded" />{/if}
			{#if p.signal === 'goodPlay'}
				<ThumbsUp class="size-3.5 shrink-0" aria-label="Has a good play" />
			{:else if p.signal === 'ratherNot'}
				<Hand class="size-3.5 shrink-0" aria-label="Would rather not go next" />
			{/if}
			<span
				class={cn(
					'min-w-5 shrink-0 rounded-full px-1.5 text-center text-xs font-semibold tabular-nums',
					p.isActive ? 'bg-primary-foreground/20' : 'bg-muted'
				)}
				title="Cards in hand">{p.handCount}</span
			>
		</li>
	{/each}
</ul>
