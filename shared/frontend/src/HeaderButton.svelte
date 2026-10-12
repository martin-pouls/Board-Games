<!--
	A button or link in a page header: the shared top bar ("All games") and the games' own lobby and
	table headers (home, rules, end game, ...). Every header item has the same size (32px high, 16px
	icon, small muted text), so the headers look alike across screens and games. Rows of them use
	headerRow, which also pulls the first and last button out so their icons line up with the content.
-->
<script lang="ts" module>
	/** The classes of a header row: same height and spacing as the shared top bar. */
	export const headerRow = '-mx-2 -mt-1 flex min-h-8 items-center gap-1.5';
</script>

<script lang="ts">
	import type { Component } from 'svelte';
	import type { HTMLAnchorAttributes } from 'svelte/elements';
	import { Button } from './ui/button/index.js';
	import { cn } from './utils.js';

	let {
		icon: Icon,
		label,
		showLabel = false,
		labelClass,
		title = label,
		destructive = false,
		href,
		rel,
		disabled,
		onclick,
		class: className
	}: {
		icon: Component<{ class?: string }>;
		/** The visible text with showLabel, else the accessible name and tooltip. */
		label: string;
		/** Shows the label next to the icon. */
		showLabel?: boolean;
		/** Classes of the visible label, e.g. 'max-[24rem]:sr-only' to hide it on narrow phones. */
		labelClass?: string;
		/** The tooltip, when it should say more than the label. */
		title?: string;
		/** Turns red on hover: for actions that end or leave something. */
		destructive?: boolean;
		href?: string;
		rel?: HTMLAnchorAttributes['rel'];
		disabled?: boolean;
		onclick?: () => void;
		class?: string;
	} = $props();
</script>

<Button
	variant="ghost"
	size={showLabel ? 'sm' : 'icon-sm'}
	class={cn('text-muted-foreground', destructive && 'hover:text-destructive', className)}
	{href}
	{rel}
	{disabled}
	{onclick}
	aria-label={showLabel ? undefined : label}
	{title}
>
	<Icon class="size-4" />
	{#if showLabel}<span class={labelClass}>{label}</span>{/if}
</Button>
