<!--
	A game tile's art band: the 88 px square on the phone's shelf, a wide band on desktop. It is a
	size container, so everything inside scales with its height (cqh). The Game shows its two
	printed start cards, Regicide a king on its misty ridges with a card index, Fox Duet a slice of
	its board with the fox on the path, The Crew its rocket flying up to Planet Nine; Mission Deep
	Sea keeps its icon until its own theme is done.
-->
<script lang="ts">
	import type { Game } from '#lib/games.js';

	let { game, class: className = '' }: { game: Game['id']; class?: string } = $props();

	const band: Record<Game['id'], string> = {
		'the-game': 'bg-the-game-band',
		regicide: 'bg-regicide-band',
		'fox-duet': 'bg-fox-band',
		'the-crew': 'bg-linear-to-br from-crew-ink via-crew-ink to-crew-tile',
		'deep-sea': 'bg-linear-to-b from-deep-ink via-deep-ink to-deep-tile'
	};

	/* Fox Duet's forest path: a gentle meander with stepping stones, a gem tile above and below. */
	const pathY = (x: number) => +(56 + 8 * Math.sin((x - 120) / 24)).toFixed(1);
	const meander =
		`M-10 ${pathY(-10)}` +
		Array.from({ length: 65 }, (_, i) => -6 + i * 4)
			.map((x) => ` L${x} ${pathY(x)}`)
			.join('');
	const FOX_STONES = [12, 42, 72, 96, 146, 172, 202, 230];
	/** Each tile hangs off a stone by a dotted line, from the stone's edge to the tile's. */
	/** The Crew's starfield: x, y, radius, opacity. */
	const CREW_STARS = [
		[18, 14, 1.3, 0.8],
		[52, 30, 1, 0.6],
		[38, 70, 1.1, 0.7],
		[84, 12, 1.2, 0.8],
		[150, 14, 1, 0.7],
		[178, 40, 1.3, 0.8],
		[214, 18, 1, 0.6],
		[228, 54, 1.1, 0.5],
		[96, 86, 0.9, 0.5],
		[140, 34, 0.9, 0.6],
		[8, 50, 0.9, 0.5]
	];
	const FOX_TILES = [
		{ x: 96, y: 24, tilt: -6, gems: [0], link: `M96 ${pathY(96) - 6}V36` },
		{ x: 146, y: 82, tilt: 5, gems: [-5, 5], link: `M146 ${pathY(146) + 6}V70` }
	];
</script>

