export type Phase = 'lobby' | 'choosingFirstPlayer' | 'playing' | 'finished';
export type Direction = 'up' | 'down';
export type Outcome = 'won' | 'lost';
export interface Card {
	value: number;
	kind: 'number';
	command?: string;
}
export interface HandCard extends Card {
	legalPiles: number[];
}
export interface GameOptions {
	variant: 'classic';
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
	plays: Play[];
}
export interface Result {
	outcome: Outcome;
	score: number;
	excellent: boolean;
	cardsLeftInHands: number;
	cardsLeftInDrawPile: number;
	stuckPlayerId?: string;
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
}
export interface GameSummary {
	gameId: string;
	phase: Phase;
	hostName: string;
	playerCount: number;
	maxPlayers: number;
	joinable: boolean;
}
export interface SessionResponse {
	gameId: string;
	playerId: string;
	token: string;
	view: PlayerView;
}
export type GameEvent =
	| { type: 'playerJoined'; playerId: string; name: string }
	| { type: 'optionsChanged' }
	| { type: 'rematchCreated'; gameId: string }
	| { type: 'gameStarted' }
	| { type: 'firstPlayerChosen'; playerId: string }
	| { type: 'turnStarted'; playerId: string; number: number }
	| { type: 'cardPlayed'; playerId: string; card: Card; pile: number }
	| { type: 'turnEnded'; playerId: string; cardsDrawn: number; auto?: boolean }
	| { type: 'playerSkipped'; playerId: string }
	| { type: 'gameFinished'; outcome: Outcome; score: number; stuckPlayerId?: string };
export type ApiErrorCode =
	| 'bad_request'
	| 'unauthorized'
	| 'not_host'
	| 'game_not_found'
	| 'not_found'
	| 'wrong_phase'
	| 'game_full'
	| 'name_taken'
	| 'not_your_turn'
	| 'conflict'
	| 'invalid_name'
	| 'invalid_options'
	| 'unknown_player'
	| 'card_not_in_hand'
	| 'invalid_pile'
	| 'illegal_placement'
	| 'minimum_not_reached'
	| 'rematch_exists'
	| 'internal'
	| 'network'
	| 'disconnected'
	| 'timeout'; // last three are client-side only
export interface ErrorBody {
	code: ApiErrorCode;
	message: string;
}
export type ServerMessage =
	| { type: 'view'; view: PlayerView; events: GameEvent[] }
	| { type: 'presence'; online: string[] }
	| { type: 'ack'; requestId: string }
	| { type: 'error'; requestId?: string; error: ErrorBody };
export type ClientCommand =
	| { type: 'setOptions'; options: GameOptions }
	| { type: 'startGame' }
	| { type: 'chooseFirstPlayer'; playerId: string }
	| { type: 'playCard'; card: number; pile: number }
	| { type: 'endTurn' }
	| { type: 'rematch' };
// sent as ClientCommand & { requestId: string }
