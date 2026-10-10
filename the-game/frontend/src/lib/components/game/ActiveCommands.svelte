<script lang="ts">
	import { toast } from 'svelte-sonner';
	import ArrowUp from '@lucide/svelte/icons/arrow-up';
	import ArrowDown from '@lucide/svelte/icons/arrow-down';
	import { commandInfo } from '#lib/game/commands.js';
	import type { ActiveCommand, Pile } from '#lib/types.js';
	import CommandChip from './CommandChip.svelte';

	/** The visible Infinity commands; tap one for its explanation. */
	let { commands, piles }: { commands: ActiveCommand[]; piles: Pile[] } = $props();

	function explain(c: ActiveCommand) {
		const info = commandInfo(c.command);
		toast(info.label, { id: `command-${c.command}`, description: info.description });
	}
</script>

{#if commands.length > 0}
	<ul class="flex flex-wrap gap-1" aria-label="Active commands">
		{#each commands as c (c.pile)}
			{@const info = commandInfo(c.command)}
			{@const up = piles.find((p) => p.index === c.pile)?.direction === 'up'}
			<li class="min-w-0">
				<button
					type="button"
					class="flex max-w-full [touch-action:manipulation] rounded-full [-webkit-tap-highlight-color:transparent] focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
					aria-label={`${info.label} on a ${up ? 'ascending' : 'descending'} pile: ${info.description}`}
					onclick={() => explain(c)}
				>
					<CommandChip command={c.command}>
						{#if up}
							<ArrowUp class="size-3 shrink-0 opacity-70" aria-hidden="true" />
						{:else}
							<ArrowDown class="size-3 shrink-0 opacity-70" aria-hidden="true" />
						{/if}
					</CommandChip>
				</button>
			</li>
		{/each}
	</ul>
{/if}
