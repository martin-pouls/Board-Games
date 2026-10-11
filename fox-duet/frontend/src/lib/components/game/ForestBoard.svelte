<script lang="ts">
	import { cubicInOut } from 'svelte/easing';
	import { prefersReducedMotion, Tween } from 'svelte/motion';
	import { displayIndex, moveLabel } from '#lib/game/format.js';
	import type { Board, End, MoveOption, PlayerSummary } from '#lib/types.js';

	let {
		board,
		myEnd,
		players,
		options = [],
		highlight = null,
		onchoose,
		onhover
	}: {
		board: Board;
		/** The viewer's end; it is drawn on the left. */
		myEnd: End;
		players: PlayerSummary[];
		/** Move options the viewer can choose from (tappable on the board). */
		options?: MoveOption[];
		/** The delta of the option to emphasise (hovered in the decision panel). */
		highlight?: number | null;
		onchoose?: (delta: number) => void;
		onhover?: (delta: number | null) => void;
	} = $props();

	const S = 36; // distance between spaces
	const PAD = 30;
	const H = 172;
	const n = $derived(board.spaces.length);
	const W = $derived(2 * PAD + (n - 1) * S);

	// Geometry by display index d (0 = the viewer's end).
	const px = (d: number) => PAD + d * S;
	const py = (d: number) => 86 + 11 * Math.sin(d * 1.15 + 0.4);
	const above = (d: number) => d % 2 === 0;
	const ly = (d: number) => (above(d) ? 30 : H - 30);
	const disp = (i: number) => displayIndex(i, n, myEnd);

	const lastUncovered = (end: End) => {
		const idx = board.spaces.map((s, i) => (s.covered ? -1 : i)).filter((i) => i >= 0);
		return end === 'left' ? idx[0] : idx[idx.length - 1];
	};

	/** Where an option is drawn: its landing space, or just past the end it leaves at. */
	function optionPoint(o: MoveOption): { x: number; y: number } {
		if (!o.beyond) return { x: px(disp(o.to)), y: py(disp(o.to)) };
		const d = disp(lastUncovered(o.beyond));
		const outward = d < n / 2 ? -1 : 1;
		return { x: px(d) + outward * S * 0.85, y: py(d) };
	}

	const traveller = new Tween(
		{ x: 0, y: 0 },
		{ duration: () => (prefersReducedMotion.current ? 0 : 600), easing: cubicInOut }
	);
	let placed = false;
	$effect(() => {
		const d = disp(board.traveller);
		const target = { x: px(d), y: py(d) };
		// Jump on the first render, glide afterwards.
		void traveller.set(target, placed ? undefined : { duration: 0 });
		placed = true;
	});

	const endName = (end: End) => {
		const p = players.find((x) => x.end === end);
		return p ? (p.isYou ? 'You' : p.name) : '';
	};
	const leftEnd = $derived<End>(myEnd);
	const rightEnd = $derived<End>(myEnd === 'left' ? 'right' : 'left');

	function gemSpots(count: number): { x: number; y: number }[] {
		const shown = Math.min(count, 4);
		const grid = [
			[-5, -5],
			[5, -5],
			[-5, 5],
			[5, 5]
		];
		if (shown === 1) return [{ x: 0, y: 0 }];
		if (shown === 2)
			return [
				{ x: -5, y: 0 },
				{ x: 5, y: 0 }
			];
		return grid.slice(0, shown).map(([x, y]) => ({ x, y }));
	}

	function keyChoose(e: KeyboardEvent, delta: number) {
		if (e.key === 'Enter' || e.key === ' ') {
			e.preventDefault();
			onchoose?.(delta);
		}
	}
</script>

