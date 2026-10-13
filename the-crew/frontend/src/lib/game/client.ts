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
	| 'me'
	| 'isHost'
	| 'setOptions'
	| 'selectMission'
	| 'startMission'
	| 'adjustTokens'
	| 'skipTokens'
	| 'answerPoll'
	| 'pick'
	| 'pickTask'
	| 'sendDistress'
	| 'skipDistress'
	| 'choosePassCard'
	| 'handover'
	| 'communicate'
	| 'playCard'
	| 'giveUp'
	| 'nextAttempt'
	| 'nextMission'
	| 'backToLobby'
	| 'endGame'
>;
