<script lang="ts">
	import './layout.css';
	import { resolve } from '$app/paths';
	import PageFrame from '@board-games/shared/PageFrame.svelte';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<!-- No onToggleTheme: the page ships no JS, the theme script in app.html handles the toggle. -->
<!-- Trailing slashes: the pages are <name>/index.html (see datenschutz/+page.ts). -->
<PageFrame imprintHref="{resolve('/impressum')}/" privacyHref="{resolve('/datenschutz')}/">
	{@render children()}
</PageFrame>
