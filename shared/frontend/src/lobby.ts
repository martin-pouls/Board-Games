// Types the shared lobby pages (HomePage, JoinPage) need from a game's REST responses. A game's
// own GameSummary and SessionResponse satisfy them structurally.

/** The part of a game's summary (GET /api/games/{code}) the lobby pages read. */
export interface LobbySummary {
	phase: string;
	hostName: string;
	playerCount: number;
	maxPlayers: number;
	joinable: boolean;
	/** RFC3339 creation time. */
	createdAt?: string;
}

/** The part of the response to create/join the lobby pages store as the player's session. */
export interface LobbySession {
	gameId: string;
	playerId: string;
	token: string;
}

/** SvelteKit's goto, as the pages take it (the game passes `goto` from $app/navigation). */
export type Goto = (href: string, opts?: { replaceState?: boolean }) => Promise<unknown>;
