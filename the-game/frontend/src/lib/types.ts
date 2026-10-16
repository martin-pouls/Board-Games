import type {
	ClientErrorCode,
	CommonErrorCode,
	ErrorBody as SharedErrorBody,
	ServerMessage as SharedServerMessage
} from '@board-games/shared/wire.js';

export type Phase = 'lobby' | 'choosingFirstPlayer' | 'playing' | 'won' | 'lost' | 'ended';
export type Direction = 'up' | 'down';
export type LossReason = 'stuck' | 'gaveUp';
export type Variant = 'classic' | 'extreme';
export type CommandName =
	| 'stop'
	| 'skull'
	| 'exactlyThree'
	| 'noCommunication'
	| 'noBackwardsTrick'
	| 'singlePile'
	| 'drawOnlyOne';
export interface Card {
	value: number;
	kind: 'number';
	command?: CommandName;
}
export interface HandCard extends Card {
	legalPiles: number[];
	/** Code of the command that blocks this card (only when legalPiles is empty). */
	blocked?: ApiErrorCode;
	/** Piles the base rules allow but a command forbids, with the reason. */
	vetoes?: PileVeto[];
}
export interface PileVeto {
	pile: number;
	code: ApiErrorCode;
}
export interface GameOptions {
	variant: Variant;
	pro: boolean;
	reducedHandSize: boolean;
}
export interface Rules {
	minCardsPerTurn: number;
	minCardsDrawPileEmpty: number;
	handSize: number;
	minPlayers: number;
	maxPlayers: number;
}
export interface PlayerSummary {
	id: string;
	name: string;
	seat: number;
	handCount: number;
	isHost: boolean;
	isYou: boolean;
	isActive: boolean;
	revealedHand?: Card[];
}
export interface Pile {
	index: number;
	direction: Direction;
	startValue: number;
	topValue: number;
	topCard: Card | null;
	cardCount: number;
	/** topCard.command is in effect now (Infinity top, or lightning top played this turn). */
	commandActive: boolean;
}
export interface ActiveCommand {
	command: CommandName;
	pile: number;
}
export interface TurnConstraints {
	/** 3 after a 3!, else null. */
	exactCount: number | null;
	singlePile: boolean;
	/** Pile all further cards must go on; null = not bound. */
	boundPile: number | null;
	skullPiles: number[];
	stopPlayed: boolean;
}
export interface Play {
	card: Card;
	pile: number;
}
export interface Turn {
	number: number;
	activePlayerId: string;
	cardsPlayed: number;
	requiredMinimum: number;
	canEndTurn: boolean;
	/** No card fits and the turn cannot end: the active player must undo or give up. */
	stuck: boolean;
	plays: Play[];
	constraints: TurnConstraints;
}
/** The result of a game that was won or lost (the view's phase says which). */
export interface Result {
	score: number;
	excellent: boolean;
	cardsLeftInHands: number;
	cardsLeftInDrawPile: number;
	/** Why the game was lost; absent when won. */
	reason?: LossReason;
	/** The stuck player or the one who gave up; absent when won. */
	playerId?: string;
}
export interface PlayerView {
	gameId: string;
	revision: number;
	phase: Phase;
	you: string;
	hostId: string;
	options: GameOptions;
	rules: Rules;
	players: PlayerSummary[];
	hand: HandCard[];
	piles: Pile[];
	drawPileCount: number;
	turn: Turn | null;
	communicationRestricted: boolean;
	result: Result | null;
	nextGameId?: string;
	/** Visible Infinity commands, in pile order. */
	activeCommands: ActiveCommand[];
}
export interface GameSummary {
	gameId: string;
	phase: Phase;
	hostName: string;
	playerCount: number;
	maxPlayers: number;
	joinable: boolean;
	/** RFC3339 UTC creation time; absent for very old games. */
	createdAt?: string;
}
export interface SessionResponse {
	gameId: string;
	playerId: string;
	token: string;
	view: PlayerView;
}
export type GameEvent =
	| { type: 'playerJoined'; playerId: string; name: string }
	| { type: 'playerLeft'; playerId: string; hostId: string }
	| { type: 'optionsChanged' }
	| { type: 'rematchCreated'; gameId: string }
	| { type: 'gameStarted' }
	| { type: 'firstPlayerChosen'; playerId: string }
	| { type: 'turnStarted'; playerId: string; number: number }
	| { type: 'cardPlayed'; playerId: string; card: Card; pile: number }
	| { type: 'playUndone'; playerId: string; card: Card; pile: number }
	| { type: 'turnEnded'; playerId: string; cardsDrawn: number }
	| { type: 'playerSkipped'; playerId: string }
	| { type: 'gameWon' }
	| { type: 'gameLost'; reason: LossReason; playerId: string; score: number }
	| { type: 'gameEnded'; byPlayerId: string };
/** The Game's own error codes, next to the ones every game shares (shared/frontend wire.ts). */
export type GameErrorCode =
	| 'not_host'
	| 'wrong_phase'
	| 'game_full'
	| 'name_taken'
	| 'not_your_turn'
	| 'invalid_name'
	| 'invalid_options'
	| 'unknown_player'
	| 'card_not_in_hand'
	| 'invalid_pile'
	| 'illegal_placement'
	| 'minimum_not_reached'
	| 'rematch_exists'
	| 'nothing_to_undo'
	| 'cannot_leave'
	| 'stop_played'
	| 'backwards_trick_blocked'
	| 'single_pile_violation'
	| 'exact_count_exceeded'
	| 'exact_count_unreachable'
	| 'exact_count_required'
	| 'stop_too_early'
	| 'skull_uncovered'
	| 'skull_uncoverable';
export type ApiErrorCode = CommonErrorCode | ClientErrorCode | GameErrorCode;
export type ErrorBody = SharedErrorBody<ApiErrorCode>;
export type ServerMessage = SharedServerMessage<PlayerView, GameEvent, ApiErrorCode>;
export type ClientCommand =
	| { type: 'setOptions'; options: GameOptions }
	| { type: 'startGame' }
	| { type: 'chooseFirstPlayer'; playerId: string }
	| { type: 'playCard'; card: number; pile: number }
	| { type: 'undoPlay' }
	| { type: 'endTurn' }
	| { type: 'giveUp' }
	| { type: 'endGame' }
	| { type: 'rematch' };
// sent as ClientCommand & { requestId: string }
