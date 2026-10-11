// Thin client for the Go REST API. In dev, Vite proxies /api and /ws to the backend.
// Paths are prefixed with the build-time base path (BASE_PATH, e.g. /regicide in production).
// The request helper and the lobby endpoints are shared (shared/frontend api.ts).
import { createRequest, gamePath as game, lobbyApi } from '@board-games/shared/api.js';
import type {
	ApiErrorCode,
	Card,
	JesterSignal,
	GameSummary,
	PlayerView,
	SessionResponse
} from '#lib/types.js';

export { ApiError } from '@board-games/shared/api.js';

const request = createRequest<ApiErrorCode>(__BASE_PATH__);

export const { getHealth, createGame, joinGame, getSummary, getView, leaveGame, giveUp, rematch } =
	lobbyApi<SessionResponse, GameSummary, PlayerView>(request);

export function startGame(
	code: string,
	token: string,
	firstPlayerId?: string
): Promise<PlayerView> {
	return request('POST', `${game(code)}/start`, { token, body: { firstPlayerId } });
}
export function play(code: string, token: string, cards: Card[]): Promise<PlayerView> {
	return request('POST', `${game(code)}/play`, { token, body: { cards } });
}
export function yieldTurn(code: string, token: string): Promise<PlayerView> {
	return request('POST', `${game(code)}/yield`, { token });
}
export function discard(code: string, token: string, cards: Card[]): Promise<PlayerView> {
	return request('POST', `${game(code)}/discard`, { token, body: { cards } });
}
export function chooseNext(code: string, token: string, playerId: string): Promise<PlayerView> {
	return request('POST', `${game(code)}/next-player`, { token, body: { playerId } });
}
export function useSoloJester(code: string, token: string): Promise<PlayerView> {
	return request('POST', `${game(code)}/solo-jester`, { token });
}
export function setSignal(code: string, token: string, signal: JesterSignal): Promise<PlayerView> {
	return request('PUT', `${game(code)}/signal`, { token, body: { signal } });
}
