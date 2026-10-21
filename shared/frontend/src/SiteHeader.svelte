<!--
	The site's heading, the same on the start page and the gate's login page, so it stays in place
	when the login leads on to the start page: the meeple and the site name on one line, the tagline
	below.
-->
<script lang="ts">
	import { siteName } from './site.js';
</script>

<header class="mb-6 text-center md:mt-2 md:mb-10">
	<div class="flex items-center justify-center gap-2.5">
		<!-- A meeple: board and card games in general. -->
		<svg
			viewBox="0 0 24 24"
			class="text-accent-piece size-7.5 shrink-0 md:size-10"
			aria-hidden="true"
		>
			<path
				d="M12 2a3 3 0 0 1 3 3c0 .9-.4 1.7-1 2.2 2.6.5 6.5 1.6 6.5 3.8 0 1.2-1.6 1.7-3.5 1.9L20 21h-5.5L12 17l-2.5 4H4l2.9-8.1C5.1 12.7 3.5 12.2 3.5 11c0-2.2 3.9-3.3 6.5-3.8-.6-.5-1-1.3-1-2.2a3 3 0 0 1 3-3z"
				fill="currentColor"
			/>
		</svg>
		<h1 class="text-3xl font-extrabold tracking-tight md:text-5xl">{siteName}</h1>
	</div>
	<p class="mx-auto mt-2 max-w-[34ch] text-balance text-muted-foreground md:max-w-none">
		Cooperative board and card games to play live with friends, right in the browser.
	</p>
</header>
