<script lang="ts">
	import './layout.css';
	import { page } from '$app/state';
	import PageFrame from '@board-games/shared/PageFrame.svelte';
	import { privacyHref } from '@board-games/shared/startPage.js';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
	const dev = import.meta.env.DEV;
</script>

<!-- No onToggleTheme: the page ships no JS, the theme script in app.html handles the toggle. -->
<PageFrame privacyHref={privacyHref(page.url, dev)}>
	{@render children()}
</PageFrame>
