import { GameSocket } from '@board-games/shared/connection.svelte.js';
import type {
	ApiErrorCode,
	Card,
	ClientCommand,
	GameEvent,
	JesterSignal,
	PlayerView
} from '#lib/types.js';

export { ActionError, type ConnectionStatus } from '@board-games/shared/connection.svelte.js';

/**
 * The live connection to one Regicide game. Connecting, reconnecting, acks and the view live in
 * the shared GameSocket; this adds Regicide's commands.
 */
export class GameConnection extends GameSocket<PlayerView, GameEvent, ClientCommand, ApiErrorCode> {
	constructor(code: string, token: string) {
		super(code, token, __BASE_PATH__);
	}

	get isMyTurn(): boolean {
		return this.view?.phase === 'playing' && this.view.turn?.activePlayerId === this.view.you;
	}

	startGame(firstPlayerId?: string): Promise<void> {
		return this.send({ type: 'startGame', firstPlayerId });
	}
	play(cards: Card[]): Promise<void> {
		return this.send({ type: 'play', cards });
	}
	yieldTurn(): Promise<void> {
		return this.send({ type: 'yield' });
	}
	discard(cards: Card[]): Promise<void> {
		return this.send({ type: 'discard', cards });
	}
	chooseNext(playerId: string): Promise<void> {
		return this.send({ type: 'chooseNext', playerId });
	}
	useSoloJester(): Promise<void> {
		return this.send({ type: 'useSoloJester' });
	}
	setSignal(signal: JesterSignal): Promise<void> {
		return this.send({ type: 'setSignal', signal });
	}
	endGame(): Promise<void> {
		return this.send({ type: 'endGame' });
	}
	giveUp(): Promise<void> {
		return this.send({ type: 'giveUp' });
	}
	rematch(): Promise<void> {
		return this.send({ type: 'rematch' });
	}
}
