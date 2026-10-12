<script lang="ts">
	import './layout.css';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import ArrowLeft from '@lucide/svelte/icons/arrow-left';
	import HeaderButton from '@board-games/shared/HeaderButton.svelte';
	import PageFrame from '@board-games/shared/PageFrame.svelte';
	import { siteName } from '#lib/site.js';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<!-- No onToggleTheme: the page ships no JS, the theme script in app.html handles the toggle. -->
<!-- Trailing slashes: the pages are <name>/index.html (see datenschutz/+page.ts). -->
<PageFrame imprintHref="{resolve('/impressum')}/" privacyHref="{resolve('/datenschutz')}/">
	{#snippet start()}
		<!-- The legal pages: the way back sits in the top bar, next to the theme toggle. -->
		{#if page.route.id !== '/'}
			<HeaderButton icon={ArrowLeft} label={siteName} showLabel href={resolve('/')} />
		{/if}
	{/snippet}
	{@render children()}
</PageFrame>
