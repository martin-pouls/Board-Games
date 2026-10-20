<script lang="ts">
	import type { Ability } from '#lib/types.js';

	let {
		ability,
		hole = 'var(--parchment)',
		class: className = ''
	}: {
		ability: Exclude<Ability, ''>;
		/** The colour of the cut-outs (eyes, ribbon), usually the surface behind the icon. */
		hole?: string;
		class?: string;
	} = $props();
</script>

<!-- The ability pictures of the odd ranks: Musician, Foxes, Gazelle, Gift, Royal Heir. -->
<svg viewBox="0 0 24 24" class={className} fill="currentColor" aria-hidden="true">
	{#if ability === 'musician'}
		<ellipse cx="7" cy="18" rx="3.3" ry="2.6" />
		<ellipse cx="17" cy="16" rx="3.3" ry="2.6" />
		<rect x="9.3" y="5.5" width="1.7" height="12.5" />
		<rect x="19.3" y="3.5" width="1.7" height="12.5" />
		<path d="M9.3 5.5 21 3.2v3L9.3 8.5Z" />
	{:else if ability === 'foxes'}
		<path d="M2 3 8.6 8.6h6.8L22 3l-1 9.2L12 21.5 3 12.2Z" />
		<circle cx="8.8" cy="12.6" r="1.2" fill={hole} />
		<circle cx="15.2" cy="12.6" r="1.2" fill={hole} />
	{:else if ability === 'gazelle'}
		<path
			d="M8.6 10.5C6 7 5.3 4 6.8 1.5M15.4 10.5C18 7 18.7 4 17.2 1.5"
			fill="none"
			stroke="currentColor"
			stroke-width="1.9"
			stroke-linecap="round"
		/>
		<path d="M8 10h8l-1.6 8.2L12 22l-2.4-3.8Z" />
		<path d="M8.2 10.6 3.4 9.2l3.8 4.2ZM15.8 10.6l4.8-1.4-3.8 4.2Z" />
	{:else if ability === 'gift'}
		<rect x="3.5" y="10.5" width="17" height="10.5" rx="1" />
		<rect x="2.5" y="7" width="19" height="4" rx="1" />
		<path d="M12 7C9.5 2.5 5 3.6 7.8 7ZM12 7c2.5-4.5 7-3.4 4.2 0Z" />
		<rect x="11" y="7" width="2" height="14" fill={hole} />
	{:else}
		<path d="M3 18 2 7l5.6 4.2L12 4l4.4 7.2L22 7l-1 11Z" />
		<rect x="3" y="19.2" width="18" height="2.6" rx="1" />
	{/if}
</svg>
