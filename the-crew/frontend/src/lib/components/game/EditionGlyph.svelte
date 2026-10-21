<script lang="ts">
	import LifeBuoy from '@lucide/svelte/icons/life-buoy';
	import Rocket from '@lucide/svelte/icons/rocket';
	import { edition } from '#lib/edition.js';
	import SonarPing from './SonarPing.svelte';
	import SuitGlyph from './SuitGlyph.svelte';

	// The icon of an idea that differs per edition: the mission (rocket / submarine), communication
	// (the radio: a green radio token with its waves / sonar ping) and the distress signal (a buoy in Deep Sea; Planet Nine has no icon).
	let {
		kind,
		class: className = '',
		'aria-label': label
	}: { kind: 'mission' | 'comm' | 'distress'; class?: string; 'aria-label'?: string } = $props();
</script>

{#if kind === 'mission'}
	{#if edition === 'deepSea'}
		<SuitGlyph suit="trump" class={className} />
	{:else}
		<Rocket class={className} aria-label={label} />
	{/if}
{:else if kind === 'comm'}
	{#if edition === 'deepSea'}
		<SonarPing class={className} aria-label={label} />
	{:else}
		<svg
			viewBox="0 0 24 24"
			class={className}
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			stroke-linecap="round"
			role={label ? 'img' : undefined}
			aria-label={label}
			aria-hidden={label ? undefined : 'true'}
		>
			<circle cx="12" cy="12" r="3.2" fill="var(--radio)" stroke="none" />
			<path
				d="M8.2 15.8a5.4 5.4 0 0 1 0-7.6M15.8 8.2a5.4 5.4 0 0 1 0 7.6M5.3 18.7a9.5 9.5 0 0 1 0-13.4M18.7 5.3a9.5 9.5 0 0 1 0 13.4"
			/>
		</svg>
	{/if}
{:else if edition === 'deepSea'}
	<LifeBuoy class={className} aria-label={label} />
{/if}
