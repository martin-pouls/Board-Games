<script lang="ts">
	import './layout.css';
	import PageFrame from '@board-games/shared/PageFrame.svelte';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<!-- No onToggleTheme: the page ships no JS, the theme script in app.html handles the toggle. -->
<PageFrame>
	{@render children()}
</PageFrame>
