<!-- Solo only: the Jester power (discard your hand, draw a new one), behind a confirmation. -->
<script lang="ts">
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import type { GameClient } from '#lib/game/client.js';
	import type { Runner } from '#lib/game/selection.js';
	import type { PlayerView } from '#lib/types.js';
	import JesterHat from './JesterHat.svelte';

	let {
		client,
		view,
		busy,
		connected,
		run
	}: { client: GameClient; view: PlayerView; busy: boolean; connected: boolean; run: Runner } =
		$props();

	let open = $state(false);
</script>

{#if view.soloJestersLeft > 0}
	<Button
		variant="outline"
		class="h-11 gap-1 border-jester/60 px-3 text-base text-jester"
		disabled={!connected || busy}
		title={`Use a Jester power (${view.soloJestersLeft} left)`}
		aria-label={`Use a Jester power (${view.soloJestersLeft} left)`}
		onclick={() => (open = true)}
	>
		<JesterHat class="size-5" />
		{view.soloJestersLeft}
	</Button>

	<AlertDialog.Root bind:open>
		<AlertDialog.Content>
			<AlertDialog.Header>
				<AlertDialog.Title>Use a Jester power?</AlertDialog.Title>
				<AlertDialog.Description>
					Discard your hand and draw up to {view.rules.maxHandSize} cards.
				</AlertDialog.Description>
			</AlertDialog.Header>
			<AlertDialog.Footer>
				<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
				<AlertDialog.Action
					onclick={() => {
						open = false;
						run(() => client.useSoloJester());
					}}
				>
					Use Jester
				</AlertDialog.Action>
			</AlertDialog.Footer>
		</AlertDialog.Content>
	</AlertDialog.Root>
{/if}
