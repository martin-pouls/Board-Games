<script lang="ts">
	import ArrowUp from '@lucide/svelte/icons/arrow-up';
	import ArrowDown from '@lucide/svelte/icons/arrow-down';
	import RotateCcw from '@lucide/svelte/icons/rotate-ccw';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import { commandInfo } from '#lib/game/commands.js';
	import { cn } from '#lib/utils.js';
	import type { Pile } from '#lib/types.js';
	import CommandIcon from './CommandIcon.svelte';

	/** Every card on a pile, in the order it was played (bottom first, top last). */
	let { pile, open = $bindable(false) }: { pile: Pile | null; open?: boolean } = $props();

	const up = $derived(pile?.direction === 'up');
	const steps = $derived(
		(pile?.cards ?? []).map((card, i, cards) => {
			const prev = i === 0 ? pile!.startValue : cards[i - 1].value;
			// Against the pile's direction is only legal as a backwards trick.
			return { card, trick: up ? card.value < prev : card.value > prev };
		})
	);
	const anyTrick = $derived(steps.some((s) => s.trick));
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="max-h-[85svh] overflow-y-auto sm:max-w-lg">
		{#if pile}
			<Dialog.Header>
				<Dialog.Title class="flex items-center gap-1.5">
					{#if up}<ArrowUp class="size-5" aria-hidden="true" />{:else}<ArrowDown
							class="size-5"
							aria-hidden="true"
						/>{/if}
					{up ? 'Ascending' : 'Descending'} pile
				</Dialog.Title>
				<Dialog.Description>
					{pile.cards.length}
					{pile.cards.length === 1 ? 'card' : 'cards'} in the order they were played. The top card is
					last.
				</Dialog.Description>
			</Dialog.Header>

			<ol class="flex flex-wrap gap-1.5" aria-label="Cards on this pile, bottom first">
				<li
					class={cn(
						'flex aspect-[2/3] w-11 flex-col items-center justify-center rounded-md border-2 border-dashed text-base font-bold tabular-nums opacity-60',
						up
							? 'border-sky-300 text-sky-700 dark:border-sky-800 dark:text-sky-300'
							: 'border-amber-300 text-amber-700 dark:border-amber-800 dark:text-amber-300'
					)}
					aria-label={`Start card ${pile.startValue}`}
				>
					{pile.startValue}
					<span class="text-[0.55rem] font-medium tracking-wide uppercase">Start</span>
				</li>
				{#each steps as { card, trick }, i (card.value)}
					{@const info = card.command ? commandInfo(card.command) : undefined}
					{@const top = i === steps.length - 1}
					<li
						class={cn(
							'relative flex aspect-[2/3] w-11 flex-col items-center justify-center gap-0.5 rounded-md border-2 bg-card text-base font-bold text-card-foreground tabular-nums shadow-xs',
							info?.kind === 'lightning' && 'border-red-400 dark:border-red-500',
							info?.kind === 'infinity' && 'border-violet-400 dark:border-violet-500',
							!info && 'border-border',
							top &&
								(up
									? 'ring-2 ring-sky-500 ring-offset-2 ring-offset-background'
									: 'ring-2 ring-amber-500 ring-offset-2 ring-offset-background')
						)}
						aria-label={`${card.value}${info ? `, ${info.label}` : ''}${trick ? ', backwards trick' : ''}${top ? ', top card' : ''}`}
						title={info ? `${info.label}: ${info.description}` : undefined}
					>
						<span class="leading-none">{card.value}</span>
						{#if card.command}
							<CommandIcon
								command={card.command}
								class={cn(
									'size-3.5',
									info?.kind === 'lightning'
										? 'text-red-600 dark:text-red-400'
										: 'text-violet-600 dark:text-violet-400'
								)}
							/>
						{/if}
						{#if trick}
							<span
								class="absolute -top-1.5 -right-1.5 flex size-4 items-center justify-center rounded-full bg-emerald-600 text-white shadow-sm"
								aria-hidden="true"
							>
								<RotateCcw class="size-2.5" strokeWidth={3} />
							</span>
						{/if}
					</li>
				{/each}
			</ol>

			{#if anyTrick}
				<p class="flex items-center gap-1.5 text-xs text-muted-foreground">
					<span
						class="flex size-4 items-center justify-center rounded-full bg-emerald-600 text-white"
						aria-hidden="true"><RotateCcw class="size-2.5" strokeWidth={3} /></span
					>
					Backwards trick
				</p>
			{/if}
		{/if}
	</Dialog.Content>
</Dialog.Root>
