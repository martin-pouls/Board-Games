import { GameSocket } from '@board-games/shared/connection.svelte.js';
import type {
	ApiErrorCode,
	ClientCommand,
	GameEvent,
	GameOptions,
	PlayerView
} from '#lib/types.js';

export { ActionError, type ConnectionStatus } from '@board-games/shared/connection.svelte.js';

/**
 * The live connection to one game of The Game. Connecting, reconnecting, acks and the view live in
 * the shared GameSocket; this adds The Game's commands.
 */
export class GameConnection extends GameSocket<PlayerView, GameEvent, ClientCommand, ApiErrorCode> {
	constructor(code: string, token: string) {
		super(code, token, __BASE_PATH__);
	}

	get isMyTurn(): boolean {
		return this.view?.phase === 'playing' && this.view.turn?.activePlayerId === this.view.you;
	}

	setOptions(options: GameOptions): Promise<void> {
		return this.send({ type: 'setOptions', options });
	}
	startGame(): Promise<void> {
		return this.send({ type: 'startGame' });
	}
	chooseFirstPlayer(playerId: string): Promise<void> {
		return this.send({ type: 'chooseFirstPlayer', playerId });
	}
	playCard(card: number, pile: number): Promise<void> {
		return this.send({ type: 'playCard', card, pile });
	}
	undoPlay(): Promise<void> {
		return this.send({ type: 'undoPlay' });
	}
	endTurn(): Promise<void> {
		return this.send({ type: 'endTurn' });
	}
	giveUp(): Promise<void> {
		return this.send({ type: 'giveUp' });
	}
	endGame(): Promise<void> {
		return this.send({ type: 'endGame' });
	}
	rematch(): Promise<void> {
		return this.send({ type: 'rematch' });
	}
}
