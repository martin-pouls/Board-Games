<!--
	The home card's picture: a strip of the forest board across its top edge, the path with stepping
	stones, a gem tile above it and one below, and the traveller (the fox on its orange disc, as on
	ForestBoard). The scene is 340 × 72 and cut to the card's width (slice), so phones show its middle.
-->
<script lang="ts">
	const W = 340;
	const BASE = 40;
	const pathY = (x: number) => +(BASE + 9 * Math.sin((x - 60) / 22)).toFixed(1);
	const path =
		`M-10 ${pathY(-10)}` +
		Array.from({ length: Math.ceil((W + 20) / 4) }, (_, i) => -6 + i * 4)
			.map((x) => ` L${x} ${pathY(x)}`)
			.join('');
	const STONES = [14, 48, 82, 116, 196, 230, 264, 298, 330];
	/** Each tile hangs off a stone by a dotted line, from the stone's edge to the tile's. */
	const TILES = [
		{ x: 116, y: 13, tilt: -6, gems: [0], link: `M116 ${pathY(116) - 6}V24` },
		{ x: 230, y: 64, tilt: 5, gems: [-5, 5], link: `M230 ${pathY(230) + 6}V56` }
	];
	const FOX = 160;
</script>

<svg
	viewBox="0 0 {W} 72"
	preserveAspectRatio="xMidYMid slice"
	class="block h-18 w-full bg-meadow"
	aria-hidden="true"
>
	<g class="fill-forest/25">
		<circle cx="-4" cy="2" r="26" />
		<circle cx="344" cy="74" r="28" />
		<circle cx="300" cy="-8" r="16" />
	</g>
	<path d={path} fill="none" class="stroke-path" stroke-width="3" stroke-linecap="round" />
	{#each STONES as x (x)}
		<circle cx={x} cy={pathY(x)} r="5" class="fill-path stroke-ink/15" />
	{/each}
	{#each TILES as t (t.x)}
		<path d={t.link} class="stroke-path" stroke-width="2" stroke-dasharray="2 3" />
		<g transform="translate({t.x} {t.y}) rotate({t.tilt})">
			<rect x="-11" y="-11" width="22" height="22" rx="4" class="fill-tile stroke-ink/20" />
			{#each t.gems as dx (dx)}
				<path d="M{dx} -5 {dx + 4} 0 {dx} 5 {dx - 4} 0Z" class="fill-gem" />
			{/each}
		</g>
	{/each}
	<g transform="translate({FOX} {pathY(FOX)}) scale(1.18)">
		<circle r="11" class="fill-fox stroke-parchment" stroke-width="2" />
		<g transform="translate(-7 -7) scale(0.5833)">
			<path d="M2 3 8.6 8.6h6.8L22 3l-1 9.2L12 21.5 3 12.2Z" class="fill-primary-foreground" />
			<circle cx="8.8" cy="12.6" r="1.2" class="fill-fox" />
			<circle cx="15.2" cy="12.6" r="1.2" class="fill-fox" />
		</g>
	</g>
</svg>
