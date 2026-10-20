<script lang="ts">
	import { cubicInOut } from 'svelte/easing';
	import { prefersReducedMotion, Tween } from 'svelte/motion';
	import type { Marks } from '#lib/game/boardMarks.svelte.js';
	import { displayIndex, moveLabel, plural } from '#lib/game/format.js';
	import { cn } from '#lib/utils.js';
	import type { Board, End, MoveOption, PlayerSummary } from '#lib/types.js';

	let {
		board,
		myEnd,
		players,
		online = [],
		marks = { added: {}, covered: [] },
		options = [],
		highlight = null,
		onchoose,
		onhover
	}: {
		board: Board;
		/** The viewer's end; it is drawn on the left. */
		myEnd: End;
		players: PlayerSummary[];
		/** Ids of the players who are connected. */
		online?: string[];
		/** What changed since the last card was led (gems added, new forest tokens). */
		marks?: Marks;
		/** Move options the viewer can choose from (tappable on the board). */
		options?: MoveOption[];
		/** The delta of the option to emphasise (hovered in the action area). */
		highlight?: number | null;
		onchoose?: (delta: number) => void;
		onhover?: (delta: number | null) => void;
	} = $props();

	const H = 160;
	const PAD = 18;
	const TILE = 26;
	let W = $state(0);
	const n = $derived(board.spaces.length);
	const step = $derived((W - 2 * PAD) / (n - 1));

	// Geometry by display index d (0 = the viewer's end): the path meanders once up and down,
	// each location sits above the path or, with a plus, below it.
	const px = (d: number) => PAD + d * step;
	const py = (d: number) => H / 2 + 21 * Math.sin((2 * Math.PI * d) / (n - 1));
	const disp = (i: number) => displayIndex(i, n, myEnd);
	const tileY = (i: number) => py(disp(i)) + (board.spaces[i].plus ? 40 : -40);
	const tilt = (d: number) => [-7, 5, -3][d % 3];

	/** A smooth curve through the spaces. */
	const pathD = $derived.by(() => {
		if (W === 0) return '';
		const pts = Array.from({ length: n }, (_, d) => [px(d), py(d)]);
		let out = `M${pts[0][0]} ${pts[0][1]}`;
		for (let k = 0; k < n - 1; k++) {
			const [p0, p1, p2, p3] = [pts[k - 1] ?? pts[k], pts[k], pts[k + 1], pts[k + 2] ?? pts[k + 1]];
			out += ` C${p1[0] + (p2[0] - p0[0]) / 6} ${p1[1] + (p2[1] - p0[1]) / 6} ${p2[0] - (p3[0] - p1[0]) / 6} ${p2[1] - (p3[1] - p1[1]) / 6} ${p2[0]} ${p2[1]}`;
		}
		return out;
	});

	const lastUncovered = (end: End) => {
		const idx = board.spaces.map((s, i) => (s.covered ? -1 : i)).filter((i) => i >= 0);
		return end === 'left' ? idx[0] : idx[idx.length - 1];
	};

	/** Where an option is drawn: its landing space, or just past the end it leaves at. */
	function optionPoint(o: MoveOption): { x: number; y: number } {
		if (!o.beyond) return { x: px(disp(o.to)), y: py(disp(o.to)) };
		const d = disp(lastUncovered(o.beyond));
		const outward = d < n / 2 ? -1 : 1;
		return { x: px(d) + outward * step * 0.85, y: py(d) };
	}

	const traveller = new Tween(
		{ x: 0, y: 0 },
		{ duration: () => (prefersReducedMotion.current ? 0 : 600), easing: cubicInOut }
	);
	let placedAt = 0;
	$effect(() => {
		if (W === 0) return;
		const d = disp(board.traveller);
		// Jump on the first render and when the board resizes, glide afterwards.
		void traveller.set({ x: px(d), y: py(d) }, placedAt === W ? undefined : { duration: 0 });
		placedAt = W;
	});

	const mate = $derived(players.find((p) => !p.isYou));
	const tokens = $derived(
		Array.from({ length: board.forestTokensTotal }, (_, i) => i < board.forestTokens)
	);

	function keyChoose(e: KeyboardEvent, delta: number) {
		if (e.key === 'Enter' || e.key === ' ') {
			e.preventDefault();
			onchoose?.(delta);
		}
	}
