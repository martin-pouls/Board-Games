<script lang="ts">
	import { resolve } from '$app/paths';
	import FixtureIndex from '@board-games/shared/FixtureIndex.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<svelte:head>
	<title>Fixtures</title>
</svelte:head>

<FixtureIndex names={data.names} href={(name) => resolve('/dev/fixtures/[name]', { name })} />
