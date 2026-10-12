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
</script>

<ul class={cn('flex flex-wrap gap-1.5', className)} aria-label="Players">
	{#each ordered as p (p.id)}
		{@const isOnline = online.includes(p.id)}
		<li
			class={cn(
				'flex shrink-0 items-center gap-1.5 rounded-full border px-2 py-0.5 text-sm transition-colors',
				p.isActive
					? 'border-primary bg-primary font-semibold text-primary-foreground shadow-sm'
					: 'bg-card'
			)}
			aria-current={p.isActive ? 'true' : undefined}
		>
			<span
				class={cn(
					'size-2 shrink-0 rounded-full',
					isOnline ? 'bg-emerald-500' : 'bg-muted-foreground/40'
				)}
				title={isOnline ? 'Online' : 'Offline'}
			></span>
			<span class="max-w-[6.5rem] truncate">{p.name}</span>
			{#if p.yieldedLastTurn}<Hourglass class="size-3.5 shrink-0" aria-label="Yielded" />{/if}
			{#if p.signal === 'goodPlay'}
				<ThumbsUp class="size-3.5 shrink-0" aria-label="Has a good play" />
			{:else if p.signal === 'ratherNot'}
				<Hand class="size-3.5 shrink-0" aria-label="Would rather not go next" />
			{/if}
			<span
				class={cn(
					'min-w-5 rounded-full px-1.5 text-center text-xs font-semibold tabular-nums',
					p.isActive ? 'bg-primary-foreground/20' : 'bg-muted'
				)}
				title="Cards in hand">{p.handCount}</span
			>
		</li>
	{/each}
</ul>
