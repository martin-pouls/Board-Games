<script lang="ts">
	import Drama from '@lucide/svelte/icons/drama';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import Sword from '@lucide/svelte/icons/sword';
	import { getCardMotion } from '@board-games/shared/cardMotion.js';
	import { handKeys, type PlayPreview } from '#lib/game/rules.js';
	import type { Enemy, PlayerSummary } from '#lib/types.js';
	import PlayedCards from './PlayedCards.svelte';
	import PlayingCard from './PlayingCard.svelte';

	let {
		enemy,
		players,
		preview
	}: { enemy: Enemy; players: PlayerSummary[]; preview?: PlayPreview | null } = $props();

	// A defeated enemy flies to the last turn's Played pile (same key there), the next one comes
	// in from the Castle deck above.
	const { send, receive } = getCardMotion();
	const key = $derived(handKeys([enemy.card])[0]);

	const pct = (n: number) => `${Math.max(0, Math.min(100, (n / enemy.health) * 100))}%`;
	const remaining = $derived(Math.max(0, enemy.health - enemy.damage));
	const hits = $derived(!!preview?.check.ok && preview.damage > 0);
	const afterHp = $derived(
		hits && preview ? Math.max(0, enemy.health - preview.damageAfter) : remaining
	);
	const showAttackAfter = $derived(
		!!preview &&
			preview.check.ok &&
			!preview.defeats &&
			preview.attackAfter !== enemy.effectiveAttack
	);
</script>

<section
	class="flex items-start gap-2.5 rounded-xl border border-health/50 bg-card/70 p-2 shadow-[0_0_24px_-8px] shadow-health/40"
	aria-label="Current enemy"
>
	<!-- One grid cell, so the leaving and the arriving enemy overlap instead of pushing. -->
	<div class="grid shrink-0">
		<!-- A keyed each of one, so the leaving enemy keeps its own key (a $derived would move on). -->
		{#each [{ card: enemy.card, key }] as e (e.key)}
			<div
				class="col-start-1 row-start-1 rounded-lg ring-2 ring-health/70"
				in:receive={{ key: e.key, offstage: 'top' }}
				out:send={{ key: e.key, offstage: 'none' }}
			>
				<PlayingCard card={e.card} size="lg" />
			</div>
		{/each}
	</div>
	<div class="flex min-w-0 flex-1 flex-col gap-1.5">
		<div
			class="relative h-6 overflow-hidden rounded-md bg-muted"
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
			{#if hits}
				<div
					class="absolute inset-y-0 bg-foreground/50"
					style="left: {pct(afterHp)}; width: {pct(
						remaining - afterHp
					)}; background-image: repeating-linear-gradient(45deg, transparent 0 4px, oklch(0 0 0 / 35%) 4px 8px)"
				></div>
			{/if}
			<span class="absolute inset-0 flex items-center justify-center">
				<span class="rounded bg-background/75 px-1.5 text-sm leading-5 font-semibold tabular-nums">
					{#if hits}{remaining} → {afterHp}{:else}{remaining}{/if} / {enemy.health}
				</span>
			</span>
		</div>

		<p
			class="flex h-6 items-center gap-1 text-base font-semibold tabular-nums"
			aria-label={`Attack ${enemy.attack}, shield ${enemy.shield}: ${enemy.effectiveAttack} damage`}
		>
			<Sword class="size-4.5" aria-hidden="true" />
			{enemy.attack}
			{#if enemy.shield > 0}
				<span class="flex items-center gap-0.5 text-shield" title="Shield">
					−<ShieldIcon class="size-4" aria-hidden="true" />{enemy.shield}
				</span>
				<span>= {enemy.effectiveAttack}</span>
			{/if}
			{#if showAttackAfter && preview}
				<span class="font-normal text-muted-foreground">→ {preview.attackAfter}</span>
			{/if}
			{#if enemy.immunityCancelled}
				<span class="ml-auto" title="A Jester cancelled the immunity">
					<Drama class="size-4.5 text-jester" aria-label="Immunity cancelled" />
				</span>
			{/if}
		</p>

		<PlayedCards plays={enemy.plays} {players} />
	</div>
</section>
