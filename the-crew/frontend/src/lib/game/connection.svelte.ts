import { GameSocket } from '@board-games/shared/connection.svelte.js';
import type { ApiErrorCode, ClientCommand, GameEvent, PlayerView } from '#lib/types.js';

export { ActionError, type ConnectionStatus } from '@board-games/shared/connection.svelte.js';

/**
 * The live connection to one The Crew game. Connecting, reconnecting, acks and the view live in
 * the shared GameSocket; this adds The Crew's commands.
 */
export class GameConnection extends GameSocket<PlayerView, GameEvent, ClientCommand, ApiErrorCode> {
	constructor(code: string, token: string) {
		super(code, token, __BASE_PATH__);
	}

	endGame(): Promise<void> {
		return this.send({ type: 'endGame' });
	}
}
