import { GameSocket } from '@board-games/shared/connection.svelte.js';
import type {
	ApiErrorCode,
	CardRef,
	ClientCommand,
	End,
	GameEvent,
	PlayerView
} from '#lib/types.js';

export { ActionError, type ConnectionStatus } from '@board-games/shared/connection.svelte.js';

/** Only suit and rank go over the wire. */
const ref = (c: CardRef): CardRef => ({ suit: c.suit, rank: c.rank });

/**
 * The live connection to one Fox Duet game. Connecting, reconnecting, acks and the view live in
 * the shared GameSocket; this adds Fox Duet's commands.
 */
export class GameConnection extends GameSocket<PlayerView, GameEvent, ClientCommand, ApiErrorCode> {
	constructor(code: string, token: string) {
		super(code, token, __BASE_PATH__);
	}

	/** The game waits for the viewer. */
	get mustAct(): boolean {
		return this.view?.pending?.playerIds.includes(this.view.you) ?? false;
	}

	startGame(level: number): Promise<void> {
		return this.send({ type: 'startGame', level });
	}
	playCard(card: CardRef): Promise<void> {
		return this.send({ type: 'playCard', card: ref(card) });
	}
	chooseFoxesPlayer(playerId: string): Promise<void> {
		return this.send({ type: 'chooseFoxesPlayer', playerId });
	}
	swapDecree(card: CardRef): Promise<void> {
		return this.send({ type: 'swapDecree', card: ref(card) });
	}
	keepDecree(): Promise<void> {
		return this.send({ type: 'keepDecree' });
	}
	giveGift(card: CardRef): Promise<void> {
		return this.send({ type: 'giveGift', card: ref(card) });
	}
	chooseMove(delta: number): Promise<void> {
		return this.send({ type: 'chooseMove', delta });
	}
	chooseForestEnd(end: End): Promise<void> {
		return this.send({ type: 'chooseForestEnd', end });
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
