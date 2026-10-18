<script lang="ts">
	import { cn } from '#lib/utils.js';
	import type { PlayerSummary } from '#lib/types.js';

	let {
		players,
		online,
		showCounts = true
	}: { players: PlayerSummary[]; online: string[]; showCounts?: boolean } = $props();

	const ordered = $derived([...players].sort((a, b) => a.seat - b.seat));
</script>

<!--
	One row: names shorten so the strip never wraps (five players may scroll sideways) and the piles
	keep their size. In game the host doesn't matter, so no crown; your own seat says "You".
-->
<ul class="flex [scrollbar-width:none] gap-1.5 overflow-x-auto py-0.5" aria-label="Players">
	{#each ordered as p (p.id)}
		{@const isOnline = online.includes(p.id)}
		<li
			class={cn(
				'flex min-w-0 items-center gap-1 rounded-full border px-2 py-1 text-sm transition-colors',
				p.isActive
					? 'border-transparent bg-go-soft font-semibold shadow-[inset_0_0_0_1.5px_var(--go)]'
					: 'bg-card'
			)}
			aria-current={p.isActive ? 'true' : undefined}
		>
			<span
				class={cn('size-2 shrink-0 rounded-full', isOnline ? 'bg-go' : 'bg-muted-foreground/40')}
				title={isOnline ? 'Online' : 'Offline'}
			></span>
			<span class="max-w-[6.5rem] min-w-[3ch] truncate">{p.isYou ? 'You' : p.name}</span>
			{#if showCounts}
				<span
					class={cn(
						'min-w-5 shrink-0 rounded-full px-1.5 text-center text-xs font-semibold tabular-nums',
						'bg-muted'
					)}
					title="Cards in hand">{p.handCount}</span
				>
			{/if}
		</li>
	{/each}
</ul>
