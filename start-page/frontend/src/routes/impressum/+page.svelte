<script lang="ts">
	import LegalText from '@board-games/shared/legal/LegalText.svelte';
	import { legalTitle } from '@board-games/shared/legal/legal.js';
	import { siteName } from '@board-games/shared/site.js';
</script>

<svelte:head>
	<title>Impressum · {siteName}</title>
</svelte:head>

<!-- The public copy of the imprint (the footers open it as a dialog; this is the way without JS). -->
<article lang="de" class="mt-2 flex flex-col gap-3">
	<h1 class="text-3xl font-bold tracking-tight sm:text-4xl">{legalTitle.impressum}</h1>
	<LegalText page="impressum" />
</article>
