import { GameSocket } from '@board-games/shared/connection.svelte.js';
import type {
	ApiErrorCode,
	Card,
	ClientCommand,
	Direction,
	GameEvent,
	PlayerView,
	Position
} from '#lib/types.js';

export { ActionError, type ConnectionStatus } from '@board-games/shared/connection.svelte.js';

/** Only suit and rank go over the wire. */
const ref = (c: Card): Card => ({ suit: c.suit, rank: c.rank });

/**
 * The live connection to one The Crew game. Connecting, reconnecting, acks and the view live in
 * the shared GameSocket; this adds The Crew's commands.
 */
export class GameConnection extends GameSocket<PlayerView, GameEvent, ClientCommand, ApiErrorCode> {
	constructor(code: string, token: string) {
		super(code, token, __BASE_PATH__);
	}

	selectMission(mission: number): Promise<void> {
		return this.send({ type: 'selectMission', mission });
	}
	startMission(): Promise<void> {
		return this.send({ type: 'startMission' });
	}
	/** Mission 23: swap the tokens of two tasks; mission 40: move a token to a task without one. */
	adjustTokens(from: number, to: number): Promise<void> {
		return this.send({ type: 'adjustTokens', from, to });
	}
	skipTokens(): Promise<void> {
		return this.send({ type: 'skipTokens' });
	}
	answerPoll(yes: boolean): Promise<void> {
		return this.send({ type: 'answerPoll', yes });
	}
	/** One seat; two for the roles of mission 50. */
	pick(seats: number[]): Promise<void> {
		return this.send({ type: 'pick', seats });
	}
	pickTask(taskId: number): Promise<void> {
		return this.send({ type: 'pickTask', taskId });
	}
	sendDistress(direction: Direction): Promise<void> {
		return this.send({ type: 'sendDistress', direction });
	}
	skipDistress(): Promise<void> {
		return this.send({ type: 'skipDistress' });
	}
	choosePassCard(card: Card): Promise<void> {
		return this.send({ type: 'choosePassCard', card: ref(card) });
	}
	handover(taskId: number, seat: number): Promise<void> {
		return this.send({ type: 'handover', taskId, seat });
	}
	communicate(card: Card, position: Position): Promise<void> {
		return this.send({ type: 'communicate', card: ref(card), position });
	}
	playCard(card: Card): Promise<void> {
		return this.send({ type: 'playCard', card: ref(card) });
	}
	giveUp(): Promise<void> {
		return this.send({ type: 'giveUp' });
	}
	nextAttempt(): Promise<void> {
		return this.send({ type: 'nextAttempt' });
	}
	nextMission(): Promise<void> {
		return this.send({ type: 'nextMission' });
	}
	backToLobby(): Promise<void> {
		return this.send({ type: 'backToLobby' });
	}
	endGame(): Promise<void> {
		return this.send({ type: 'endGame' });
	}
}
