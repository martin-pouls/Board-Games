<!--
	The light/dark toggle. The icon follows the `dark` class on <html> (CSS only), so it is right
	from the first paint and also in pages that ship no JS (the start page, the gate's login): there
	the click is handled by the theme script in their app.html, which listens for
	[data-theme-toggle]. The games pass mode-watcher's toggleMode as onclick. Sized like HeaderButton.
-->
<script lang="ts">
	import Sun from '@lucide/svelte/icons/sun';
	import Moon from '@lucide/svelte/icons/moon';
	import { Button } from './ui/button/index.js';
	import { cn } from './utils.js';

	let { onclick, class: className }: { onclick?: () => void; class?: string } = $props();
</script>

<Button
	variant="ghost"
	size="icon-sm"
	class={cn('text-muted-foreground', className)}
	data-theme-toggle
	{onclick}
	aria-label="Toggle light or dark mode"
	title="Toggle light or dark mode"
>
	<Moon class="size-4 dark:hidden" />
	<Sun class="hidden size-4 dark:block" />
</Button>
