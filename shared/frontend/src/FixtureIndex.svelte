<script lang="ts">
	/**
	 * Dev only: the list on a game's `/dev/fixtures` route. Each link opens one fixture view on
	 * the game screen, with the in-game layout (screenshots, visual review).
	 */
	let { names, href }: { names: string[]; href: (name: string) => string } = $props();
</script>

<div class="mx-auto flex w-full max-w-md flex-col gap-3">
	<h1 class="text-2xl font-semibold">Fixtures</h1>
	<p class="text-sm text-muted-foreground">
		Dev only. Sample views from <code>src/lib/game/fixtures.ts</code> on the game screen. Buttons do nothing
		(commands fail with "disconnected").
	</p>
	<ul class="flex flex-col divide-y divide-border">
		{#each names as name (name)}
			<li><a class="block py-2 font-mono text-sm hover:underline" href={href(name)}>{name}</a></li>
		{/each}
	</ul>
</div>
