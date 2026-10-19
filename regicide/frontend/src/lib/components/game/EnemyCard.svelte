<script lang="ts">
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import Sword from '@lucide/svelte/icons/sword';
	import { getCardMotion } from '@board-games/shared/cardMotion.js';
	import { cardName } from '#lib/game/format.js';
	import { handKeys, type PlayPreview } from '#lib/game/rules.js';
	import type { Enemy, PlayerSummary } from '#lib/types.js';
	import CastleTrack from './CastleTrack.svelte';
	import JesterHat from './JesterHat.svelte';
	import PlayedCards from './PlayedCards.svelte';
	import PlayingCard from './PlayingCard.svelte';

	let {
		enemy,
		players,
		castleCount,
		preview
	}: {
		enemy: Enemy;
		players: PlayerSummary[];
		castleCount: number;
		preview?: PlayPreview | null;
	} = $props();

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
	class="flex flex-col gap-2 rounded-[14px] border border-health/35 bg-card p-2"
	aria-label="Current enemy"
>
	<CastleTrack {castleCount} />
	<div class="flex gap-2.5">
		<!-- One grid cell, so the leaving and the arriving enemy overlap instead of pushing. -->
		<div class="relative grid shrink-0">
			<!-- A keyed each of one, so the leaving enemy keeps its own key (a $derived would move on). -->
			{#each [{ card: enemy.card, key }] as e (e.key)}
				<div
					class="col-start-1 row-start-1"
					in:receive={{ key: e.key, offstage: 'top' }}
					out:send={{ key: e.key, offstage: 'none' }}
				>
					<PlayingCard card={e.card} size="lg" />
				</div>
			{/each}
			{#if enemy.immunityCancelled}
				<span
					class="absolute -top-1.5 -right-1.5 z-10 grid size-6.5 place-items-center rounded-full bg-jester text-jester-foreground ring-2 ring-card"
					title="A Jester cancelled the immunity"
				>
					<JesterHat class="size-4" label="Immunity cancelled" />
				</span>
			{/if}
		</div>
		<div class="flex min-w-0 flex-1 flex-col gap-[7px]">
			<h2 class="truncate text-lg leading-[22px] font-bold">{cardName(enemy.card)}</h2>
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
						class="hatch absolute inset-y-0"
						style="left: {pct(afterHp)}; width: {pct(remaining - afterHp)}"
					></div>
				{/if}
				<span class="absolute inset-0 grid place-items-center">
					<span class="rounded bg-card/80 px-1.5 text-sm leading-5 font-bold tabular-nums">
						{#if hits}{remaining} → {afterHp}{:else}{remaining}{/if} / {enemy.health}
					</span>
				</span>
			</div>

			<p
				class="flex h-[22px] items-center gap-1 text-base font-bold whitespace-nowrap tabular-nums"
				aria-label={`Attack ${enemy.attack}, shield ${enemy.shield}: ${enemy.effectiveAttack} damage`}
			>
				<Sword class="size-4.5 shrink-0" aria-hidden="true" />
				{enemy.attack}
				{#if enemy.shield > 0}
					<span class="flex items-center gap-0.5 text-shield" title="Shield">
						−<ShieldIcon class="size-4" aria-hidden="true" />{enemy.shield}
					</span>
					= {enemy.effectiveAttack}
				{/if}
				damage
				{#if showAttackAfter && preview}
					<span class="font-normal text-muted-foreground">→ {preview.attackAfter}</span>
				{/if}
			</p>

			<PlayedCards plays={enemy.plays} {players} />
		</div>
	</div>
</section>

<style>
	.hatch {
		background: repeating-linear-gradient(
			45deg,
			color-mix(in oklab, var(--foreground) 45%, transparent) 0 4px,
			transparent 4px 8px
		);
	}
</style>
