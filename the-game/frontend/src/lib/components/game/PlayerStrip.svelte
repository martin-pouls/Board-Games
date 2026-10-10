<script lang="ts">
	import Crown from '@lucide/svelte/icons/crown';
	import { cn } from '#lib/utils.js';
	import type { PlayerSummary } from '#lib/types.js';

	let {
		players,
		online,
		showCounts = true
	}: { players: PlayerSummary[]; online: string[]; showCounts?: boolean } = $props();

	const ordered = $derived([...players].sort((a, b) => a.seat - b.seat));
</script>

<ul class="flex flex-wrap gap-1.5 py-0.5" aria-label="Players">
	{#each ordered as p (p.id)}
		{@const isOnline = online.includes(p.id)}
		<li
			class={cn(
				'flex shrink-0 items-center gap-1.5 rounded-full border px-2.5 py-1 text-sm transition-colors',
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
			{#if p.isHost}<Crown class="size-3.5 shrink-0 text-amber-500" aria-label="Host" />{/if}
			<span class="max-w-[6.5rem] truncate">{p.name}</span>
			{#if p.isYou}<span class="text-xs opacity-70">(you)</span>{/if}
			{#if showCounts}
				<span
					class={cn(
						'min-w-5 rounded-full px-1.5 text-center text-xs font-semibold tabular-nums',
						p.isActive ? 'bg-primary-foreground/20' : 'bg-muted'
					)}
					title="Cards in hand">{p.handCount}</span
				>
			{/if}
		</li>
	{/each}
</ul>
