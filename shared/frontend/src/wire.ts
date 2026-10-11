// TypeScript mirror of the JSON envelope every game speaks (shared/backend/wire). The game's
// view, events, commands and its own error codes go inside it.

/** Error codes every game's server sends (shared/backend/wire Code*). */
export type CommonErrorCode =
	| 'bad_request'
	| 'unauthorized'
	| 'game_not_found'
	| 'not_found'
	| 'conflict'
	| 'rate_limited'
	// sent by the access gate in front of the apps, not by the game server
	| 'login_required'
	| 'internal';

/** Error codes the client produces itself (never sent by the server). */
export type ClientErrorCode = 'network' | 'disconnected' | 'timeout';

export interface ErrorBody<C extends string = string> {
	code: C;
	message: string;
}

/** A server -> client WebSocket message carrying the game's view V and events E. */
export type ServerMessage<V, E, C extends string = string> =
	| { type: 'view'; view: V; events: E[] }
	| { type: 'presence'; online: string[] }
	| { type: 'ack'; requestId: string }
	| { type: 'error'; requestId?: string; error: ErrorBody<C> };

/** The part of a player view the shared connection relies on. */
export interface BaseView {
	/** Increases with every committed change; older views are dropped. */
	revision: number;
	/** The viewer's player ID. */
	you: string;
	hostId: string;
	players: { id: string; isYou: boolean }[];
}

/** WebSocket close codes after which the client must not reconnect (shared/backend/wskit). */
export const CLOSE_UNAUTHORIZED = 4401;
export const CLOSE_GAME_NOT_FOUND = 4404;
