<script lang="ts">
	import { fade } from 'svelte/transition';
	import { prefersReducedMotion } from 'svelte/motion';
</script>

<!--
	While a No communication card is visible: a violet glow around the whole screen. It takes no
	room from the board yet stays in view; the solid chip on its pile names the rule.
-->
<div
	transition:fade={{ duration: prefersReducedMotion.current ? 0 : 200 }}
	class="pointer-events-none fixed inset-0 z-40 shadow-[inset_0_0_0_3px_var(--color-violet-500),inset_0_0_28px_--alpha(var(--color-violet-500)/45%)]"
	aria-hidden="true"
></div>
<p class="sr-only" role="status">
	No communication: don't talk about the game until the card is covered.
</p>
