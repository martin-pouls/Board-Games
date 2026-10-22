<!--
	The home card's picture: a strip of the start-page tile's scene along its top edge, so the tile
	you tapped carries on into the page it opens. Planet Nine: the rocket flies up a dotted trail
	towards the lit rim of the planet. Deep Sea: the bathyscaphe dives down a trail of bubbles
	through the lamp beam towards the sunken temple. The scene is 340 × 76 and cut to the card's
	width (slice), so phones show its middle. Colours: the --band-* tokens of each edition.
-->
<script lang="ts">
	import { edition } from '#lib/edition.js';

	/** Stars and motes: x, y, radius, opacity. */
	const STARS = [
		[14, 12, 1.2, 0.8],
		[44, 30, 0.9, 0.6],
		[30, 58, 1, 0.6],
		[80, 10, 1.1, 0.8],
		[112, 50, 0.8, 0.5],
		[128, 20, 1, 0.7],
		[210, 14, 1.2, 0.8],
		[236, 44, 0.9, 0.6],
		[262, 10, 1, 0.7],
		[196, 64, 0.8, 0.5],
		[326, 20, 1.1, 0.7],
		[96, 70, 0.8, 0.4],
		[300, 6, 0.8, 0.6]
	];
	const MOTES = [
		[18, 40, 0.9, 0.5],
		[60, 62, 0.7, 0.4],
		[110, 8, 0.8, 0.45],
		[214, 58, 0.9, 0.4],
		[250, 16, 0.7, 0.5],
		[330, 40, 0.8, 0.35],
		[92, 30, 0.6, 0.4]
	];
	const BUBBLES = [
		[120, 18, 1.6, 0.55],
		[106, 26, 1.1, 0.45],
		[132, 9, 1.2, 0.4]
	];
	const TEMPLE_COLUMNS = [-28, -17, -6, 5, 16, 27];
	/** The beam's gradient needs an id of its own on the page. */
	const uid = $props.id();
</script>

<svg
	viewBox="0 0 340 76"
	preserveAspectRatio="xMidYMid slice"
	class="band block h-19 w-full text-(--band-accent)"
	aria-hidden="true"
>
	{#if edition === 'deepSea'}
		<defs>
			<linearGradient id="{uid}-beam" x1="0" y1="0" x2="0" y2="1">
				<stop offset="0" stop-color="var(--band-beam)" stop-opacity="0.34" />
				<stop offset="1" stop-color="var(--band-beam)" stop-opacity="0" />
			</linearGradient>
		</defs>
		{#each MOTES as [x, y, r, o] (`${x},${y}`)}
			<circle cx={x} cy={y} {r} fill="white" fill-opacity={o} />
		{/each}
		<path d="M246 -4 L306 -4 L330 80 L222 80z" fill="url(#{uid}-beam)" />
		<g transform="translate(290 66) scale(0.82)" fill="var(--band-shape)">
			<path d="M-27 -2a27 27 0 0 1 54 0z" />
			<rect x="-2" y="-34" width="4" height="7" rx="1" />
			<rect x="-34" y="-2" width="68" height="5" rx="1" />
			{#each TEMPLE_COLUMNS as c (c)}
				<rect x={c - 1.5} y="3" width="4.5" height="20" />
			{/each}
			<path
				d="M-27 -2a27 27 0 0 1 54 0"
				fill="none"
				stroke="var(--band-beam)"
				stroke-width="1.6"
				stroke-opacity="0.75"
			/>
		</g>
		<path
			d="M-4 8C40 12 92 20 142 32"
			fill="none"
			stroke="white"
			stroke-opacity="0.55"
			stroke-width="1.8"
			stroke-dasharray="0.1 5.5"
			stroke-linecap="round"
		/>
		{#each BUBBLES as [x, y, r, o] (`${x},${y}`)}
			<circle cx={x} cy={y} {r} fill="none" stroke="white" stroke-opacity={o} stroke-width="0.9" />
		{/each}
		<!-- The bathyscaphe: float tank, hull on skids, the glass dome in front, its lamp. -->
		<g transform="translate(150 14) rotate(18) scale(1.9)">
			<rect x="4.5" y="4.6" width="9" height="3.6" rx="1.8" fill="currentColor" />
			<path d="M9 8v2" stroke="currentColor" stroke-width="1.8" />
			<rect x="1.5" y="9.5" width="13" height="8" rx="4" fill="currentColor" />
			<path d="M4.5 19.6h9" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" />
			<path d="M6 17v2.6M12 17v2.6" stroke="currentColor" stroke-width="1.2" />
			<circle cx="15" cy="13.5" r="5.2" fill="currentColor" />
			<circle cx="15.6" cy="13.3" r="3.6" fill="var(--band-window)" />
			<circle cx="16.7" cy="12.2" r="0.95" fill="currentColor" fill-opacity="0.6" />
			<path d="M20 11.7 23.6 9.6v8l-3.6-2.1z" fill="currentColor" fill-opacity="0.42" />
		</g>
	{:else}
		{#each STARS as [x, y, r, o] (`${x},${y}`)}
			<circle cx={x} cy={y} {r} fill="white" fill-opacity={o} />
		{/each}
		<circle cx="300" cy="128" r="78" fill="var(--band-shape)" />
		<circle
			cx="300"
			cy="128"
			r="78"
			fill="none"
			stroke="currentColor"
			stroke-width="8"
			stroke-opacity="0.2"
		/>
		<circle cx="300" cy="128" r="78" fill="none" stroke="currentColor" stroke-width="2" />
		<path
			d="M-4 74C44 70 100 60 150 42"
			fill="none"
			stroke="white"
			stroke-opacity="0.55"
			stroke-width="1.8"
			stroke-dasharray="0.1 5"
			stroke-linecap="round"
		/>
		<!-- The rocket, tilted up along its trail. -->
		<g transform="translate(140 4) scale(1.75) rotate(40 14 17)">
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
	{/if}
</svg>

<style>
	.band {
		background: var(--band-bg);
	}
</style>