</script>

{#snippet gem(x: number, y: number)}
	<path
		d="M{x} {y - 6.5}l4.5 4.3-4.5 8.7-4.5-8.7z"
		class="fill-gem stroke-ink/35"
		stroke-width="0.6"
	/>
{/snippet}

<figure class="flex flex-col gap-1">
	<figcaption
		class="flex h-5 items-center justify-between gap-2 text-xs whitespace-nowrap text-muted-foreground"
	>
		<span class="min-w-0 truncate">← <b class="font-extrabold text-foreground">You</b></span>
		<span class="flex items-center gap-2.5">
			<span class="flex items-center gap-1" title="Gems left in the forest">
				<svg viewBox="0 0 24 24" class="h-3.5 w-3 fill-gem" aria-hidden="true">
					<path d="M12 2 21 10 12 22 3 10Z" />
				</svg>
				<b class="font-extrabold text-foreground">{board.gemsLeft}</b> left
			</span>
			<span
				class="flex items-center gap-0.5"
				title="{plural(board.forestTokens, 'forest token')} left of {board.forestTokensTotal}"
			>
				{#each tokens as left, i (i)}
					<span
						class={cn(
							'size-2.5 rounded-full border-2',
							left ? 'border-wood bg-wood' : 'border-muted-foreground/60'
						)}
					></span>
				{/each}
				<span class="sr-only">{plural(board.forestTokens, 'forest token')} left</span>
			</span>
		</span>
		<span class="flex min-w-0 items-center gap-1.5">
			{#if mate}
				<span
					class={cn(
						'size-[7px] shrink-0 rounded-full',
						online.includes(mate.id) ? 'bg-online' : 'bg-muted-foreground/40'
					)}
					title={online.includes(mate.id) ? 'Online' : 'Offline'}
				></span>
				<span class="truncate"><b class="font-extrabold text-foreground">{mate.name}</b> →</span>
			{/if}
		</span>
	</figcaption>

	<div class="h-40" bind:clientWidth={W}>
		{#if W > 0}
			<svg
				width={W}
				height={H}
				viewBox="0 0 {W} {H}"
				class="block"
				role="img"
				aria-label="Forest path: traveller on space {disp(board.traveller) +
					1} of {n} from your end, {plural(board.gemsLeft, 'gem')} left"
			>
				<rect
					x="0.5"
					y="0.5"
					width={W - 1}
					height={H - 1}
					rx="14"
					class="fill-meadow stroke-border"
				/>
				<!-- The forest edge in the corners, as on the printed board -->
				<g class="fill-forest" opacity="0.22">
					<circle cx="6" cy="8" r="20" />
					<circle cx="28" cy="-4" r="14" />
					<circle cx={W - 4} cy={H - 6} r="22" />
					<circle cx={W - 30} cy={H + 6} r="14" />
					<circle cx={W - 8} cy="10" r="12" />
					<circle cx="4" cy={H - 4} r="12" />
				</g>

				<!-- Dotted spurs from the spaces to their locations -->
				{#each board.spaces as space, i (i)}
					<line
						x1={px(disp(i))}
						y1={py(disp(i))}
						x2={px(disp(i))}
						y2={tileY(i)}
						class="stroke-path"
						stroke-width="3"
						stroke-dasharray="1 4"
						stroke-linecap="round"
						opacity={space.covered ? 0.4 : 1}
					/>
				{/each}

				<path d={pathD} class="fill-none stroke-path" stroke-width="2.5" stroke-linecap="round" />
				{#each Array.from({ length: n - 1 }, (_, d) => d) as d (d)}
					<circle cx={(px(d) + px(d + 1)) / 2} cy={py(d + 0.5)} r="2.6" class="fill-path" />
				{/each}

				<!-- Locations: a tile with its gems, a plus where gems come at round end -->
				{#each board.spaces as space, i (i)}
					{@const d = disp(i)}
					{@const added = space.covered ? 0 : (marks.added[i] ?? 0)}
					<g transform="translate({px(d)} {tileY(i)})" opacity={space.covered ? 0.4 : 1}>
						<g transform="rotate({tilt(d)})">
							<rect
								x={-TILE / 2}
								y={-TILE / 2}
								width={TILE}
								height={TILE}
								rx="4"
								class="fill-tile stroke-ink/25"
								stroke-width="1.2"
							/>
							<rect
								x={-TILE / 2 + 3}
								y={-TILE / 2 + 3}
								width={TILE - 6}
								height={TILE - 6}
								rx="2"
								class="fill-none stroke-ink/12"
								stroke-dasharray="2 2"
							/>
						</g>
						{#if !space.covered}
							{#if space.gems === 1}
								{@render gem(0, 0)}
							{:else if space.gems === 2}
								{@render gem(-5, 0)}
								{@render gem(5, 0)}
							{:else if space.gems > 2}
								{@render gem(-5, 0)}
								<text x="6" y="4.5" text-anchor="middle" class="fill-ink text-xs font-extrabold"
									>{space.gems}</text
								>
							{/if}
						{/if}
						{#if added > 0}
							<g transform="translate({TILE / 2 - 1} {TILE / 2 - 1})">
								<rect x="-10" y="-7" width="20" height="14" rx="7" class="fill-fox" />
								<text
									y="3.5"
									text-anchor="middle"
									class="fill-primary-foreground text-[10px] font-extrabold">+{added}</text
								>
							</g>
						{:else if space.plus && !space.covered}
							<g transform="translate({TILE / 2 - 1} {TILE / 2 - 1})">
								<rect
									x="-5.5"
									y="-5.5"
									width="11"
									height="11"
									rx="2"
									class="fill-ribbon stroke-ink/40"
								/>
								<path d="M-3 0h6M0 -3v6" class="stroke-ink" stroke-width="1.5" />
							</g>
						{/if}
					</g>
				{/each}

				<!-- Spaces: a stepping stone, or a wooden forest token where the path is covered -->
				{#each board.spaces as space, i (i)}
					{@const x = px(disp(i))}
					{@const y = py(disp(i))}
					{#if space.covered}
						<circle cx={x} cy={y} r="11" class="fill-wood stroke-ink/35" />
						<circle cx={x} cy={y} r="7.5" class="fill-none stroke-ink/20" />
						<path
							transform="translate({x - 7} {y - 7}) scale(0.5833)"
							d="M12 2 5 11h3.5L4 17.5h6.5V22h3v-4.5H20L15.5 11H19Z"
							class="fill-forest"
						/>
						{#if marks.covered.includes(i)}
							<circle cx={x} cy={y} r="15" class="fill-none stroke-fox" stroke-width="2.5" />
						{/if}
					{:else}
						<circle cx={x} cy={y} r="7.5" class="fill-path stroke-ink/20" stroke-width="1.2" />
						{#if i === board.start}
							<circle cx={x} cy={y} r="11.5" class="fill-none stroke-path" stroke-width="2" />
						{/if}
					{/if}
				{/each}

				<!-- Move options the viewer can pick -->
				{#each options as o (o.delta)}
					{@const p = optionPoint(o)}
					{@const label = moveLabel({ players }, o)}
					{@const on = highlight === o.delta}
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
						<circle cx={p.x} cy={p.y} r={Math.min(22, step / 2)} fill="transparent" />
						<circle
							cx={p.x}
							cy={p.y}
							r="13"
							class={o.lost ? 'fill-destructive/20 stroke-destructive' : 'fill-fox/25 stroke-fox'}
							stroke-width={on ? 3.5 : 2.5}
							stroke-dasharray={on ? undefined : '4 3'}
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

				<!-- The traveller: the fox on an orange disc -->
				<g
					transform="translate({traveller.current.x} {traveller.current.y})"
					class="pointer-events-none"
				>
					<circle r="11" class="fill-fox stroke-parchment" stroke-width="2" />
					<g transform="translate(-7 -7) scale(0.5833)">
						<path
							d="M2 3 8.6 8.6h6.8L22 3l-1 9.2L12 21.5 3 12.2Z"
							class="fill-primary-foreground"
						/>
						<circle cx="8.8" cy="12.6" r="1.2" class="fill-fox" />
						<circle cx="15.2" cy="12.6" r="1.2" class="fill-fox" />
					</g>
				</g>
			</svg>
		{/if}
	</div>
</figure>
