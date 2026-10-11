<script lang="ts">
	import { resolve } from '$app/paths';
	import { operator as controller, siteName } from '#lib/site.js';

	const authority =
		'Landesbeauftragter für den Datenschutz und die Informationsfreiheit Baden-Württemberg, Lautenschlagerstraße 20, 70173 Stuttgart';
</script>

<svelte:head>
	<title>Datenschutz · {siteName}</title>
</svelte:head>

<a
	href={resolve('/')}
	class="-ml-1 self-start rounded px-1 text-sm text-muted-foreground hover:text-foreground"
>
	← {siteName}
</a>

<article lang="de" class="legal mt-4">
	<h1 class="text-3xl font-bold tracking-tight sm:text-4xl">Datenschutzerklärung</h1>
	<p class="text-sm text-muted-foreground">Stand: Oktober 2026</p>

	<p>
		Diese Website ist ein privates, nicht-kommerzielles Hobbyprojekt. Die Spiele lassen sich ohne
		Konto spielen. Es gibt keine Cookies, keine Werbung, kein Tracking und keine Analyse-Tools. Hier
		steht, welche Daten trotzdem anfallen. Die Erklärung gilt für alle Seiten unter
		yoshi-und-kruemli.de: die Übersicht unter /board-games/ und die Spiele „The Game“ (/the-game/),
		„Regicide“ (/regicide/) und „Fox Duet“ (/fox-duet/).
	</p>

	<h2>1. Verantwortlicher</h2>
	<p>Verantwortlich im Sinne der Datenschutz-Grundverordnung (DSGVO) ist:</p>
	<address>
		{controller.name}<br />
		{controller.street}<br />
		{controller.city}<br />
		E-Mail: {controller.email}
	</address>
	<p>Für Fragen zum Datenschutz und zur Ausübung Ihrer Rechte genügt eine formlose E-Mail.</p>

	<h2>2. Hosting</h2>
	<p>
		Die Website läuft auf einem Server der Hetzner Online GmbH, Industriestr. 25, 91710
		Gunzenhausen, in einem Rechenzentrum in Nürnberg. Hetzner verarbeitet die Daten nur im Auftrag
		des Verantwortlichen; mit Hetzner besteht ein Vertrag zur Auftragsverarbeitung nach Art. 28
		DSGVO.
	</p>

	<h2>3. Aufruf der Website</h2>
	<p>
		Beim Aufruf jeder Seite überträgt Ihr Browser technisch notwendige Daten an den Server: Ihre
		IP-Adresse, Datum und Uhrzeit, die aufgerufene Adresse sowie Angaben zu Browser und
		Betriebssystem. Sie werden nur verarbeitet, um die Seite auszuliefern und den Betrieb sicher zu
		halten.
	</p>
	<p>
		Zugriffe werden nicht protokolliert: Es gibt keine Zugriffs-Logdateien. Nur wenn auf dem Server
		ein technischer Fehler auftritt, kann die Fehlermeldung die IP-Adresse enthalten.
		Fehlermeldungen werden nur zur Fehlerbehebung verwendet und automatisch überschrieben, sobald
		das Protokoll eine feste Größe erreicht.
	</p>
	<p>
		Rechtsgrundlage ist Art. 6 Abs. 1 lit. f DSGVO. Das berechtigte Interesse besteht darin, die
		Website auszuliefern und sicher zu betreiben.
	</p>

	<h2>4. Schutz vor Missbrauch</h2>
	<p>
		Um die Spiele vor automatisierten Anfragen zu schützen, etwa dem massenhaften Anlegen von
		Spielen oder dem Durchprobieren von Spielcodes, zählt der Server, wie viele Anfragen von einer
		IP-Adresse kommen (bei IPv6 von ihrem /64-Adressbereich). Die Adresse wird dafür nur im
		Arbeitsspeicher gehalten, nicht gespeichert und verworfen, sobald sie dafür nicht mehr gebraucht
		wird, in der Regel nach wenigen Minuten. Wird eine Grenze überschritten, lehnt der Server
		weitere Anfragen kurzzeitig ab, ohne die IP-Adresse zu protokollieren.
	</p>
	<p>
		Rechtsgrundlage ist Art. 6 Abs. 1 lit. f DSGVO. Das berechtigte Interesse besteht darin,
		Missbrauch zu verhindern und die Spiele für alle verfügbar zu halten.
	</p>

	<h2>5. Spielen</h2>
	<p>
		Wenn Sie ein Spiel erstellen oder einem Spiel beitreten, speichert der Server den von Ihnen
		gewählten Spielernamen, den Spielverlauf und den Spielstand sowie die Zeitpunkte der Erstellung
		und der letzten Aktivität. Ein zufälliger Zugangsschlüssel ordnet Ihren Browser Ihrem Platz im
		Spiel zu. Ihr Spielername ist für Ihre Mitspieler sichtbar und für alle, die den Spielcode
		kennen. Ein echter Name ist nicht nötig, ein Spitzname genügt.
	</p>
	<p>
		Spiele, in denen sieben Tage lang nichts passiert ist, werden automatisch mit allen zugehörigen
		Daten gelöscht.
	</p>
	<p>
		Rechtsgrundlage ist Art. 6 Abs. 1 lit. b DSGVO: Die Daten sind nötig, um Ihnen das gewünschte
		Spiel bereitzustellen.
	</p>

	<h2>6. Speicherung in Ihrem Browser</h2>
	<p>
		Die Website setzt keine Cookies. Sie nutzt den lokalen Speicher Ihres Browsers (Local Storage)
		und zwar nur für:
	</p>
	<ul>
		<li>
			Ihre Spielsitzungen (Spielcode, Spielername, Zugangsschlüssel), damit Sie nach dem Neuladen
			weiterspielen können und Ihre letzten Spiele angezeigt werden,
		</li>
		<li>den zuletzt verwendeten Spielernamen, damit Sie ihn nicht erneut eingeben müssen,</li>
		<li>Ihre Wahl zwischen hellem und dunklem Design.</li>
	</ul>
	<p>
		Diese Daten bleiben auf Ihrem Gerät. An den Server geht davon nur, was zum Spielen nötig ist
		(Spielcode und Zugangsschlüssel). Sie können die Daten jederzeit in Ihren Browsereinstellungen
		löschen. Rechtsgrundlage ist § 25 Abs. 2 Nr. 2 TDDDG: Die Speicherung ist unbedingt
		erforderlich, um die von Ihnen gewünschten Funktionen bereitzustellen.
	</p>

	<h2>7. Keine Weitergabe, keine externen Dienste</h2>
	<p>
		Daten werden an niemanden weitergegeben außer an den Hoster (Abschnitt 2). Alle Inhalte,
		einschließlich der Schriftarten, kommen vom eigenen Server; Dienste Dritter sind nicht
		eingebunden. Eine Übermittlung in Länder außerhalb der Europäischen Union findet nicht statt.
		Die Verbindung ist per HTTPS verschlüsselt.
	</p>

	<h2>8. Ihre Rechte</h2>
	<p>Sie haben das Recht auf</p>
	<ul>
		<li>Auskunft über die zu Ihnen gespeicherten Daten (Art. 15 DSGVO),</li>
		<li>Berichtigung unrichtiger Daten (Art. 16 DSGVO),</li>
		<li>Löschung (Art. 17 DSGVO),</li>
		<li>Einschränkung der Verarbeitung (Art. 18 DSGVO),</li>
		<li>Datenübertragbarkeit (Art. 20 DSGVO),</li>
		<li>
			Widerspruch gegen Verarbeitungen nach Art. 6 Abs. 1 lit. f DSGVO aus Gründen, die sich aus
			Ihrer besonderen Situation ergeben (Art. 21 DSGVO).
		</li>
	</ul>
	<p>
		Da es keine Konten gibt, lassen sich Daten Ihnen nur zuordnen, wenn Sie den Spielcode und Ihren
		Spielernamen nennen.
	</p>
	<p>
		Sie haben außerdem das Recht, sich bei einer Datenschutz-Aufsichtsbehörde zu beschweren (Art. 77
		DSGVO), etwa bei der Behörde Ihres Wohnorts oder bei der für den Verantwortlichen zuständigen
		Behörde: {authority}.
	</p>

	<h2>9. Pflicht zur Bereitstellung, automatisierte Entscheidungen</h2>
	<p>
		Sie sind nicht verpflichtet, Daten bereitzustellen. Ohne die technisch notwendigen
		Verbindungsdaten kann die Website jedoch nicht angezeigt werden, und ohne Spielernamen ist kein
		Mitspielen möglich. Eine automatisierte Entscheidungsfindung einschließlich Profiling findet
		nicht statt.
	</p>
</article>
