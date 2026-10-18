<script lang="ts">
	import Copy from '@lucide/svelte/icons/copy';
	import Check from '@lucide/svelte/icons/check';
	import Share2 from '@lucide/svelte/icons/share-2';
	import type { Snippet } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { Button } from './ui/button/index.js';
	import { Input } from './ui/input/index.js';

	/**
	 * title is the game's name, shown as the card's heading and used for navigator.share.
	 * link is the absolute join URL (the game builds it with resolve('/join/[code]')).
	 * badge is shown right after the title (e.g. the edition or variant being set up).
	 */
	let { code, link, title, badge }: { code: string; link: string; title: string; badge?: Snippet } =
		$props();

	const canShare = typeof navigator !== 'undefined' && typeof navigator.share === 'function';

	let input = $state<HTMLInputElement | null>(null);
	let copied = $state<'code' | 'link' | null>(null);

	function selectLink() {
		input?.focus();
		input?.select();
	}

	async function writeClipboard(text: string): Promise<boolean> {
		if (globalThis.isSecureContext && navigator.clipboard?.writeText) {
			try {
				await navigator.clipboard.writeText(text);
				return true;
			} catch {
				// fall through to the legacy path
			}
		}
		return false;
	}

	function markCopied(what: 'code' | 'link') {
		copied = what;
		toast.success(what === 'code' ? 'Code copied' : 'Link copied');
		setTimeout(() => {
			if (copied === what) copied = null;
		}, 1500);
	}

	async function copyLink() {
		let ok = await writeClipboard(link);
		if (!ok) {
			selectLink();
			try {
				ok = document.execCommand('copy');
			} catch {
				ok = false;
			}
		}
		if (ok) markCopied('link');
		else toast('Long-press the link to copy it.');
	}

	async function copyCode() {
		if (await writeClipboard(code)) markCopied('code');
		else toast('Long-press the code to copy it.');
	}

	async function share() {
		try {
			await navigator.share({ title, text: `Join my game (${code})`, url: link });
		} catch {
			// cancelled or unsupported
		}
	}
</script>

<div class="flex flex-col gap-3">
	<div class="flex items-center gap-2">
		<div class="flex min-w-0 flex-1 items-center gap-2">
			<h2 class="min-w-0 truncate text-2xl font-semibold">{title}</h2>
			{@render badge?.()}
		</div>
		<span class="text-xs font-medium tracking-wide text-muted-foreground uppercase">Code</span>
		<span class="font-mono text-base font-bold tracking-[0.2em] select-all">{code}</span>
		<Button
			variant="ghost"
			size="icon-sm"
			class="-ml-1.5"
			aria-label="Copy code"
			title="Copy code"
			onclick={copyCode}
		>
			{#if copied === 'code'}<Check />{:else}<Copy />{/if}
		</Button>
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
		<Button variant="secondary" class="h-11 flex-1 text-base" onclick={copyLink}>
			{#if copied === 'link'}<Check />{:else}<Copy />{/if}
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
