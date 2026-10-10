import type { GameConnection } from './connection.svelte.js';

/**
 * The part of a GameConnection the game components use. Components depend on this
 * instead of the concrete class so they can be rendered from fixtures with a fake.
 */
export type GameClient = Pick<
	GameConnection,
	| 'code'
	| 'view'
	| 'online'
	| 'status'
	| 'lastEvents'
	| 'me'
	| 'isHost'
	| 'isMyTurn'
	| 'setOptions'
	| 'startGame'
	| 'chooseFirstPlayer'
	| 'playCard'
	| 'undoPlay'
	| 'endTurn'
	| 'giveUp'
	| 'endGame'
	| 'rematch'
>;
