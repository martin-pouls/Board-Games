import { GameSocket } from '@board-games/shared/connection.svelte.js';
import type {
	AbilityId,
	ApiErrorCode,
	ClientCommand,
	GameEvent,
	PlayerView,
	SpaceId
} from '#lib/types.js';

export { ActionError, type ConnectionStatus } from '@board-games/shared/connection.svelte.js';

/**
 * The live connection to one Sky Team game. Connecting, reconnecting, acks and the view live in
 * the shared GameSocket; this adds Sky Team's commands.
 */
export class GameConnection extends GameSocket<PlayerView, GameEvent, ClientCommand, ApiErrorCode> {
	constructor(code: string, token: string) {
		super(code, token, __BASE_PATH__);
	}

	/** The game waits for the viewer (the server's `mustAct` for their seat). */
	get mustAct(): boolean {
		return this.view?.players.find((p) => p.id === this.view?.you)?.mustAct ?? false;
	}

	configure(scenario: string, abilities: AbilityId[]): Promise<void> {
		return this.send({ type: 'configure', scenario, abilities });
	}
	swapRoles(): Promise<void> {
		return this.send({ type: 'swapRoles' });
	}
	startGame(): Promise<void> {
		return this.send({ type: 'startGame' });
	}
	ready(): Promise<void> {
		return this.send({ type: 'ready' });
	}
	/** `coffee` is the net modifier; `tokenSpace` only when `space` is an intern space. */
	place(slot: number, space: SpaceId, coffee: number, tokenSpace?: SpaceId): Promise<void> {
		return this.send(
			tokenSpace
				? { type: 'place', slot, space, coffee, tokenSpace }
				: { type: 'place', slot, space, coffee }
		);
	}
	spendReroll(): Promise<void> {
		return this.send({ type: 'spendReroll' });
	}
	confirmReroll(slots: number[]): Promise<void> {
		return this.send({ type: 'confirmReroll', slots });
	}
	adapt(slot: number): Promise<void> {
		return this.send({ type: 'adapt', slot });
	}
	anticipate(slot: number): Promise<void> {
		return this.send({ type: 'anticipate', slot });
	}
	offerSwap(slot: number): Promise<void> {
		return this.send({ type: 'offerSwap', slot });
	}
	answerSwap(slot: number): Promise<void> {
		return this.send({ type: 'answerSwap', slot });
	}
	placeSync(space: SpaceId): Promise<void> {
		return this.send({ type: 'placeSync', space });
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
