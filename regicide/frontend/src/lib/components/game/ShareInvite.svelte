<script lang="ts">
	import Copy from '@lucide/svelte/icons/copy';
	import Check from '@lucide/svelte/icons/check';
	import Share2 from '@lucide/svelte/icons/share-2';
	import { resolve } from '$app/paths';
	import { toast } from 'svelte-sonner';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';

	let { code }: { code: string } = $props();

	const link = $derived(`${globalThis.location?.origin ?? ''}${resolve('/join/[code]', { code })}`);
	const canShare = typeof navigator !== 'undefined' && typeof navigator.share === 'function';

	let input = $state<HTMLInputElement | null>(null);
	let copied = $state(false);

	function selectLink() {
		input?.focus();
		input?.select();
	}

	async function copy() {
		let ok = false;
		if (globalThis.isSecureContext && navigator.clipboard?.writeText) {
			try {
				await navigator.clipboard.writeText(link);
				ok = true;
			} catch {
				// fall through to the legacy path
			}
		}
		if (!ok) {
			selectLink();
			try {
				ok = document.execCommand('copy');
			} catch {
				ok = false;
			}
		}
		if (ok) {
			copied = true;
			toast.success('Link copied');
			setTimeout(() => (copied = false), 1500);
		} else {
			toast('Long-press the link to copy it.');
		}
	}

	async function share() {
		try {
			await navigator.share({ title: 'Regicide', text: `Join my game (${code})`, url: link });
		} catch {
			// cancelled or unsupported
		}
	}
</script>

<div class="flex flex-col gap-3">
	<div class="text-center">
		<p class="text-xs font-medium tracking-wide text-muted-foreground uppercase">Game code</p>
		<p class="font-mono text-4xl font-bold tracking-[0.3em]">{code}</p>
	</div>
	<Input
		bind:ref={input}
		readonly
		value={link}
		aria-label="Invite link"
		class="h-10 font-mono text-base md:text-sm"
		onfocus={selectLink}
	/>
	<div class="flex gap-2">
		<Button variant="secondary" class="h-11 flex-1 text-base" onclick={copy}>
			{#if copied}<Check />{:else}<Copy />{/if}
			Copy link
		</Button>
		{#if canShare}
			<Button variant="secondary" class="h-11 flex-1 text-base" onclick={share}>
				<Share2 />
				Share
			</Button>
		{/if}
	</div>
</div>
