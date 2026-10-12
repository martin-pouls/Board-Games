<script lang="ts">
	import { operator as controller, siteName } from '#lib/site.js';
</script>

<svelte:head>
	<title>Datenschutz · {siteName}</title>
</svelte:head>

<article lang="de" class="legal mt-2">
	<h1 class="text-3xl font-bold tracking-tight sm:text-4xl">Datenschutzerklärung</h1>
	<p class="text-sm text-muted-foreground">Stand: Oktober 2026</p>

	<h2>1. Datenschutz auf einen Blick</h2>
	<h3>Allgemeine Hinweise</h3>
	<p>
		Die folgenden Hinweise geben einen Überblick darüber, was mit Ihren personenbezogenen Daten
		passiert, wenn Sie diese Website besuchen. Personenbezogene Daten sind alle Daten, mit denen Sie
		persönlich identifiziert werden können. Diese Erklärung gilt für alle Seiten unter
		yoshi-und-kruemli.de und martin-pouls.de (derselbe Server, derselbe Betreiber): die Übersicht
		unter /board-games/ und die dort verlinkten Spiele.
	</p>

	<h3>Wer ist verantwortlich für die Datenerfassung auf dieser Website?</h3>
	<p>
		Die Datenverarbeitung auf dieser Website erfolgt durch den Websitebetreiber. Dessen Kontaktdaten
		finden Sie im Abschnitt „Hinweis zur verantwortlichen Stelle“ in dieser Datenschutzerklärung.
	</p>

	<h3>Wie erfassen wir Ihre Daten?</h3>
	<p>
		Einige Daten geben Sie selbst ein, zum Beispiel Ihren Spielernamen. Andere Daten werden beim
		Besuch der Website automatisch durch unsere IT-Systeme erfasst. Das sind vor allem technische
		Daten (z. B. IP-Adresse, Browser, Uhrzeit des Seitenaufrufs).
	</p>

	<h3>Wofür nutzen wir Ihre Daten?</h3>
	<p>
		Die Daten werden erhoben, um die Website und die Spiele fehlerfrei bereitzustellen und vor
		Missbrauch zu schützen. Diese Website ist ein privates, nicht-kommerzielles Hobbyprojekt. Es
		gibt keine Werbung, keine Analyse-Tools, kein Tracking und keine eingebundenen Dienste Dritter.
	</p>

	<h3>Welche Rechte haben Sie bezüglich Ihrer Daten?</h3>
	<p>
		Sie haben jederzeit das Recht, unentgeltlich Auskunft über Herkunft, Empfänger und Zweck Ihrer
		gespeicherten personenbezogenen Daten zu erhalten. Sie haben außerdem ein Recht, die
		Berichtigung oder Löschung dieser Daten zu verlangen. Hierzu sowie zu weiteren Fragen zum
		Datenschutz können Sie sich jederzeit an uns wenden. Des Weiteren steht Ihnen ein
		Beschwerderecht bei der zuständigen Aufsichtsbehörde zu. Details finden Sie in Abschnitt 3.
	</p>

	<h2>2. Hosting</h2>
	<p>
		Wir hosten diese Website bei der Hetzner Online GmbH, Industriestr. 25, 91710 Gunzenhausen,
		Deutschland. Der Server steht in einem Rechenzentrum in Nürnberg. Alle personenbezogenen Daten,
		die auf dieser Website erfasst werden, werden auf diesem Server gespeichert.
	</p>
	<p>
		Die Verwendung von Hetzner erfolgt auf Grundlage von Art. 6 Abs. 1 lit. f DSGVO. Wir haben ein
		berechtigtes Interesse an einer zuverlässigen Bereitstellung unserer Website.
	</p>
	<h3>Auftragsverarbeitung</h3>
	<p>
		Wir haben mit Hetzner einen Vertrag über Auftragsverarbeitung (AVV) nach Art. 28 DSGVO
		geschlossen. Hetzner verarbeitet die Daten unserer Websitebesucher nur nach unseren Weisungen
		und unter Einhaltung der DSGVO.
	</p>

	<h2>3. Allgemeine Hinweise und Pflichtinformationen</h2>
	<h3>Datenschutz</h3>
	<p>
		Wir behandeln Ihre personenbezogenen Daten vertraulich und entsprechend den gesetzlichen
		Datenschutzvorschriften sowie dieser Datenschutzerklärung. Wir weisen darauf hin, dass die
		Datenübertragung im Internet (z. B. bei der Kommunikation per E-Mail) Sicherheitslücken
		aufweisen kann. Ein lückenloser Schutz der Daten vor dem Zugriff durch Dritte ist nicht möglich.
	</p>

	<h3>Hinweis zur verantwortlichen Stelle</h3>
	<p>Verantwortlich für die Datenverarbeitung auf dieser Website ist:</p>
	<address>
		{controller.name}<br />
		{controller.street}<br />
		{controller.city}<br />
		E-Mail: {controller.email}
	</address>
	<p>
		Verantwortliche Stelle ist die natürliche oder juristische Person, die allein oder gemeinsam mit
		anderen über die Zwecke und Mittel der Verarbeitung personenbezogener Daten entscheidet.
	</p>

	<h3>Speicherdauer</h3>
	<p>
		Soweit innerhalb dieser Datenschutzerklärung keine speziellere Speicherdauer genannt wurde,
		verbleiben Ihre personenbezogenen Daten bei uns, bis der Zweck für die Datenverarbeitung
		entfällt. Wenn Sie ein berechtigtes Löschersuchen geltend machen, werden Ihre Daten gelöscht,
		sofern wir keine anderen rechtlich zulässigen Gründe für die Speicherung haben.
	</p>

	<h3>Rechtsgrundlagen der Datenverarbeitung</h3>
	<p>
		Wir verarbeiten Ihre Daten, soweit sie zur Bereitstellung der von Ihnen genutzten Spiele
		erforderlich sind, auf Grundlage von Art. 6 Abs. 1 lit. b DSGVO und im Übrigen auf Grundlage
		unseres berechtigten Interesses nach Art. 6 Abs. 1 lit. f DSGVO. Das Speichern von Informationen
		in Ihrem Endgerät (Cookie, lokaler Speicher) erfolgt nur, soweit es unbedingt erforderlich ist,
		auf Grundlage von § 25 Abs. 2 Nr. 2 TDDDG. Eine Einwilligung holen wir nicht ein, weil keine
		Verarbeitung sie erfordert. Die im Einzelfall einschlägige Rechtsgrundlage ist in den folgenden
		Abschnitten angegeben.
	</p>

	<h3>Empfänger von personenbezogenen Daten</h3>
	<p>
		Daten, die beim Besuch der Website anfallen, geben wir an niemanden weiter außer an unseren
		Hoster (Abschnitt 2); sie verlassen die Europäische Union nicht. Zu E-Mails siehe „Anfrage per
		E-Mail“ in Abschnitt 4.
	</p>

	<h3>Widerspruchsrecht gegen die Datenerhebung in besonderen Fällen (Art. 21 DSGVO)</h3>
	<p>
		Wenn die Datenverarbeitung auf Grundlage von Art. 6 Abs. 1 lit. f DSGVO erfolgt, haben Sie
		jederzeit das Recht, aus Gründen, die sich aus Ihrer besonderen Situation ergeben, gegen die
		Verarbeitung Ihrer personenbezogenen Daten Widerspruch einzulegen. Wenn Sie Widerspruch
		einlegen, werden wir Ihre betroffenen personenbezogenen Daten nicht mehr verarbeiten, es sei
		denn, wir können zwingende schutzwürdige Gründe für die Verarbeitung nachweisen, die Ihre
		Interessen, Rechte und Freiheiten überwiegen, oder die Verarbeitung dient der Geltendmachung,
		Ausübung oder Verteidigung von Rechtsansprüchen.
	</p>

	<h3>Beschwerderecht bei der zuständigen Aufsichtsbehörde</h3>
	<p>
		Im Falle von Verstößen gegen die DSGVO steht Ihnen ein Beschwerderecht bei einer
		Aufsichtsbehörde zu, insbesondere in dem Mitgliedstaat Ihres gewöhnlichen Aufenthalts, Ihres
		Arbeitsplatzes oder des Orts des mutmaßlichen Verstoßes (Art. 77 DSGVO). Für uns zuständig ist
		der Landesbeauftragte für den Datenschutz und die Informationsfreiheit Baden-Württemberg,
		Lautenschlagerstraße 20, 70173 Stuttgart.
	</p>

	<h3>Recht auf Datenübertragbarkeit</h3>
	<p>
		Sie haben das Recht, Daten, die wir auf Grundlage eines Vertrags automatisiert verarbeiten, an
		sich oder an einen Dritten in einem gängigen, maschinenlesbaren Format aushändigen zu lassen
		(Art. 20 DSGVO).
	</p>

	<h3>Auskunft, Berichtigung, Löschung und Einschränkung der Verarbeitung</h3>
	<p>
		Sie haben im Rahmen der geltenden gesetzlichen Bestimmungen jederzeit das Recht auf
		unentgeltliche Auskunft über Ihre gespeicherten personenbezogenen Daten, deren Herkunft und
		Empfänger und den Zweck der Datenverarbeitung (Art. 15 DSGVO) sowie ein Recht auf Berichtigung
		(Art. 16 DSGVO), Löschung (Art. 17 DSGVO) und Einschränkung der Verarbeitung (Art. 18 DSGVO).
		Hierzu genügt eine formlose E-Mail an die oben genannte Adresse. Da es auf dieser Website keine
		Benutzerkonten gibt, können wir Daten Ihnen nur zuordnen, wenn Sie uns den Spielcode und Ihren
		Spielernamen nennen.
	</p>

	<h3>SSL- bzw. TLS-Verschlüsselung</h3>
	<p>
		Diese Website nutzt aus Sicherheitsgründen eine SSL- bzw. TLS-Verschlüsselung. Eine
		verschlüsselte Verbindung erkennen Sie daran, dass die Adresszeile des Browsers mit „https://“
		beginnt und an dem Schloss-Symbol in Ihrer Browserzeile.
	</p>

	<h3>Pflicht zur Bereitstellung, automatisierte Entscheidungen</h3>
	<p>
		Sie sind nicht verpflichtet, Daten bereitzustellen. Ohne die technisch notwendigen
		Verbindungsdaten kann die Website jedoch nicht angezeigt werden, und ohne Spielernamen ist kein
		Mitspielen möglich. Eine automatisierte Entscheidungsfindung einschließlich Profiling findet
		nicht statt.
	</p>

	<h2>4. Datenerfassung auf dieser Website</h2>
	<h3>Server-Log-Dateien</h3>
	<p>
		Beim Aufruf jeder Seite überträgt Ihr Browser technisch notwendige Daten an den Server: Ihre
		IP-Adresse, Datum und Uhrzeit, die aufgerufene Adresse sowie Angaben zu Browser und
		Betriebssystem. Diese Daten werden nur verarbeitet, um die Seite auszuliefern.
	</p>
	<p>
		Zugriffe werden nicht in Log-Dateien gespeichert. Nur wenn auf dem Server ein technischer Fehler
		auftritt, kann die Fehlermeldung eine IP-Adresse enthalten. Fehlermeldungen dienen
		ausschließlich der Fehlerbehebung und werden nach spätestens 24 Stunden automatisch gelöscht.
	</p>
	<p>
		Rechtsgrundlage ist Art. 6 Abs. 1 lit. f DSGVO. Wir haben ein berechtigtes Interesse an der
		technisch fehlerfreien Darstellung und dem sicheren Betrieb unserer Website.
	</p>

	<h3>Schutz vor Missbrauch</h3>
	<p>
		Um die Website vor automatisierten Anfragen zu schützen, etwa dem massenhaften Anlegen von
		Spielen oder dem Durchprobieren von Spielcodes oder Passwörtern, zählt der Server, wie viele
		Anfragen von einer IP-Adresse kommen. Die Adresse wird dafür nur im Arbeitsspeicher gehalten,
		nicht dauerhaft gespeichert und in der Regel nach wenigen Minuten verworfen. Wird eine Grenze
		überschritten, lehnt der Server weitere Anfragen kurzzeitig ab.
	</p>
	<p>
		Rechtsgrundlage ist Art. 6 Abs. 1 lit. f DSGVO. Wir haben ein berechtigtes Interesse daran,
		Missbrauch zu verhindern und die Spiele für alle verfügbar zu halten.
	</p>

	<h3>Passwortschutz und Cookie</h3>
	<p>
		Die Übersicht und die Spiele sind durch ein gemeinsames Passwort geschützt. Der Server
		vergleicht das eingegebene Passwort mit den hinterlegten Passwörtern; das eingegebene Passwort
		wird weder gespeichert noch protokolliert. Im Server-Protokoll stehen nur ein erfolgreicher oder
		fehlgeschlagener Login, der Zeitpunkt und die Bezeichnung der Passwortgruppe.
	</p>
	<p>
		Nach erfolgreichem Login setzt der Server ein Cookie namens <code>bg_gate</code>, damit Sie das
		Passwort nicht bei jedem Aufruf neu eingeben müssen. Es enthält die Bezeichnung der
		Passwortgruppe, den Zeitpunkt des Logins und eine Signatur, die Fälschungen verhindert. Es dient
		nicht dem Tracking. Das Cookie gilt ein Jahr, ist für Skripte auf der Seite nicht lesbar
		(HttpOnly) und wird nur an diese Website gesendet. Wenn Sie es in Ihrem Browser löschen, sind
		Sie abgemeldet. Weitere Cookies verwenden wir nicht.
	</p>
	<p>
		Rechtsgrundlage für das Cookie ist § 25 Abs. 2 Nr. 2 TDDDG (unbedingt erforderlich, um den von
		Ihnen gewünschten Dienst bereitzustellen), für die damit verbundene Verarbeitung Art. 6 Abs. 1
		lit. f DSGVO. Wir haben ein berechtigtes Interesse daran, die Seiten nur dem gewünschten
		Personenkreis zugänglich zu machen.
	</p>

	<h3>Spiele</h3>
	<p>
		Wenn Sie ein Spiel erstellen oder einem Spiel beitreten, speichert der Server den von Ihnen
		gewählten Spielernamen, den Spielverlauf und den Spielstand sowie die Zeitpunkte der Erstellung
		und der letzten Aktivität. Ein zufälliger Zugangsschlüssel ordnet Ihren Browser Ihrem Platz im
		Spiel zu. Ihr Spielername ist für alle sichtbar, die den Spielcode kennen. Ein echter Name ist
		nicht nötig, ein Spitzname genügt.
	</p>
	<p>
		Spiele werden nach der letzten Aktivität automatisch mit allen zugehörigen Daten gelöscht: in
		der Regel nach sieben Tagen, bei Spielen mit fortlaufender Kampagne nach 30 Tagen, damit der
		Fortschritt erhalten bleibt. Rechtsgrundlage ist Art. 6 Abs. 1 lit. b DSGVO: Die Daten sind
		nötig, um Ihnen das gewünschte Spiel bereitzustellen.
	</p>

	<h3>Lokaler Speicher im Browser</h3>
	<p>Die Website nutzt den lokalen Speicher Ihres Browsers (Local Storage), und zwar nur für:</p>
	<ul>
		<li>
			Ihre Spielsitzungen (Spielcode, Spielername, Zugangsschlüssel), damit Sie nach dem Neuladen
			weiterspielen können und Ihre letzten Spiele angezeigt werden,
		</li>
		<li>den zuletzt verwendeten Spielernamen,</li>
		<li>Ihre Wahl zwischen hellem und dunklem Design.</li>
	</ul>
	<p>
		Diese Daten bleiben auf Ihrem Gerät, bis Sie sie in Ihren Browsereinstellungen löschen. An den
		Server geht davon nur, was zum Spielen nötig ist (Spielcode und Zugangsschlüssel).
		Rechtsgrundlage ist § 25 Abs. 2 Nr. 2 TDDDG.
	</p>

	<h3>Schriftarten</h3>
	<p>
		Alle Schriftarten werden vom eigenen Server ausgeliefert. Es besteht keine Verbindung zu Servern
		Dritter, etwa von Google.
	</p>

	<h3>Anfrage per E-Mail</h3>
	<p>
		Wenn Sie uns per E-Mail kontaktieren, verarbeiten wir Ihre Anfrage einschließlich der daraus
		hervorgehenden personenbezogenen Daten (Name, Anfrage) zum Zweck der Bearbeitung Ihres
		Anliegens. Diese Daten geben wir nicht ohne Ihre Einwilligung weiter. Rechtsgrundlage ist Art. 6
		Abs. 1 lit. f DSGVO; unser berechtigtes Interesse liegt in der Beantwortung Ihrer Anfrage. Die
		Daten verbleiben bei uns, bis Ihr Anliegen erledigt ist und der Zweck der Speicherung entfällt.
	</p>
	<p>
		Unser E-Mail-Postfach wird von Google (Google Ireland Limited, Gordon House, Barrow Street,
		Dublin 4, Irland) bereitgestellt. Dabei kann eine Übermittlung in die USA nicht ausgeschlossen
		werden; Google ist nach dem EU-US Data Privacy Framework zertifiziert, auf dessen Grundlage
		(Angemessenheitsbeschluss nach Art. 45 DSGVO) die Übermittlung erfolgt.
	</p>
</article>
