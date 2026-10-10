<script lang="ts">
	import { onMount } from 'svelte';
	import { getHealth } from '#lib/api.js';
	import { Badge } from '#lib/components/ui/badge/index.js';
	import * as Card from '#lib/components/ui/card/index.js';

	let backend = $state<'checking' | 'ok' | 'down'>('checking');

	onMount(async () => {
		try {
			backend = (await getHealth()).status === 'ok' ? 'ok' : 'down';
		} catch {
			backend = 'down';
		}
	});
</script>

<div class="flex flex-1 items-center justify-center">
	<Card.Root class="w-full max-w-md">
		<Card.Header>
			<Card.Title class="text-3xl">The Game</Card.Title>
			<Card.Description>Play as a team — as long as you can.</Card.Description>
		</Card.Header>
		<Card.Content class="flex items-center gap-2 text-sm">
			Backend:
			<Badge
				variant={backend === 'ok' ? 'default' : backend === 'down' ? 'destructive' : 'secondary'}
			>
				{backend}
			</Badge>
		</Card.Content>
	</Card.Root>
</div>
