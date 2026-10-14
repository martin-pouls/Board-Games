<!--
	The header's sort button: an icon that opens a menu of the hand orders, each previewed with the
	same example hand. Display only; the parent stores the choice.
-->
<script lang="ts">
	import ArrowDownUp from '@lucide/svelte/icons/arrow-down-up';
	import { buttonVariants } from '#lib/components/ui/button/index.js';
	import * as DropdownMenu from '#lib/components/ui/dropdown-menu/index.js';
	import { cardLabel, isRed } from '#lib/game/format.js';
	import { handKeys, HAND_ORDERS, sortHand, type HandOrder } from '#lib/game/rules.js';
	import { cn } from '#lib/utils.js';
	import type { Card } from '#lib/types.js';

	let {
		order,
		onchange,
		anchor,
		class: className
	}: {
		order: HandOrder;
		onchange: (order: HandOrder) => void;
		/** The menu's right edge lines up with this element's (the header's icon row), not with the icon. */
		anchor?: HTMLElement;
		class?: string;
	} = $props();

	const NAMES: Record<HandOrder, string> = {
		rank: 'Rank, ace low',
		rankAceHigh: 'Rank, ace high',
		suit: 'Suit, then rank, ace low',
		suitAceHigh: 'Suit, then rank, ace high'
	};

	// Two Aces in different suits, so every order looks different.
	const EXAMPLE: Card[] = [
		{ suit: 'hearts', rank: 1 },
		{ suit: 'hearts', rank: 3 },
		{ suit: 'hearts', rank: 13 },
		{ suit: 'clubs', rank: 7 },
		{ suit: 'spades', rank: 1 },
		{ suit: 'spades', rank: 9 }
	];

	function previewClass(c: Card): string {
		if (c.suit === '') return 'text-jester';
		return isRed(c.suit) ? 'text-suit-red' : 'text-foreground';
	}
</script>

<DropdownMenu.Root>
	<DropdownMenu.Trigger
		class={cn(
			buttonVariants({ variant: 'ghost', size: 'icon-sm' }),
			'text-muted-foreground',
			className
		)}
		aria-label="Sort hand (now: {NAMES[order].toLowerCase()})"
		title="Sort hand"
	>
		<ArrowDownUp class="size-4" />
	</DropdownMenu.Trigger>
	<DropdownMenu.Content align="end" customAnchor={anchor} collisionPadding={8} class="w-72">
		<DropdownMenu.Label class="font-display tracking-wider">Sort hand</DropdownMenu.Label>
		<DropdownMenu.RadioGroup value={order} onValueChange={(v) => onchange(v as HandOrder)}>
			{#each HAND_ORDERS as o (o)}
				{@const cards = sortHand(EXAMPLE, o)}
				{@const keys = handKeys(cards)}
				<DropdownMenu.RadioItem value={o} closeOnSelect class="min-h-11">
					<span class="flex flex-col gap-0.5">
						<span class="text-base">{NAMES[o]}</span>
						<span
							class="flex flex-wrap gap-x-1.5 text-xs tabular-nums"
							style="font-family: system-ui, sans-serif"
							aria-hidden="true"
						>
							{#each cards as c, i (keys[i])}
								<span class={previewClass(c)}>{cardLabel(c)}</span>
							{/each}
						</span>
					</span>
				</DropdownMenu.RadioItem>
			{/each}
		</DropdownMenu.RadioGroup>
	</DropdownMenu.Content>
</DropdownMenu.Root>
