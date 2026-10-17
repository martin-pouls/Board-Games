import { GameSocket } from '@board-games/shared/connection.svelte.js';
import type {
	ApiErrorCode,
	Card,
	ClientCommand,
	Direction,
	GameEvent,
	GameOptions,
	PlayerView,
	Position
} from '#lib/types.js';

export { ActionError, type ConnectionStatus } from '@board-games/shared/connection.svelte.js';

/** What the host chooses for the next attempt after a loss (Deep Sea). */
export interface NextAttemptChoice {
	newTasks?: boolean;
	uncounted?: boolean;
}

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

	/** Host only, lobby only. */
	setOptions(options: GameOptions): Promise<void> {
		return this.send({ type: 'setOptions', options });
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
	/** Deep Sea: leave the card in the draft. */
	passTask(): Promise<void> {
		return this.send({ type: 'passTask' });
	}
	/** Deep Sea, captain: put a displayed task back and lay out another of the same difficulty. */
	replaceTask(taskId: number): Promise<void> {
		return this.send({ type: 'replaceTask', taskId });
	}
	/** Deep Sea: note how many tricks the owner of a prediction task will win. */
	predict(taskId: number, tricks: number): Promise<void> {
		return this.send({ type: 'predict', taskId, tricks });
	}
	/** Deep Sea, captain: give a task to a member (free selection, the split of mission 26). */
	assignTask(taskId: number, seat: number): Promise<void> {
		return this.send({ type: 'assignTask', taskId, seat });
	}
	/**
	 * After a loss. Deep Sea: `newTasks` draws new task cards, `uncounted` leaves the attempt out
	 * of the logbook. Without a choice no flag is sent (Planet Nine, and the defaults).
	 */
	nextAttempt(choice: NextAttemptChoice = {}): Promise<void> {
		const cmd: Extract<ClientCommand, { type: 'nextAttempt' }> = { type: 'nextAttempt' };
		if (choice.newTasks) cmd.newTasks = true;
		if (choice.uncounted) cmd.uncounted = true;
		return this.send(cmd);
	}
	nextMission(): Promise<void> {
		return this.send({ type: 'nextMission' });
	}
	/** Deep Sea: `uncounted` leaves the lost attempt out of the logbook; false sends no flag. */
	backToLobby(uncounted = false): Promise<void> {
		return this.send(uncounted ? { type: 'backToLobby', uncounted } : { type: 'backToLobby' });
	}
	endGame(): Promise<void> {
		return this.send({ type: 'endGame' });
	}
}
