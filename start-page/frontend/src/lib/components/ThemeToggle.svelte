<script lang="ts">
	import Sun from '@lucide/svelte/icons/sun';
	import Moon from '@lucide/svelte/icons/moon';
	import { Button } from '#lib/components/ui/button/index.js';
</script>

<!-- Static markup: the click is handled by the theme script in app.html (the page ships no framework JS). -->
<Button variant="ghost" size="icon-sm" data-theme-toggle aria-label="Toggle light or dark mode">
	<Moon class="size-4 dark:hidden" />
	<Sun class="hidden size-4 dark:block" />
</Button>
