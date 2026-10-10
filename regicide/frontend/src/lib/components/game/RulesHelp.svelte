<script lang="ts">
	import * as Dialog from '#lib/components/ui/dialog/index.js';

	let { open = $bindable(false) }: { open?: boolean } = $props();
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="max-h-[85dvh] overflow-y-auto">
		<Dialog.Header>
			<Dialog.Title>Rules at a glance</Dialog.Title>
			<Dialog.Description
				>Defeat all 12 enemies together. Do not reveal your hand.</Dialog.Description
			>
		</Dialog.Header>
		<div class="flex flex-col gap-3 text-sm leading-snug">
			<section>
				<h3 class="mb-1 font-semibold">Suit powers</h3>
				<ul class="list-disc pl-5">
					<li><b>♥ Heal:</b> shuffle the discard, move that many cards under the Tavern.</li>
					<li>
						<b>♦ Draw:</b> that many cards, one at a time clockwise from you (full hands skip).
					</li>
					<li><b>♣ Double damage:</b> the play deals twice its value.</li>
					<li><b>♠ Shield:</b> reduces the enemy attack for the rest of the fight, cumulative.</li>
				</ul>
				<p class="mt-1 text-muted-foreground">
					Hearts resolve before Diamonds. Each suit counts once per play. An enemy is immune to the
					power of its own suit (its cards still deal damage).
				</p>
			</section>
			<section>
				<h3 class="mb-1 font-semibold">Legal plays</h3>
				<ul class="list-disc pl-5">
					<li>Any single card.</li>
					<li>A combo: 2-4 number cards of the same rank, total 10 or less.</li>
					<li>An Animal Companion (Ace) with one other card; Ace counts 1.</li>
					<li>The Jester, always alone.</li>
				</ul>
			</section>
			<section>
				<h3 class="mb-1 font-semibold">Taking damage</h3>
				<p>
					If the enemy survives, discard cards worth at least its attack (minus shields). Jack 10,
					Queen 15, King 20 in hand. If you cannot, everyone loses. Defeating an enemy with exactly
					its health puts it on top of the Tavern.
				</p>
			</section>
			<section>
				<h3 class="mb-1 font-semibold">Yielding</h3>
				<p>Skip your play and take the damage. Not allowed if everyone else yielded last turn.</p>
			</section>
			<section>
				<h3 class="mb-1 font-semibold">The Jester</h3>
				<p>
					Cancels the enemy's immunity for the fight. No damage and no discard; you choose who goes
					next (yourself included). Others may signal while you decide.
				</p>
			</section>
			<section>
				<h3 class="mb-1 font-semibold">Solo</h3>
				<p>
					Two Jester powers: discard your hand and refill to 8, at the start of your turn or before
					discarding. Win with none used for Gold, one for Silver, two for Bronze.
				</p>
			</section>
		</div>
	</Dialog.Content>
</Dialog.Root>
