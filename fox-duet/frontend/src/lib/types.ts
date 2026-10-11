// TypeScript mirror of the Go wire contract (backend/internal/adapters/apidto). Keep in sync.
import type {
	ClientErrorCode,
	CommonErrorCode,
	ErrorBody as SharedErrorBody,
	ServerMessage as SharedServerMessage
} from '@board-games/shared/wire.js';

export type Phase = 'lobby' | 'playing' | 'won' | 'lost' | 'ended';
export type Step = 'lead' | 'follow' | 'foxesChoose' | 'foxesSwap' | 'gift' | 'move' | 'forestEnd';
export type Suit = 'doves' | 'roses' | 'stars';
export type Ability = '' | 'musician' | 'foxes' | 'gazelle' | 'gift' | 'royalHeir';
/** The left end belongs to seat 0, the right end to seat 1. */
export type End = 'left' | 'right';
export type LossReason = 'lostInForest' | 'outOfTime' | 'gaveUp';

/** A card. Movement and ability follow from the rank; the server sends them along. */
export interface Card {
	suit: Suit;
	rank: number;
	movement: number;
	ability: Ability;
}
export interface PlayerSummary {
	id: string;
	name: string;
	seat: number;
	end: End;
	handCount: number;
	isHost: boolean;
	isYou: boolean;
	isDealer: boolean;
	mustAct: boolean;
	/** Only once the game is over. */
	revealedHand?: Card[];
}
export interface Space {
	gems: number;
	plus: boolean;
	covered: boolean;
}
/** Space indices run from the left end to the right end. */
export interface Board {
	side: 'A' | 'B';
	start: number;
	spaces: Space[];
	traveller: number;
	forestTokens: number;
	forestTokensTotal: number;
	gemsLeft: number;
}
export interface Trick {
	number: number;
	leaderId: string;
	lead: Card | null;
	follow: Card | null;
	/** Once both cards are on the table. */
	winnerId?: string;
}
export interface LastTrick {
	number: number;
	leaderId: string;
	lead: Card;
	follow: Card;
	winnerId: string;
	from: number;
	to: number;
	delta: number;
	beyond: End | '';
	collected: boolean;
}
export interface MoveOption {
	/** Negative = toward the left end. */
	delta: number;
	/** Landing space; the start after leaving the path; the old space when lost. */
	to: number;
	beyond: End | '';
	gem: boolean;
	lost: boolean;
}
export interface Pending {
	step: Step;
	playerIds: string[];
	chosenId?: string;
	moveOptions: MoveOption[];
	ends: End[];
	yourGift: Card | null;
}
export interface Gift {
	gave: Card;
	received: Card;
}
export interface Score {
	level: number;
	cards: number;
	earlyWin: number;
	forestTokens: number;
	total: number;
}
export interface Result {
	outcome: 'won' | 'lost';
	reason?: LossReason;
	playerId?: string;
	round: number;
	gemsLeft: number;
	score: Score | null;
}
export interface PlayerView {
	gameId: string;
	revision: number;
	phase: Phase;
	you: string;
	hostId: string;
	/** 0 in the lobby. */
	level: number;
	players: PlayerSummary[];
	hand: Card[];
	/** The viewer's legal cards when they lead or follow now. */
	playable: Card[];
	board: Board | null;
	round: number;
	decree: Card | null;
	trick: Trick | null;
	lastTrick: LastTrick | null;
	pending: Pending | null;
	lastGift: Gift | null;
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
	/** RFC3339 UTC; absent when unknown. */
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
	| { type: 'gameStarted'; level: number }
	| { type: 'roundStarted'; round: number; dealerId: string; card: Card }
	| { type: 'cardPlayed'; playerId: string; card: Card; lead?: boolean }
	| { type: 'foxesChosen'; byPlayerId: string; playerId: string; skipped?: boolean }
	| { type: 'decreeSwapped'; playerId: string; old: Card; new: Card }
	| { type: 'decreeKept'; playerId: string }
	| { type: 'giftChosen'; playerId: string; auto?: boolean }
	| { type: 'giftExchanged' }
	| { type: 'trickWon'; playerId: string; number: number }
	| { type: 'travellerMoved'; from: number; to: number; delta: number; end?: End }
	| {
			type: 'forestTokenPlaced';
			end: End;
			space: number;
			gemsMoved: number;
			roundEnd?: boolean;
			left: number;
	  }
	| { type: 'gemCollected'; space: number; left: number }
	| { type: 'gemsAdded'; count: number }
	| { type: 'gameWon'; score: number }
	| { type: 'gameLost'; reason: LossReason; playerId?: string }
	| { type: 'gameEnded'; byPlayerId: string }
	| { type: 'rematchCreated'; gameId: string };
/** Fox Duet's own error codes, next to the ones every game shares (shared/frontend wire.ts). */
export type GameErrorCode =
	| 'not_host'
	| 'wrong_phase'
	| 'wrong_step'
	| 'game_full'
	| 'name_taken'
	| 'not_your_turn'
	| 'cannot_leave'
	| 'invalid_name'
	| 'unknown_player'
	| 'not_enough_players'
	| 'invalid_level'
	| 'card_not_in_hand'
	| 'must_follow_suit'
	| 'gift_already_chosen'
	| 'illegal_move'
	| 'illegal_end'
	| 'rematch_exists';
export type ApiErrorCode = CommonErrorCode | ClientErrorCode | GameErrorCode;
export type ErrorBody = SharedErrorBody<ApiErrorCode>;
export type ServerMessage = SharedServerMessage<PlayerView, GameEvent, ApiErrorCode>;
/** Wire cards on input need only suit and rank. */
export type CardRef = Pick<Card, 'suit' | 'rank'>;
export type ClientCommand =
	| { type: 'startGame'; level: number }
	| { type: 'playCard'; card: CardRef }
	| { type: 'chooseFoxesPlayer'; playerId: string }
	| { type: 'swapDecree'; card: CardRef }
	| { type: 'keepDecree' }
	| { type: 'giveGift'; card: CardRef }
	| { type: 'chooseMove'; delta: number }
	| { type: 'chooseForestEnd'; end: End }
	| { type: 'giveUp' }
	| { type: 'endGame' }
	| { type: 'rematch' };
// sent as ClientCommand & { requestId: string }
