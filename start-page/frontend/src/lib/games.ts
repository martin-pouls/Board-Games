/** A game listed on the start page. Each game is its own app, served by Caddy next to this one. */
export interface Game {
	id: 'the-game' | 'regicide' | 'fox-duet';
	name: string;
	/** Absolute path of the game's app on the same host in production (outside this app's base path). */
	path: `/${string}/`;
	/** Port of the game's Vite dev server (`server.port` in its vite.config.ts), linked to in dev. */
	devPort: number;
	players: string;
	tagline: string;
	features: string[];
}

export const games: Game[] = [
	{
		id: 'the-game',
		name: 'The Game',
		path: '/the-game/',
		devPort: 5173,
		players: '1–5 players',
		tagline:
			'Play all 98 cards onto two rising and two falling piles, without saying which numbers you hold.',
		features: ['Pro variant', 'The Game Extreme']
	},
	{
		id: 'regicide',
		name: 'Regicide',
		path: '/regicide/',
		devPort: 5174,
		players: '1–4 players',
		tagline:
			'Defeat twelve corrupted royals with a standard deck, using the powers of the four suits.',
		features: ['Solo mode with Jesters', 'Gold, Silver and Bronze victories']
	},
	{
		id: 'fox-duet',
		name: 'Fox Duet',
		path: '/fox-duet/',
		devPort: 5176,
		players: '2 players',
		tagline:
			'The Fox in the Forest: Duet. Cooperative trick-taking: work together to win the right tricks.',
		features: ['3 difficulty levels']
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
