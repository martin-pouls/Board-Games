<!--
	The light/dark toggle for a game's own header. Game pages hide the shared top bar (AppFrame
	topBar={false}) to save a row on phones, so every header on them (lobby, table) shows this one.
	Same icons as ThemeToggle, sized like the other header icons.
-->
<script lang="ts">
	import Moon from '@lucide/svelte/icons/moon';
	import Sun from '@lucide/svelte/icons/sun';
	import { toggleMode } from 'mode-watcher';
	import { cn } from './utils.js';

	let { class: className }: { class?: string } = $props();
</script>

<button
	type="button"
	class={cn('rounded-md p-1 text-muted-foreground hover:text-foreground', className)}
	onclick={toggleMode}
	aria-label="Toggle light or dark mode"
	title="Toggle light or dark mode"
>
	<Moon class="size-5 dark:hidden" />
	<Sun class="hidden size-5 dark:block" />
</button>
