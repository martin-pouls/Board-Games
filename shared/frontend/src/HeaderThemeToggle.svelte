<!--
	The light/dark toggle for a game's own header. Game pages hide the shared top bar (AppFrame
	topBar={false}) to save a row on phones, so every header on them (lobby, table) shows this one:
	the top bar's ThemeToggle, wired to mode-watcher.
-->
<script lang="ts">
	import { toggleMode } from 'mode-watcher';
	import ThemeToggle from './ThemeToggle.svelte';

	let { class: className }: { class?: string } = $props();
</script>

<ThemeToggle onclick={toggleMode} class={className} />
