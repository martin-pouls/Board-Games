// TypeScript mirror of the Go wire contract (backend/internal/adapters/apidto). Keep in sync.
import type {
	ClientErrorCode,
	CommonErrorCode,
	ErrorBody as SharedErrorBody,
	ServerMessage as SharedServerMessage
} from '@board-games/shared/wire.js';

export type Phase = 'lobby' | 'playing' | 'debrief' | 'ended';

/** The four colours and trump ("rockets" in the UI). */
export type Suit = 'pink' | 'blue' | 'green' | 'yellow' | 'trump';
/** A playing card: colours 1–9, trump 1–4. */
export interface Card {
	suit: Suit;
	rank: number;
}
/** Task tokens: 1–5 (fixed position), omega (last), arrow1–arrow4 (relative order). */
export type Token =
	'' | '1' | '2' | '3' | '4' | '5' | 'omega' | 'arrow1' | 'arrow2' | 'arrow3' | 'arrow4';

export interface PlayerSummary {
	id: string;
	name: string;
	/** Join order = seat order. */
	seat: number;
	isHost: boolean;
	isYou: boolean;
}
export interface PlayerView {
	gameId: string;
	revision: number;
	phase: Phase;
	you: string;
	hostId: string;
	players: PlayerSummary[];
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
	| { type: 'gameEnded'; byPlayerId: string };
/** The Crew's own error codes, next to the ones every game shares (shared/frontend wire.ts). */
export type GameErrorCode =
	| 'not_host'
	| 'wrong_phase'
	| 'game_full'
	| 'name_taken'
	| 'cannot_leave'
	| 'invalid_name'
	| 'unknown_player'
	| 'not_enough_players';
export type ApiErrorCode = CommonErrorCode | ClientErrorCode | GameErrorCode;
export type ErrorBody = SharedErrorBody<ApiErrorCode>;
export type ServerMessage = SharedServerMessage<PlayerView, GameEvent, ApiErrorCode>;
export type ClientCommand = { type: 'endGame' };
// sent as ClientCommand & { requestId: string }
