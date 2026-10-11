<script lang="ts">
	import { resolve } from '$app/paths';
	import { operator as controller, siteName } from '#lib/site.js';
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
		Diese Website (yoshi-und-kruemli.de und martin-pouls.de) ist ein privates, nicht-kommerzielles
		Hobbyprojekt. Es gibt keine Konten, keine Werbung, kein Tracking und keine Dienste Dritter; alle
		Inhalte einschließlich der Schriftarten kommen vom eigenen Server.
	</p>

	<h2>Verantwortlicher</h2>
	<address>
		{controller.name}<br />
		{controller.street}<br />
		{controller.city}<br />
		E-Mail: {controller.email}
	</address>

	<h2>Hosting</h2>
	<p>
		Der Server steht bei der Hetzner Online GmbH (Industriestr. 25, 91710 Gunzenhausen) in einem
		Rechenzentrum in Nürnberg. Hetzner verarbeitet die Daten nur in unserem Auftrag (Vertrag nach
		Art. 28 DSGVO). Darüber hinaus werden keine Daten weitergegeben, auch nicht in Länder außerhalb
		der EU.
	</p>

	<h2>Aufruf der Website</h2>
	<p>
		Beim Aufruf überträgt Ihr Browser technisch notwendige Daten, etwa Ihre IP-Adresse. Zugriffe
		werden nicht protokolliert. Damit niemand massenhaft Spiele anlegt, Spielcodes oder das Passwort
		durchprobiert, zählt der Server Anfragen pro IP-Adresse, nur im Arbeitsspeicher und für wenige
		Minuten. Fehlermeldungen des Servers können eine IP-Adresse enthalten; sie werden nach
		spätestens 24 Stunden gelöscht. Rechtsgrundlage ist Art. 6 Abs. 1 lit. f DSGVO (sicherer Betrieb
		der Website).
	</p>

	<h2>Login</h2>
	<p>
		Die Seiten sind durch ein gemeinsames Passwort geschützt. Nach dem Login setzt der Server das
		Cookie <code>bg_gate</code>, das ein Jahr lang gilt. Es enthält nur die Passwortgruppe (zum
		Beispiel „Familie“), den Zeitpunkt des Logins und eine Signatur. Das eingegebene Passwort wird
		weder gespeichert noch protokolliert. Rechtsgrundlage ist § 25 Abs. 2 Nr. 2 TDDDG (unbedingt
		erforderlich, keine Einwilligung nötig) und Art. 6 Abs. 1 lit. f DSGVO (Zugang nur für den
		gewünschten Personenkreis).
	</p>

	<h2>Spielen</h2>
	<p>
		Für ein Spiel speichert der Server Ihren Spielernamen (ein Spitzname genügt), den Spielverlauf
		und einen zufälligen Zugangsschlüssel. Der Name ist für alle sichtbar, die den Spielcode kennen.
		Spiele werden sieben Tage nach der letzten Aktivität automatisch gelöscht, bei The Crew nach 30
		Tagen (damit das Logbuch einer Kampagne erhalten bleibt). Rechtsgrundlage ist Art. 6 Abs. 1 lit.
		b DSGVO (Bereitstellung des Spiels).
	</p>
	<p>
		Im lokalen Speicher Ihres Browsers liegen Ihre Spielsitzungen, der zuletzt verwendete Name und
		die Wahl zwischen hellem und dunklem Design. Diese Daten bleiben auf Ihrem Gerät, bis Sie sie
		löschen (§ 25 Abs. 2 Nr. 2 TDDDG).
	</p>

	<h2>Ihre Rechte</h2>
	<p>
		Sie haben das Recht auf Auskunft, Berichtigung, Löschung, Einschränkung der Verarbeitung,
		Datenübertragbarkeit und Widerspruch (Art. 15–18, 20, 21 DSGVO); eine formlose E-Mail genügt. Da
		es keine Konten gibt, brauchen wir dafür den Spielcode und Ihren Spielernamen. Außerdem können
		Sie sich bei einer Datenschutz-Aufsichtsbehörde beschweren (Art. 77 DSGVO), zum Beispiel beim
		Landesbeauftragten für den Datenschutz und die Informationsfreiheit Baden-Württemberg.
	</p>
</article>
