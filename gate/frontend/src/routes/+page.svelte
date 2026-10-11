<script lang="ts">
	import { Button } from '@board-games/shared/ui/button/index.js';
	import { Input } from '@board-games/shared/ui/input/index.js';
	import * as Card from '#lib/components/ui/card/index.js';

	// The gate renders the built page as a Go html/template (backend/internal/gate/ui.go): the
	// actions below stay as text in the HTML and are filled in per request, escaped. In dev the page
	// is not a template, so it gets plain values instead.
	const dev = import.meta.env.DEV;
	const action = dev ? '#' : '{{.Action}}';
	const next = dev ? '/board-games/' : '{{.Next}}';
</script>

<svelte:head>
	<title>Board Games</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<div class="flex flex-1 items-center justify-center py-6">
	<Card.Root class="w-full max-w-sm">
		<Card.Header>
			<!-- The start page's meeple: board games in general. -->
			<svg viewBox="0 0 24 24" class="mb-2 size-8 text-accent-piece" aria-hidden="true">
				<path
					d="M12 2a3 3 0 0 1 3 3c0 .9-.4 1.7-1 2.2 2.6.5 6.5 1.6 6.5 3.8 0 1.2-1.6 1.7-3.5 1.9L20 21h-5.5L12 17l-2.5 4H4l2.9-8.1C5.1 12.7 3.5 12.2 3.5 11c0-2.2 3.9-3.3 6.5-3.8-.6-.5-1-1.3-1-2.2a3 3 0 0 1 3-3z"
					fill="currentColor"
				/>
			</svg>
			<Card.Title class="text-2xl font-bold tracking-tight">Board Games</Card.Title>
			<Card.Description>Enter the password to play.</Card.Description>
		</Card.Header>
		<Card.Content>
			<form method="post" {action} class="flex flex-col gap-2">
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
		</Card.Content>
	</Card.Root>
</div>
