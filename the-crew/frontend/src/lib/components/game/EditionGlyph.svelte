<script lang="ts">
	import LifeBuoy from '@lucide/svelte/icons/life-buoy';
	import Radio from '@lucide/svelte/icons/radio';
	import Rocket from '@lucide/svelte/icons/rocket';
	import { edition } from '#lib/edition.js';
	import SonarPing from './SonarPing.svelte';
	import SuitGlyph from './SuitGlyph.svelte';

	// The icon of an idea that differs per edition: the mission (rocket / submarine), communication
	// (radio / sonar ping) and the distress signal (a buoy in Deep Sea; Planet Nine has no icon).
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
		<Radio class={className} aria-label={label} />
	{/if}
{:else if edition === 'deepSea'}
	<LifeBuoy class={className} aria-label={label} />
{/if}