{#snippet startCard(dir: 'up' | 'down')}
	<!-- The printed start cards: 1 (the pile runs up to 99) and 100 (down to 2). -->
	<span class={['start-card font-the-game', dir]}>
		<span class="big">{dir === 'up' ? 1 : 100}</span>
		<svg class="arrow" viewBox="0 0 20 26" fill="none" stroke="currentColor" stroke-width="2.2">
			{#if dir === 'up'}
				<path d="M10 2 2.5 12.5h4.5V24h6V12.5h4.5z" stroke-linejoin="round" />
				<path d="M7 16.5h6M7 20h6" stroke-linecap="round" />
			{:else}
				<path d="M10 24 2.5 13.5h4.5V2h6v11.5h4.5z" stroke-linejoin="round" />
				<path d="M7 9.5h6M7 6h6" stroke-linecap="round" />
			{/if}
		</svg>
		<span class="range">{dir === 'up' ? 99 : 2}</span>
	</span>
{/snippet}

<div class={['art', band[game], className]} aria-hidden="true">
	{#if game === 'the-game'}
		<div class="start-cards">
			{@render startCard('up')}
			{@render startCard('down')}
		</div>
	{:else if game === 'regicide'}
		<svg class="ridges" viewBox="0 0 400 100" preserveAspectRatio="none">
			<path
				d="M0 30 L56 12 L104 22 L168 0 L224 18 L280 6 L336 20 L400 10 V100 H0Z"
				fill="var(--regicide-ridge-1)"
			/>
			<path
				d="M0 58 L72 38 L128 50 L192 30 L248 46 L304 34 L360 46 L400 40 V100 H0Z"
				fill="var(--regicide-ridge-2)"
			/>
			<path
				d="M0 80 L48 66 L112 76 L176 58 L232 72 L304 62 L400 74 V100 H0Z"
				fill="var(--regicide-ridge-3)"
			/>
		</svg>
		<span class="index"><b class="font-alegreya">K</b><i class="suit-glyph">♠︎</i></span>
		<span class="king font-alegreya">K</span>
	{:else if game === 'fox-duet'}
		<!-- A slice of the board, 240 × 100, cut to the band: the phone's square shows the middle. -->
		<svg class="scene" viewBox="0 0 240 100" preserveAspectRatio="xMidYMid slice">
			<g fill="var(--fox-forest)">
				<circle cx="-2" cy="4" r="22" />
				<circle cx="246" cy="102" r="26" />
				<circle cx="236" cy="-6" r="14" />
			</g>
			<path
				d={meander}
				fill="none"
				stroke="var(--fox-path)"
				stroke-width="3"
				stroke-linecap="round"
			/>
			{#each FOX_TILES as t (t.x)}
				<path d={t.link} stroke="var(--fox-path)" stroke-width="2" stroke-dasharray="2 3" />
				<g transform="translate({t.x} {t.y}) rotate({t.tilt})">
					<rect x="-12" y="-12" width="24" height="24" rx="4" class="plate" />
					<rect x="-9" y="-9" width="18" height="18" rx="2" class="plate-inner" />
					{#each t.gems as gx (gx)}
						<path d="M{gx} -6.5l4.5 4.3-4.5 8.7-4.5-8.7z" class="gem" />
					{/each}
				</g>
			{/each}
			{#each FOX_STONES as x (x)}
				<circle cx={x} cy={pathY(x)} r="5.5" class="stone" />
			{/each}
			<!-- The traveller: the fox on an orange disc, as on the board. -->
			<g transform="translate(120 {pathY(120)})">
				<circle r="15" fill="var(--fox-disc)" stroke="var(--fox-disc-rim)" stroke-width="2.5" />
				<g transform="translate(-9.6 -9.6) scale(0.8)">
					<path d="M2 3 8.6 8.6h6.8L22 3l-1 9.2L12 21.5 3 12.2Z" fill="var(--fox-disc-ink)" />
					<circle cx="8.8" cy="12.6" r="1.2" fill="var(--fox-disc)" />
					<circle cx="15.2" cy="12.6" r="1.2" fill="var(--fox-disc)" />
				</g>
			</g>
		</svg>
	{:else if game === 'the-crew'}
		<!-- A slice of space, 240 × 100: the rocket flies up a dotted trail towards the lit rim of
		     Planet Nine rising from the corner. -->
		<svg class="scene text-crew-accent" viewBox="0 0 240 100" preserveAspectRatio="xMidYMid slice">
			{#each CREW_STARS as [x, y, r, o] (`${x},${y}`)}
				<circle cx={x} cy={y} {r} fill="white" fill-opacity={o} />
			{/each}
			<circle cx="172" cy="128" r="62" fill="var(--crew-planet)" />
			<circle
				cx="172"
				cy="128"
				r="62"
				fill="none"
				stroke="currentColor"
				stroke-width="7"
				stroke-opacity="0.22"
			/>
			<circle cx="172" cy="128" r="62" fill="none" stroke="currentColor" stroke-width="2" />
			<path
				d="M14 98C46 94 80 84 106 68"
				fill="none"
				stroke="white"
				stroke-opacity="0.55"
				stroke-width="1.8"
				stroke-dasharray="0.1 5"
				stroke-linecap="round"
			/>
			<g transform="translate(96 18) scale(1.9) rotate(40 14 17)">
				<path
					d="M14 4c3 2.6 4.4 6 4.4 10v7h-8.8v-7c0-4 1.4-7.4 4.4-10z"
					fill="currentColor"
					stroke="currentColor"
					stroke-width="1"
					stroke-linejoin="round"
				/>
				<circle cx="14" cy="12.5" r="2" fill="white" fill-opacity="0.85" />
				<path
					d="M9.6 17 6 22.5l3.6-1zm8.8 0 3.6 5.5-3.6-1z"
					fill="currentColor"
					stroke="currentColor"
					stroke-width="1"
					stroke-linejoin="round"
				/>
				<path d="M11.8 22.5h4.4L14 27.5z" fill="white" fill-opacity="0.85" />
			</g>
		</svg>
	{:else}
		<svg viewBox="0 0 32 32" class="icon text-deep-accent">
			<circle cx="5" cy="8" r="1.1" fill="white" fill-opacity="0.55" />
			<circle cx="8" cy="4.5" r="0.7" fill="white" fill-opacity="0.45" />
			<circle cx="27" cy="9" r="0.9" fill="white" fill-opacity="0.5" />
			<circle cx="29" cy="24" r="0.7" fill="white" fill-opacity="0.4" />
			<path
				d="M19 4h5v2.5h-3v3h-2z"
				fill="currentColor"
				stroke="currentColor"
				stroke-width="1"
				stroke-linejoin="round"
			/>
			<path
				d="M11 10.5h7v3h-7z"
				fill="currentColor"
				stroke="currentColor"
				stroke-width="1"
				stroke-linejoin="round"
			/>
			<ellipse cx="15" cy="18.5" rx="11" ry="5.5" fill="currentColor" />
			<path
				d="M25 16.5 29 13.5v10l-4-3z"
				fill="currentColor"
				stroke="currentColor"
				stroke-width="1"
				stroke-linejoin="round"
			/>
			<circle cx="9" cy="18.5" r="1.5" fill="white" fill-opacity="0.85" />
			<circle cx="14" cy="18.5" r="1.5" fill="white" fill-opacity="0.85" />
			<circle cx="19" cy="18.5" r="1.5" fill="white" fill-opacity="0.85" />
			<path
				d="M6 28.5c1.2-1.6 2.8-1.6 4 0s2.8 1.6 4 0 2.8-1.6 4 0 2.8 1.6 4 0"
				fill="none"
				stroke="currentColor"
				stroke-width="1"
				stroke-linecap="round"
				stroke-opacity="0.5"
			/>
		</svg>
	{/if}
</div>

<style>
	.art {
		position: relative;
		container-type: size;
		overflow: hidden;
		flex: none;
		display: grid;
		place-items: center;
	}
	.icon {
		width: 66cqh;
		height: 66cqh;
	}

	/* The Game: navy cards with cream numerals, fanned 5° outwards (as on its home page). */
	.start-cards {
		display: flex;
		align-items: center;
		gap: 6cqh;
	}
	.start-card {
		--w: 40cqh;
		position: relative;
		display: block;
		width: var(--w);
		height: calc(var(--w) * 1.4);
		border-radius: calc(var(--w) * 0.1);
		background: var(--the-game-face);
		color: var(--the-game-ink);
		box-shadow:
			inset 0 0 0 1px var(--the-game-edge),
			0 2px 6px var(--the-game-shade);
		line-height: 1;
		text-align: center;
		font-variant-numeric: tabular-nums;
	}
	.start-card.up {
		rotate: -5deg;
	}
	.start-card.down {
		rotate: 5deg;
	}
	.start-card > * {
		position: absolute;
		left: 0;
		right: 0;
	}
	.big {
		font-size: calc(var(--w) * 0.56);
	}
	.range {
		font-size: calc(var(--w) * 0.17);
	}
	.arrow {
		left: 50%;
		width: calc(var(--w) * 0.42);
		height: calc(var(--w) * 0.52);
		translate: -50% 0;
	}
	.up .range {
		top: calc(var(--w) * 0.07);
	}
	.up .arrow {
		top: calc(var(--w) * 0.25);
		color: var(--the-game-up);
	}
	.up .big {
		bottom: calc(var(--w) * 0.08);
	}
	.down .big {
		top: calc(var(--w) * 0.08);
		font-size: calc(var(--w) * 0.5);
	}
	.down .arrow {
		bottom: calc(var(--w) * 0.25);
		color: var(--the-game-down);
	}
	.down .range {
		bottom: calc(var(--w) * 0.07);
	}

	/* Fox Duet: the board's pieces in its tokens. */
	.scene {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
	}
	.plate {
		fill: var(--fox-plate);
		stroke: color-mix(in oklch, var(--fox-ink) 25%, transparent);
		stroke-width: 1.2;
	}
	.plate-inner {
		fill: none;
		stroke: color-mix(in oklch, var(--fox-ink) 12%, transparent);
		stroke-dasharray: 2 2;
	}
	.gem {
		fill: var(--fox-gem);
		stroke: color-mix(in oklch, var(--fox-ink) 35%, transparent);
		stroke-width: 0.6;
	}
	.stone {
		fill: var(--fox-path);
		stroke: color-mix(in oklch, var(--fox-ink) 20%, transparent);
		stroke-width: 1.2;
	}

	/* Regicide: the king standing on three ridges, a card index in the corner. */
	.ridges {
		position: absolute;
		inset: auto 0 0;
		width: 100%;
		height: 46%;
	}
	.king {
		position: absolute;
		left: 50%;
		bottom: 16cqh;
		translate: -50% 0;
		font-size: 82cqh;
		font-weight: 700;
		line-height: 0.8;
		color: var(--regicide-letter);
	}
	.index {
		position: absolute;
		top: 7cqh;
		left: 8cqh;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2cqh;
		color: var(--regicide-letter);
	}
	.index b {
		font-size: 19cqh;
		font-weight: 700;
		line-height: 0.82;
	}
	.index i {
		font-size: 14cqh;
		line-height: 1;
	}
</style>
