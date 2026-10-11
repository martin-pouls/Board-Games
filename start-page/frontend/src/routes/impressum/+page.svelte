<script lang="ts">
	import { resolve } from '$app/paths';
	import { operator, siteName } from '#lib/site.js';
</script>

<svelte:head>
	<title>Impressum · {siteName}</title>
</svelte:head>

<a
	href={resolve('/')}
	class="-ml-1 self-start rounded px-1 text-sm text-muted-foreground hover:text-foreground"
>
	← {siteName}
</a>

<article lang="de" class="legal mt-4">
	<h1 class="text-3xl font-bold tracking-tight sm:text-4xl">Impressum</h1>

	<h2>Angaben gemäß § 5 DDG und § 18 Abs. 1 MStV</h2>
	<address>
		{operator.name}<br />
		{operator.street}<br />
		{operator.city}<br />
		Deutschland
	</address>

	<h2>Kontakt</h2>
	<p>E-Mail: {operator.email}</p>

	<p>
		Diese Website ist ein privates, nicht-kommerzielles Hobbyprojekt. Es gibt keine Werbung und
		nichts zu kaufen.
	</p>

	<h2>Haftung für Inhalte</h2>
	<p>
		Als Diensteanbieter ist der Betreiber nach § 7 Abs. 1 DDG für eigene Inhalte auf diesen Seiten
		nach den allgemeinen Gesetzen verantwortlich. Nach §§ 8 bis 10 DDG ist er jedoch nicht
		verpflichtet, übermittelte oder gespeicherte fremde Informationen zu überwachen oder nach
		Umständen zu forschen, die auf eine rechtswidrige Tätigkeit hinweisen. Fremde Inhalte sind hier
		vor allem die Spielernamen, die Spielende selbst wählen.
	</p>
	<p>
		Verpflichtungen zur Entfernung oder Sperrung der Nutzung von Informationen nach den allgemeinen
		Gesetzen bleiben hiervon unberührt. Eine Haftung ist erst ab dem Zeitpunkt der Kenntnis einer
		konkreten Rechtsverletzung möglich. Sobald der Betreiber von einer Rechtsverletzung erfährt,
		entfernt er den betreffenden Inhalt umgehend.
	</p>

	<h2>Haftung für Links</h2>
	<p>
		Sollte diese Website auf externe Websites Dritter verlinken, hat der Betreiber auf deren Inhalte
		keinen Einfluss und kann für diese fremden Inhalte keine Gewähr übernehmen. Für die Inhalte der
		verlinkten Seiten ist stets der jeweilige Anbieter oder Betreiber verantwortlich. Die verlinkten
		Seiten werden zum Zeitpunkt der Verlinkung auf mögliche Rechtsverstöße überprüft. Eine
		permanente inhaltliche Kontrolle ist ohne konkrete Anhaltspunkte einer Rechtsverletzung nicht
		zumutbar. Bei Bekanntwerden von Rechtsverletzungen werden derartige Links umgehend entfernt.
	</p>

	<h2>Urheberrecht</h2>
	<p>
		Die vom Betreiber erstellten Inhalte und Werke auf diesen Seiten unterliegen dem deutschen
		Urheberrecht. Vervielfältigung, Bearbeitung, Verbreitung und jede Art der Verwertung außerhalb
		der Grenzen des Urheberrechts bedürfen der schriftlichen Zustimmung des Betreibers. Soweit
		Inhalte nicht vom Betreiber erstellt wurden, werden die Urheberrechte Dritter beachtet. Sollten
		Sie trotzdem auf eine Urheberrechtsverletzung aufmerksam werden, wird um einen entsprechenden
		Hinweis per E-Mail gebeten. Bei Bekanntwerden von Rechtsverletzungen werden derartige Inhalte
		umgehend entfernt.
	</p>
</article>
