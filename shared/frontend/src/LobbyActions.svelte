<!--
	The lobby header's way out, the same in every game: alone in the lobby the host can close it
	("Close lobby", X, no confirmation: it affects nobody else); otherwise everyone can "Leave"
	(DoorOpen, confirmation) and the host can "End game" for everyone (Flag, confirmation). The
	buttons are plain header buttons; red appears only on the confirm button inside the dialogs.
	Put it in the lobby's header row, before the theme toggle.
-->
<script lang="ts">
	import DoorOpen from '@lucide/svelte/icons/door-open';
	import Flag from '@lucide/svelte/icons/flag';
	import X from '@lucide/svelte/icons/x';
	import HeaderButton from './HeaderButton.svelte';
	import * as AlertDialog from './ui/alert-dialog/index.js';

	let {
		alone,
		isHost,
		connected,
		busy = false,
		onend,
		onleave
	}: {
		/** The player is the only one in the lobby. */
		alone: boolean;
		isHost: boolean;
		/** The live connection is open (ending the lobby needs it). */
		connected: boolean;
		/** The lobby is busy (e.g. starting the game): the buttons are disabled. */
		busy?: boolean;
		/** Closes the lobby for everyone (host only) and leaves the page. */
		onend: () => Promise<void> | void;
		/** Removes the player from the lobby and leaves the page. */
		onleave: () => Promise<void> | void;
	} = $props();

	let confirmLeave = $state(false);
	let confirmEnd = $state(false);
	/** A leave or end is in flight. */
	let working = $state(false);

	const disabled = $derived(busy || working);

	async function finish(action: () => Promise<void> | void, close: () => void) {
		if (disabled) return;
		working = true;
		try {
			await action();
		} finally {
			working = false;
			close();
		}
	}

	const end = () =>
		finish(onend, () => {
			confirmEnd = false;
		});
	const leave = () =>
		finish(onleave, () => {
			confirmLeave = false;
		});
</script>

{#if alone}
	<!-- A sole player cannot leave; closing affects nobody else, so no confirmation. -->
	<HeaderButton
		icon={X}
		label="Close lobby"
		showLabel
		disabled={!connected || disabled}
		onclick={end}
	/>
{:else}
	<HeaderButton
		icon={DoorOpen}
		label="Leave"
		showLabel
		{disabled}
		onclick={() => (confirmLeave = true)}
	/>
	{#if isHost}
		<HeaderButton
			icon={Flag}
			label="End game"
			showLabel
			disabled={!connected || disabled}
			onclick={() => (confirmEnd = true)}
		/>
	{/if}
{/if}

<AlertDialog.Root bind:open={confirmLeave}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Leave this game?</AlertDialog.Title>
			<AlertDialog.Description>
				You will be removed from the lobby. You can rejoin with the code while it is still open.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel {disabled}>Stay</AlertDialog.Cancel>
			<AlertDialog.Action
				class="bg-destructive text-white hover:bg-destructive/90"
				{disabled}
				onclick={leave}
			>
				Leave
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>

<AlertDialog.Root bind:open={confirmEnd}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>End this game for everyone?</AlertDialog.Title>
			<AlertDialog.Description>
				The lobby closes before the game starts and moves to everyone's history.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel {disabled}>Keep lobby</AlertDialog.Cancel>
			<AlertDialog.Action
				class="bg-destructive text-white hover:bg-destructive/90"
				disabled={disabled || !connected}
				onclick={end}
			>
				End game
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