<figure class="flex flex-col gap-0.5">
	<figcaption class="flex justify-between px-1 text-xs font-semibold text-muted-foreground">
		<span>← {endName(leftEnd)}</span>
		<span>{endName(rightEnd)} →</span>
	</figcaption>
	<svg
		viewBox="0 0 {W} {H}"
		class="w-full rounded-xl border border-forest/30 bg-meadow"
		role="img"
		aria-label="Forest path: traveller on space {disp(board.traveller) +
			1} of {n} from your end, {board.gemsLeft} gems left"
	>
		<!-- A few trees for the forest edge -->
		<g class="fill-forest/25">
			<circle cx={8} cy={12} r={16} />
			<circle cx={W - 6} cy={H - 10} r={18} />
			<circle cx={W - 14} cy={8} r={12} />
			<circle cx={4} cy={H - 6} r={12} />
		</g>

		<!-- Path between the spaces -->
		<polyline
			points={Array.from({ length: n }, (_, d) => `${px(d)},${py(d)}`).join(' ')}
			class="fill-none stroke-path"
			stroke-width="5"
			stroke-linecap="round"
			stroke-linejoin="round"
		/>

		{#each board.spaces as space, i (i)}
			{@const d = disp(i)}
			{@const x = px(d)}
			{@const y = py(d)}
			{@const top = ly(d)}
			<g opacity={space.covered ? 0.4 : 1}>
				<line x1={x} y1={y} x2={x} y2={top} class="stroke-path" stroke-width="3" />
				<g transform="translate({x} {top}) rotate({above(d) ? -7 : 7})">
					<rect
						x="-15"
						y="-15"
						width="30"
						height="30"
						rx="6"
						class="fill-location stroke-ink/30"
						stroke-width="1.2"
					/>
					{#if space.plus}
						<g transform="translate(12 {above(d) ? -12 : 12})">
							<rect
								x="-5"
								y="-5"
								width="10"
								height="10"
								rx="2"
								class="fill-parchment stroke-ink/40"
							/>
							<path d="M-2.5 0h5M0 -2.5v5" class="stroke-ink/70" stroke-width="1.4" />
						</g>
					{/if}
					{#each gemSpots(space.gems) as g, k (k)}
						<path
							d="M{g.x} {g.y - 4.5}l4 3.2-4 5.8-4-5.8z"
							class="fill-gem stroke-ink/40"
							stroke-width="0.6"
						/>
					{/each}
					{#if space.gems > 4}
						<text x="0" y="13" text-anchor="middle" class="fill-ink text-[8px] font-bold">
							×{space.gems}
						</text>
					{/if}
				</g>
			</g>
			{#if space.covered}
				<g transform="translate({x} {y})">
					<circle r="13" class="fill-forest stroke-forest" />
					<path d="M-6 4l3-8 3 8zM0 5l3.5-10 3.5 10z" class="fill-moss" />
				</g>
			{:else}
				<circle cx={x} cy={y} r="10" class="fill-path stroke-ink/25" stroke-width="1" />
				{#if i === board.start}
					<circle cx={x} cy={y} r="14" class="fill-none stroke-path" stroke-width="2" />
				{/if}
			{/if}
		{/each}

		<!-- Move options the viewer can pick -->
		{#each options as o (o.delta)}
			{@const p = optionPoint(o)}
			{@const label = moveLabel({ players }, o)}
			<g
				role="button"
				tabindex="0"
				aria-label="{label.title}: {label.result}"
				class="cursor-pointer focus:outline-none"
				onclick={() => onchoose?.(o.delta)}
				onkeydown={(e) => keyChoose(e, o.delta)}
				onpointerenter={() => onhover?.(o.delta)}
				onpointerleave={() => onhover?.(null)}
			>
				<circle cx={p.x} cy={p.y} r="17" fill="transparent" />
				<circle
					cx={p.x}
					cy={p.y}
					r="13"
					class={o.lost ? 'fill-destructive/20 stroke-destructive' : 'fill-fox/20 stroke-fox'}
					stroke-width={highlight === o.delta ? 3.5 : 2}
					stroke-dasharray={highlight === o.delta ? undefined : '4 3'}
				/>
				{#if o.beyond}
					<path
						d="M{p.x - 4} {p.y - 4}l8 8M{p.x + 4} {p.y - 4}l-8 8"
						class={o.lost ? 'stroke-destructive' : 'stroke-fox'}
						stroke-width="2"
					/>
				{/if}
			</g>
		{/each}

		<!-- The traveller -->
		<g
			transform="translate({traveller.current.x} {traveller.current.y})"
			class="pointer-events-none"
		>
			<circle r="9" class="fill-fox stroke-parchment" stroke-width="2" />
			<path d="M0 -5v10M-5 0h10" class="stroke-parchment" stroke-width="1.6" />
			<circle r="2.6" class="fill-fox stroke-parchment" stroke-width="1.2" />
		</g>
	</svg>
</figure>
