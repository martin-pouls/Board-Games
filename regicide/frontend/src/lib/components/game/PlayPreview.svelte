<script lang="ts">
	import {
		cardName,
		playKindLabel,
		playProblemText,
		suitGlyph,
		suitPowerName
	} from '#lib/game/format.js';
	import type { PlayPreview, SuitEffect } from '#lib/game/rules.js';
	import { cn } from '#lib/utils.js';
	import type { Enemy } from '#lib/types.js';

	let { preview, enemy }: { preview: PlayPreview; enemy: Enemy } = $props();

	const check = $derived(preview.check);
	const rankName = $derived(cardName(enemy.card).split(' of ')[0]);

	function chip(e: SuitEffect): string {
		switch (e.suit) {
			case 'hearts':
				return `Heal ${e.amount}`;
			case 'diamonds':
				return `Draw ${e.amount}`;
			case 'clubs':
				return `×2 → ${preview.damage} dmg`;
			case 'spades':
				return `Shield ${e.amount}`;
		}
	}

	const outcome = $derived.by(() => {
		if (!check.ok) return playProblemText(check.problem);
		if (check.kind === 'jester') {
			return preview.cancelsImmunity
				? 'Cancels immunity, then you choose who goes next.'
				: 'No effect on the enemy; you choose who goes next.';
		}
		if (preview.defeats) {
			return preview.exact
				? `Defeats the ${rankName}: exact, it goes to the Tavern.`
				: `Defeats the ${rankName}.`;
		}
		const left = enemy.health - preview.damageAfter;
		return preview.attackAfter > 0
			? `Deals ${preview.damage} → HP ${left} left, then take ${preview.attackAfter} damage.`
			: `Deals ${preview.damage} → HP ${left} left; the shield holds, no damage to take.`;
	});
</script>

<div class="flex min-h-[4.25rem] flex-col gap-1 text-sm" aria-live="polite">
	{#if check.ok}
		<p class="font-semibold">
			{playKindLabel(check.kind)}{check.kind === 'jester' ? '' : ` · ${check.value}`}
		</p>
		{#if preview.effects.length > 0}
			<ul class="flex flex-wrap gap-1">
				{#each preview.effects as e (e.suit)}
					<li
						class={cn(
							'rounded-full border px-2 py-0.5 text-xs',
							e.blocked ? 'text-muted-foreground line-through opacity-70' : 'border-primary/50'
						)}
						title={e.blocked ? 'Immune' : undefined}
					>
						<span style="font-family: system-ui, sans-serif">{suitGlyph(e.suit)}&#xFE0E;</span>
						{e.blocked ? suitPowerName(e.suit) : chip(e)}
					</li>
				{/each}
			</ul>
		{/if}
		<p class="text-muted-foreground">{outcome}</p>
	{:else}
		<p class="text-muted-foreground">{outcome}</p>
	{/if}
</div>
