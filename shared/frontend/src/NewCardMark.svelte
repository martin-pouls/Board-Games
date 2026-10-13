<!--
@component
Marks a card that just came into the hand: a dot that lands with the card and a soft glow that
pulses around it once. Place it inside a positioned element with the card's shape (it fills it
and inherits its corner radius); show it while the card is new, e.g. `{#if fresh.has(key)}`, and
it fades out when the mark ends. The colour is the app's `--new-card` token.

The glow is its own layer and only its opacity animates: a filter on the card itself would move
the card to a layer of its own while it runs and back afterwards, which shifts it by a fraction
of a pixel.
-->
<script lang="ts">
	import { fade } from 'svelte/transition';
	import { cn } from './utils.js';

	let {
		dot = 'inside',
		ring = 'var(--card)',
		class: className
	}: {
		/**
		 * Where the dot sits: inside the top right corner, or on the top left or top right corner
		 * (for cards that overlap, the corner that stays visible).
		 */
		dot?: 'inside' | 'corner-left' | 'corner-right';
		/** The colour of the thin ring around the dot, usually the card's face. */
		ring?: string;
		class?: string;
	} = $props();
</script>

<span
	class={cn('new-card pointer-events-none absolute inset-0 rounded-[inherit]', className)}
	style:--ring={ring}
	aria-hidden="true"
	out:fade={{ duration: 300 }}
>
	<span class="glow"></span>
	<span class="dot {dot}"></span>
</span>

<style>
	.glow {
		position: absolute;
		inset: 0;
		border-radius: inherit;
		box-shadow:
			0 0 2px 1px var(--new-card),
			0 0 10px 3px var(--new-card);
		opacity: 0;
		animation: glow 2.4s ease-in-out 0.15s;
	}
	.dot {
		position: absolute;
		width: 9px;
		height: 9px;
		border-radius: 9999px;
		background: var(--new-card);
		box-shadow: 0 0 0 2px var(--ring);
		animation: dot-in 500ms ease-out;
	}
	.inside {
		top: 5px;
		right: 5px;
	}
	.corner-right {
		top: -4px;
		right: -4px;
	}
	.corner-left {
		top: -4px;
		left: -4px;
	}
	@keyframes glow {
		0%,
		100% {
			opacity: 0;
		}
		25%,
		75% {
			opacity: 1;
		}
		50% {
			opacity: 0.35;
		}
	}
	@keyframes dot-in {
		0% {
			transform: scale(0.3);
		}
		60% {
			transform: scale(1.4);
		}
		100% {
			transform: scale(1);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.glow,
		.dot {
			animation: none;
		}
	}
</style>
