<!--
	The Jester's three-pointed cap with bells. On the card it is two-tone (side points and bells in
	--j1, middle point and band in --j2, set by the card); as a UI icon it takes currentColor, with
	the middle a little lighter so the shape still reads.
-->
<script lang="ts">
	import { cn } from '#lib/utils.js';

	let { class: className, label }: { class?: string; label?: string } = $props();
</script>

<svg
	class={cn('jester-hat', className)}
	viewBox="0 0 24 24"
	role={label ? 'img' : undefined}
	aria-label={label}
	aria-hidden={label ? undefined : 'true'}
>
	<path class="side" d="M5 16.5 C5.5 10 3.5 6.5 0.8 8.6 C3.6 8.4 6.4 11.4 9 16.5 Z" />
	<path class="side" d="M19 16.5 C18.5 10 20.5 6.5 23.2 8.6 C20.4 8.4 17.6 11.4 15 16.5 Z" />
	<path class="mid" d="M8 16.5 C8.6 9 10.6 4.4 14.6 2.4 C12.6 5.6 13.8 11 16 16.5 Z" />
	<circle class="side" cx="1.4" cy="11" r="1.7" />
	<circle class="mid" cx="15.6" cy="4.6" r="1.7" />
	<circle class="side" cx="22.6" cy="11" r="1.7" />
	<path class="mid" d="M4.5 16.2 Q12 14.6 19.5 16.2 L19.5 20 Q12 18.4 4.5 20 Z" />
</svg>

<style>
	.jester-hat {
		flex: none;
	}
	.side {
		fill: var(--j1, currentColor);
	}
	.mid {
		fill: var(--j2, currentColor);
		opacity: var(--j2-opacity, 0.72);
	}
</style>
