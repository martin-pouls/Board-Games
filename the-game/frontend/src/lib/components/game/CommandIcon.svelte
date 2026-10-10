<script lang="ts">
	import OctagonX from '@lucide/svelte/icons/octagon-x';
	import Skull from '@lucide/svelte/icons/skull';
	import MessageCircleOff from '@lucide/svelte/icons/message-circle-off';
	import Undo2 from '@lucide/svelte/icons/undo-2';
	import Slash from '@lucide/svelte/icons/slash';
	import { cn } from '#lib/utils.js';
	import type { CommandName } from '#lib/types.js';

	/** Decorative glyph for a command (the surrounding element carries the text). */
	let { command, class: className = '' }: { command: CommandName; class?: string } = $props();

	const cls = $derived(cn('size-4 shrink-0', className));
	/** lucide's stroke style, for the glyphs lucide doesn't have. */
	const svg = {
		xmlns: 'http://www.w3.org/2000/svg',
		viewBox: '0 0 24 24',
		fill: 'none',
		stroke: 'currentColor',
		'stroke-width': 2,
		'stroke-linecap': 'round',
		'stroke-linejoin': 'round',
		'aria-hidden': 'true'
	} as const;
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
	<!-- every card onto one stack -->
	<svg {...svg} class={cls}>
		<path d="M12 2.5v8" />
		<path d="m8.5 7 3.5 3.5L15.5 7" />
		<rect x="4" y="14" width="16" height="4" rx="1" />
		<path d="M6 21.5h12" />
	</svg>
{:else if command === 'drawOnlyOne'}
	<!-- a single card (marked 1) drawn into the hand -->
	<svg {...svg} class={cls}>
		<rect x="2.5" y="3" width="12" height="17" rx="2" />
		<path d="M7 9l2.5-2v9" />
		<path d="M19 6v13" />
		<path d="m16 16 3 3 3-3" />
	</svg>
{/if}
