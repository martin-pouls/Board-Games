<script lang="ts">
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import { briefing } from '#lib/game/format.js';
	import type { MissionDef } from '#lib/types.js';

	let { open = $bindable(false), mission }: { open?: boolean; mission: MissionDef | null } =
		$props();
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="max-h-[85dvh] overflow-y-auto">
		<Dialog.Header>
			<Dialog.Title>Rules at a glance</Dialog.Title>
			<Dialog.Description>
				Fulfil every task together. Don't talk about your cards; the only way to share is to
				communicate one card.
			</Dialog.Description>
		</Dialog.Header>
		<div class="flex flex-col gap-3 text-sm leading-snug">
			{#if mission}
				<section>
					<h3 class="mb-1 font-semibold">Mission {mission.number}</h3>
					<ul class="list-disc pl-5">
						{#each briefing(mission) as line (line)}
							<li>{line}</li>
						{:else}
							<li>No special rules.</li>
						{/each}
					</ul>
				</section>
			{/if}
			<section>
				<h3 class="mb-1 font-semibold">Tricks</h3>
				<ul class="list-disc pl-5">
					<li>
						The commander (whoever holds rocket 4) leads the first trick; the winner leads the next.
					</li>
					<li>Follow the colour that was led if you can. Otherwise play any card.</li>
					<li>The highest rocket wins; without a rocket, the highest card of the led colour.</li>
				</ul>
			</section>
			<section>
				<h3 class="mb-1 font-semibold">Tasks</h3>
				<ul class="list-disc pl-5">
					<li>A task is fulfilled when its owner wins the trick with that card.</li>
					<li>If anyone else wins it, the attempt is lost.</li>
					<li>
						Tokens fix the order: 1–5 in that order, Ω last of all, and › before ›› before ›››
						before ››››.
					</li>
				</ul>
			</section>
			<section>
				<h3 class="mb-1 font-semibold">Communication</h3>
				<ul class="list-disc pl-5">
					<li>
						Once per mission, before a trick starts, you may show one card: your highest (▲), only
						(●) or lowest (▼) card of its colour. Never a rocket.
					</li>
					<li>In a dead zone (?) the position stays secret.</li>
					<li>
						Distress signal: before the first trick, the commander may have everyone pass one
						non-rocket card left or right. It counts as an extra attempt.
					</li>
				</ul>
			</section>
			<section>
				<h3 class="mb-1 font-semibold">JARVIS (2 players)</h3>
				<ul class="list-disc pl-5">
					<li>
						JARVIS plays only its face-up cards, chosen by the commander. A hidden card turns face
						up after the trick in which the card above it was played.
					</li>
				</ul>
			</section>
		</div>
	</Dialog.Content>
</Dialog.Root>
