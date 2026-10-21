<script lang="ts">
	import SiteHeader from '@board-games/shared/SiteHeader.svelte';
	import { siteName } from '@board-games/shared/site.js';
	import { Button } from '@board-games/shared/ui/button/index.js';
	import { Input } from '@board-games/shared/ui/input/index.js';

	// The gate renders the built page as a Go html/template (backend/internal/gate/ui.go): the
	// actions below stay as text in the HTML and are filled in per request, escaped. In dev the page
	// is not a template, so it gets plain values instead.
	const dev = import.meta.env.DEV;
	const action = dev ? '#' : '{{.Action}}';
	const next = dev ? '/board-games/' : '{{.Next}}';
</script>

<svelte:head>
	<title>{siteName}</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<!-- The start page's header in the same place, and the form where its game list will be. -->
<SiteHeader />
<form
	method="post"
	{action}
	class="mx-auto flex w-full max-w-sm flex-col gap-2 rounded-2xl bg-card p-5 text-card-foreground shadow-md ring-1 shadow-tile-shadow ring-border"
>
	<label for="password" class="text-sm font-medium">Password</label>
	<!-- A masked text field, not type=password: browsers would offer to save the shared password. -->
	<Input
		id="password"
		name="password"
		type="text"
		autocomplete="off"
		autocapitalize="off"
		spellcheck="false"
		required
		autofocus
		class="[-webkit-text-security:disc]"
	/>
	<input type="hidden" name="next" value={next} />
	<Button type="submit" class="mt-2 w-full">Enter</Button>
	{#if !dev}
		<!-- Shown only when the gate passes a message (wrong password, too many attempts). -->
		{'{{if .Message}}'}
		<p class="text-sm text-destructive" role="alert">{'{{.Message}}'}</p>
		{'{{end}}'}
	{/if}
</form>
