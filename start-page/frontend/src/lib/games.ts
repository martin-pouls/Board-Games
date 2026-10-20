/** A game listed on the start page. Each game is its own app, served by Caddy next to this one. */
export interface Game {
	id: 'the-game' | 'regicide' | 'fox-duet' | 'the-crew' | 'deep-sea';
	name: string;
	/** Absolute path of the game's app on the same host in production (outside this app's base path). */
	path: `/${string}/`;
	/** Port of the game's Vite dev server (`server.port` in its vite.config.ts), linked to in dev. */
	devPort: number;
	players: string;
	/** One line for the phone's shelf tile. */
	tagline: string;
	/** The longer text for the desktop tile. */
	description: string;
}

export const games: Game[] = [
	{
		id: 'the-game',
		name: 'The Game',
		path: '/the-game/',
		devPort: 5173,
		players: '1–5 players',
		tagline: 'Play all 98 cards onto four piles without naming a number.',
		description:
			'Play all 98 cards onto two rising and two falling piles, without saying which numbers you hold.'
	},
	{
		id: 'regicide',
		name: 'Regicide',
		path: '/regicide/',
		devPort: 5174,
		players: '1–4 players',
		tagline: 'Defeat twelve royals with a standard deck.',
		description:
			'Defeat twelve corrupted royals with a standard deck, using the powers of the four suits.'
	},
	{
		id: 'fox-duet',
		name: 'Fox Duet',
		path: '/fox-duet/',
		devPort: 5176,
		players: '2 players',
		tagline: 'Trick-taking for two along the forest path.',
		description:
			'The Fox in the Forest: Duet. Cooperative trick-taking: work together to win the right tricks.'
	},
	{
		id: 'the-crew',
		name: 'The Crew',
		path: '/the-crew/',
		devPort: 5177,
		players: '2–5 players',
		tagline: 'Trick-taking through 50 missions, without talking.',
		description: 'Cooperative trick-taking through 50 missions. No talking about your cards.'
	},
	{
		id: 'deep-sea',
		name: 'The Crew: Mission Deep Sea',
		path: '/deep-sea/',
		devPort: 5179,
		players: '2–5 players',
		tagline: 'Trick-taking missions with task cards on the ocean floor.',
		description:
			'Cooperative trick-taking on the ocean floor: submarines, sonar and new tasks. No talking about your cards.'
	}
];

/**
 * Where a game's tile links to. In production every app sits on the same host behind Caddy, so
 * it is the game's path. In dev each app runs its own Vite server, so it is the game's dev port on
 * the hostname the start page was opened with (localhost, or a LAN IP when testing on a phone).
 */
export function gameHref(
	game: Game,
	url: Pick<URL, 'protocol' | 'hostname'>,
	dev: boolean
): string {
	return dev ? `${url.protocol}//${url.hostname}:${game.devPort}/` : game.path;
}
