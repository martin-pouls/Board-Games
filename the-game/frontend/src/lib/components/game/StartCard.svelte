<!--
	A printed start card: 1 (the pile runs up to 99, arrow up) or 100 (down to 2, arrow down), a
	navy face like GameCard's, sized by the inherited `--w`. The game's mark on its home page (and
	on the start page's tile); the table shows empty piles as GameCard's dashed `start` outline.
	Decorative: the element around it carries the label.
-->
<script lang="ts">
	let { dir, class: className = '' }: { dir: 'up' | 'down'; class?: string } = $props();
</script>

<span class={['start-card font-display', dir, className]} aria-hidden="true">
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

<style>
	.start-card {
		position: relative;
		display: block;
		flex: none;
		width: var(--w);
		height: calc(var(--w) * 1.4);
		border-radius: calc(var(--w) * 0.1);
		background: var(--card-face);
		color: var(--card-ink);
		box-shadow:
			inset 0 0 0 1px var(--card-edge),
			0 2px 6px var(--card-shade);
		line-height: 1;
		text-align: center;
		font-variant-numeric: tabular-nums;
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
		color: var(--start-up);
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
		color: var(--start-down);
	}
	.down .range {
		bottom: calc(var(--w) * 0.07);
	}
</style>
