// TypeScript mirror of the Go wire contract (backend/internal/adapters/apidto). Keep in sync.
import type {
	ClientErrorCode,
	CommonErrorCode,
	ErrorBody as SharedErrorBody,
	ServerMessage as SharedServerMessage
} from '@board-games/shared/wire.js';

export type Phase = 'lobby' | 'playing' | 'won' | 'lost' | 'ended';

export interface PlayerSummary {
	id: string;
	name: string;
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
/** Sky Team's own error codes, next to the ones every game shares (shared/frontend wire.ts). */
export type GameErrorCode =
	| 'not_host'
	| 'wrong_phase'
	| 'game_full'
	| 'name_taken'
	| 'cannot_leave'
	| 'invalid_name'
	| 'unknown_player';
export type ApiErrorCode = CommonErrorCode | ClientErrorCode | GameErrorCode;
export type ErrorBody = SharedErrorBody<ApiErrorCode>;
export type ServerMessage = SharedServerMessage<PlayerView, GameEvent, ApiErrorCode>;
export type ClientCommand = { type: 'endGame' };
// sent as ClientCommand & { requestId: string }
