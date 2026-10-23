<!--
	A legal page as a native dialog over the page it was opened from, in that app's theme. PageFrame
	renders one per page next to its footer links (which stay plain links to the start page's pages
	as the fallback). Opening and closing on a backdrop tap: PageFrame's handlers in the games, the
	inline script in app.html on the pages that ship no JS (start page, gate). Esc and the close
	button (a dialog form) need no JS. The ids are what the footer links name in `data-legal`.
-->
<script lang="ts">
	import X from '@lucide/svelte/icons/x';
	import { Button } from '../ui/button/index.js';
	import LegalText from './LegalText.svelte';
	import { legalDialogId, legalTitle, type LegalPage } from './legal.js';

	let { page }: { page: LegalPage } = $props();

	const id = $derived(legalDialogId(page));

	/** A tap on the backdrop lands on the dialog itself (its content fills it). */
	function closeOnBackdrop(e: MouseEvent & { currentTarget: HTMLDialogElement }) {
		if (e.target === e.currentTarget) e.currentTarget.close();
	}
</script>

<!-- The click handler only catches a backdrop tap; Esc closes natively. -->
<dialog
	{id}
	aria-labelledby="{id}-title"
	class="legal-dialog m-auto h-fit max-h-[85dvh] w-[calc(100%-2rem)] max-w-lg flex-col overflow-hidden rounded-xl bg-popover p-0 text-popover-foreground ring-1 ring-foreground/10 backdrop:bg-black/30 open:flex"
	onclick={closeOnBackdrop}
>
	<header class="flex items-center gap-3 border-b py-2 pr-2 pl-6">
		<h2 id="{id}-title" class="flex-1 text-xl font-semibold">{legalTitle[page]}</h2>
		<form method="dialog">
			<Button type="submit" variant="ghost" size="icon" class="size-11">
				<X />
				<span class="sr-only">Close</span>
			</Button>
		</form>
	</header>
	<div class="min-h-0 overflow-y-auto px-6 pt-2 pb-6">
		<LegalText {page} class="text-sm" />
	</div>
</dialog>

<style>
	/* The page behind doesn't scroll while a dialog is open. */
	:global(html:has(.legal-dialog[open])) {
		overflow: hidden;
	}
</style>
