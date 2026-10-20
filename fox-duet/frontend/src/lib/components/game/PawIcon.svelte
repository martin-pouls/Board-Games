<script lang="ts">
	let { class: className = '' }: { class?: string } = $props();
</script>

<svg viewBox="0 0 24 24" class={className} fill="currentColor" aria-hidden="true">
	<ellipse cx="12" cy="16.2" rx="5.2" ry="4.4" />
	<circle cx="5.2" cy="10.4" r="2.3" />
	<circle cx="9.4" cy="6.2" r="2.4" />
	<circle cx="14.6" cy="6.2" r="2.4" />
	<circle cx="18.8" cy="10.4" r="2.3" />
</svg>
