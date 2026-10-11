<script lang="ts">
	import { mode, toggleMode } from 'mode-watcher';
	import Sun from '@lucide/svelte/icons/sun';
	import Moon from '@lucide/svelte/icons/moon';
	import { Button } from './ui/button/index.js';
</script>

<Button
	variant="ghost"
	size="icon-sm"
	onclick={toggleMode}
	aria-label={mode.current === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
	title={mode.current === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
>
	{#if mode.current === 'dark'}
		<Sun class="size-4" />
	{:else}
		<Moon class="size-4" />
	{/if}
</Button>
