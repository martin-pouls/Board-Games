/** A game listed on the start page. Each game is its own app, served by Caddy next to this one. */
export interface Game {
	id: 'the-game' | 'regicide';
	name: string;
	/** Absolute path of the game's app on the same host (outside this app's base path). */
	href: `/${string}/`;
	players: string;
	tagline: string;
	features: string[];
}

export const games: Game[] = [
	{
		id: 'the-game',
		name: 'The Game',
		href: '/the-game/',
		players: '1–5 players',
		tagline:
			'Play all 98 cards onto two rising and two falling piles, without saying which numbers you hold.',
		features: ['Pro variant', 'The Game Extreme']
	},
	{
		id: 'regicide',
		name: 'Regicide',
		href: '/regicide/',
		players: '1–4 players',
		tagline:
			'Defeat twelve corrupted royals with a standard deck, using the powers of the four suits.',
		features: ['Solo mode with Jesters', 'Gold, Silver and Bronze victories']
	}
];
