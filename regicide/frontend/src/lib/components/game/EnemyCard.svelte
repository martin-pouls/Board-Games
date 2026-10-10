<script lang="ts">
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import Sword from '@lucide/svelte/icons/sword';
	import { suitGlyph } from '#lib/game/format.js';
	import type { PlayPreview } from '#lib/game/rules.js';
	import type { Enemy, PlayerSummary } from '#lib/types.js';
	import PlayedCards from './PlayedCards.svelte';
	import PlayingCard from './PlayingCard.svelte';

	let {
		enemy,
		players,
		preview
	}: { enemy: Enemy; players: PlayerSummary[]; preview?: PlayPreview | null } = $props();

	const pct = (n: number) => `${Math.max(0, Math.min(100, (n / enemy.health) * 100))}%`;
	const remaining = $derived(Math.max(0, enemy.health - enemy.damage));
	const afterHp = $derived(preview ? Math.max(0, enemy.health - preview.damageAfter) : remaining);
	const showAttackAfter = $derived(
		!!preview &&
			preview.check.ok &&
			!preview.defeats &&
			preview.attackAfter !== enemy.effectiveAttack
	);
	const glyph = $derived(suitGlyph(enemy.immunity) + '︎');
</script>

<section
	class="flex gap-3 rounded-xl border border-health/50 bg-card/70 p-3 shadow-[0_0_24px_-8px] shadow-health/40"
	aria-label="Current enemy"
>
	<div class="shrink-0 rounded-xl ring-2 ring-health/70">
		<PlayingCard card={enemy.card} size="lg" />
	</div>
	<div class="flex min-w-0 flex-1 flex-col gap-2">
		<div>
			<div class="mb-1 flex items-baseline justify-between text-sm">
				<span class="font-display font-semibold">HP {remaining}/{enemy.health}</span>
				{#if preview?.check.ok && preview.damage > 0}
					<span class="text-xs text-muted-foreground tabular-nums">
						{preview.defeats ? 'defeated' : `→ ${afterHp}`}
					</span>
				{/if}
			</div>
			<div
				class="relative h-3 overflow-hidden rounded-full bg-muted"
				role="progressbar"
				aria-label="Enemy health"
				aria-valuemin={0}
				aria-valuemax={enemy.health}
				aria-valuenow={remaining}
			>
				<div
					class="absolute inset-y-0 left-0 bg-health transition-[width] duration-300"
					style="width: {pct(remaining)}"
				></div>
				{#if preview?.check.ok && preview.damage > 0}
					<div
						class="absolute inset-y-0 bg-foreground/60"
						style="left: {pct(afterHp)}; width: {pct(
							remaining - afterHp
						)}; background-image: repeating-linear-gradient(45deg, transparent 0 4px, oklch(0 0 0 / 35%) 4px 8px)"
					></div>
				{/if}
			</div>
		</div>

		<p class="flex flex-wrap items-center gap-x-1.5 text-sm">
			<Sword class="size-4 shrink-0" aria-hidden="true" />
			<span>Attack {enemy.attack}</span>
			{#if enemy.shield > 0}
				<span class="flex items-center gap-1 text-shield">
					− {enemy.shield}
					<ShieldIcon class="size-3.5" aria-label="shield" />
				</span>
				<span class="font-semibold">= {enemy.effectiveAttack}</span>
			{/if}
			{#if showAttackAfter && preview}
				<span class="text-xs text-muted-foreground">(→ {preview.attackAfter})</span>
			{/if}
		</p>

		<p class="flex items-center gap-1.5 text-sm">
			<span
				class="inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs"
				title={enemy.immunityCancelled
					? 'A Jester cancelled the immunity'
					: `Immune to the ${glyph} power`}
			>
				<ShieldIcon class="size-3" aria-hidden="true" />
				{#if enemy.immunityCancelled}
					<span class="line-through opacity-70" style="font-family: system-ui, sans-serif"
						>Immune to {glyph}</span
					>
					<span class="font-semibold text-jester">Jester</span>
				{:else}
					<span style="font-family: system-ui, sans-serif">Immune to {glyph}</span>
				{/if}
			</span>
		</p>

		<PlayedCards plays={enemy.plays} {players} />
	</div>
</section>
