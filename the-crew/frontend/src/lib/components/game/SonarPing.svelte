<script lang="ts">
	// Deep Sea's counterpart of the radio icon (communication): a sonar scope with a ping. Drawn like
	// the lucide icons (24 grid, 2 px round stroke), so it takes their classes.
	let { class: className = '', 'aria-label': label }: { class?: string; 'aria-label'?: string } =
		$props();
</script>

<svg
	viewBox="0 0 24 24"
	class={className}
	fill="none"
	stroke="currentColor"
	stroke-width="2"
	stroke-linecap="round"
	stroke-linejoin="round"
	role={label ? 'img' : undefined}
	aria-label={label}
	aria-hidden={label ? undefined : true}
>
	<circle cx="12" cy="12" r="1.2" fill="currentColor" />
	<circle cx="12" cy="12" r="5" />
	<path d="M12 2a10 10 0 0 1 10 10" />
	<path d="M12 12 18.5 5.5" />
</svg>
