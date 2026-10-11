<script lang="ts">
	import PawPrint from '@lucide/svelte/icons/paw-print';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import { ABILITY, MOVEMENT } from '#lib/game/format.js';

	let { open = $bindable(false) }: { open?: boolean } = $props();

	const ranks = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
	const ABILITY_RANK = {
		1: 'musician',
		3: 'foxes',
		5: 'gazelle',
		7: 'gift',
		9: 'royalHeir'
	} as const;
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="max-h-[85dvh] overflow-y-auto">
		<Dialog.Header>
			<Dialog.Title>Rules at a glance</Dialog.Title>
			<Dialog.Description>
				Collect every gem within 3 rounds without getting lost. Don't talk about your cards.
			</Dialog.Description>
		</Dialog.Header>
		<div class="flex flex-col gap-3 text-sm leading-snug">
			<section>
				<h3 class="mb-1 font-semibold">Tricks</h3>
				<ul class="list-disc pl-5">
					<li>The leader plays any card. The follower must follow the lead suit if possible.</li>
					<li>
						The highest trump (the decree's suit) wins; without trump, the highest card of the lead
						suit.
					</li>
					<li>
						The traveller moves the sum of both cards' paws toward the winner's end. Ending on a
						location with gems collects one, even without moving.
					</li>
					<li>
						Past the end of the path: back to the start, and a forest token covers that end (its
						gems move one location inward). No token left: lost in the forest.
					</li>
				</ul>
			</section>
			<section>
				<h3 class="mb-1 font-semibold">Abilities</h3>
				<ul class="flex flex-col gap-1">
					{#each Object.entries(ABILITY_RANK) as [rank, ability] (rank)}
						<li>
							<b>{rank} {ABILITY[ability].name}:</b>
							{ABILITY[ability].text}
						</li>
					{/each}
				</ul>
			</section>
			<section>
				<h3 class="mb-1 font-semibold">Movement</h3>
				<div class="grid grid-cols-10 gap-0.5 text-center">
					{#each ranks as r (r)}
						<div class="flex flex-col items-center rounded bg-muted/60 py-1">
							<b>{r}</b>
							<span class="flex items-center gap-0.5 text-muted-foreground">
								{MOVEMENT[r]}<PawPrint class="size-3" />
							</span>
						</div>
					{/each}
				</div>
			</section>
			<section>
				<h3 class="mb-1 font-semibold">Round end (rounds 1 and 2)</h3>
				<p>
					Each location with a plus gets a gem. Then a forest token covers one end (the host
					chooses; it must be the other end if the traveller stands at one). The next player deals.
				</p>
			</section>
			<section>
				<h3 class="mb-1 font-semibold">Score on a win</h3>
				<p>
					10/20/30 for the level, +1 per card left in hand, +10 for a win in round 2, +3 per forest
					token left.
				</p>
			</section>
		</div>
	</Dialog.Content>
</Dialog.Root>
