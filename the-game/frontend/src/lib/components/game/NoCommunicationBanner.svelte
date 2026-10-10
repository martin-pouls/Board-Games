<script lang="ts">
	import MessageCircleOff from '@lucide/svelte/icons/message-circle-off';
	import { slide } from 'svelte/transition';
	import { prefersReducedMotion } from 'svelte/motion';
</script>

<div
	transition:slide={{ duration: prefersReducedMotion.current ? 0 : 150 }}
	role="status"
	class="flex items-center gap-2 rounded-lg bg-violet-600 px-3 py-1.5 text-sm leading-snug font-medium text-white"
>
	<MessageCircleOff class="size-4 shrink-0" aria-hidden="true" />
	<span>No communication: don't talk about the game until the card is covered.</span>
</div>
