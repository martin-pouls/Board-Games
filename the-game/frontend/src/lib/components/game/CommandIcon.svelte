<script lang="ts">
	import OctagonX from '@lucide/svelte/icons/octagon-x';
	import Skull from '@lucide/svelte/icons/skull';
	import MessageCircleOff from '@lucide/svelte/icons/message-circle-off';
	import Undo2 from '@lucide/svelte/icons/undo-2';
	import Slash from '@lucide/svelte/icons/slash';
	import Target from '@lucide/svelte/icons/target';
	import ArrowDownToLine from '@lucide/svelte/icons/arrow-down-to-line';
	import { cn } from '#lib/utils.js';
	import type { CommandName } from '#lib/types.js';

	/** Decorative glyph for a command (the surrounding element carries the text). */
	let { command, class: className = '' }: { command: CommandName; class?: string } = $props();

	const cls = $derived(cn('size-4 shrink-0', className));
</script>

{#if command === 'stop'}
	<OctagonX class={cls} aria-hidden="true" />
{:else if command === 'skull'}
	<Skull class={cls} aria-hidden="true" />
{:else if command === 'exactlyThree'}
	<!-- lucide has no glyph for 3!: text sized to the icon box -->
	<span
		class={cn('[container-type:size] inline-flex items-center justify-center', cls)}
		aria-hidden="true"
	>
		<span class="leading-none font-black tracking-tighter" style="font-size: 82cqh">3!</span>
	</span>
{:else if command === 'noCommunication'}
	<MessageCircleOff class={cls} aria-hidden="true" />
{:else if command === 'noBackwardsTrick'}
	<span class={cn('relative inline-block', cls)} aria-hidden="true">
		<Undo2 class="absolute inset-0 size-full" />
		<Slash class="absolute inset-0 size-full" />
	</span>
{:else if command === 'singlePile'}
	<Target class={cls} aria-hidden="true" />
{:else if command === 'drawOnlyOne'}
	<ArrowDownToLine class={cls} aria-hidden="true" />
{/if}
