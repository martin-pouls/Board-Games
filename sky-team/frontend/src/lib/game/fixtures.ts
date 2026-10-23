// Sample views for tests and the dev fixtures route (/dev/fixtures). Not in production builds.
import type { PlayerView } from '#lib/types.js';

/** A lobby for Ann (host, seat 0) with the given number of seats taken. */
function lobby(players: 1 | 2): PlayerView {
	const seated = [
		{ id: 'ann', name: 'Ann', seat: 0, isHost: true, isYou: true },
		{ id: 'ben', name: 'Ben', seat: 1, isHost: false, isYou: false }
	];
	return {
		gameId: 'K7MQ2',
		revision: players,
		phase: 'lobby',
		you: 'ann',
		hostId: 'ann',
		players: seated.slice(0, players)
	};
}

/** The same lobby seen by the second player. */
function asBen(v: PlayerView): PlayerView {
	return {
		...v,
		you: 'ben',
		players: v.players.map((p) => ({ ...p, isYou: p.id === 'ben' }))
	};
}

/** The host alone in the lobby. */
export const lobbyAloneView = () => lobby(1);
/** Both seats taken, seen by the host. */
export const lobbyHostView = () => lobby(2);
/** Both seats taken, seen by the guest. */
export const lobbyGuestView = () => asBen(lobby(2));
/** The host closed the lobby (seen by the guest). */
export const endedView = (): PlayerView => ({ ...lobbyGuestView(), phase: 'ended' });

/**
 * Every named fixture; each call returns a fresh view. The dev route `/dev/fixtures/<name>`
 * shows them on the game screen.
 */
export const fixtures: Record<string, () => PlayerView> = {
	lobbyAloneView,
	lobbyHostView,
	lobbyGuestView,
	endedView
};
