<script lang="ts">
	import WifiOff from '@lucide/svelte/icons/wifi-off';
	import { slide } from 'svelte/transition';
</script>

<div
	transition:slide={{ duration: 150 }}
	role="status"
	class="fixed inset-x-0 top-0 z-50 flex items-center justify-center gap-2 bg-amber-500 px-3 py-1 pt-[max(0.25rem,env(safe-area-inset-top))] text-sm font-medium text-amber-950"
>
	<WifiOff class="size-4" />
	Reconnecting…
</div>
