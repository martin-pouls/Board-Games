// Thin client for the Go REST API. In dev, Vite proxies /api and /ws to the backend.
// Paths are prefixed with the build-time base path (BASE_PATH, e.g. /the-game in production).
import { createRequest, gamePath as game, lobbyApi } from '@board-games/shared/api.js';
import type {
	ApiErrorCode,
	GameOptions,
	GameSummary,
	PlayerView,
	SessionResponse
} from '#lib/types.js';

export { ApiError } from '@board-games/shared/api.js';

const request = createRequest<ApiErrorCode>(__BASE_PATH__);

export const { getHealth, joinGame, getSummary, getView, giveUp, rematch } = lobbyApi<
	SessionResponse,
	GameSummary,
	PlayerView
>(request);

export function createGame(name: string, options?: GameOptions): Promise<SessionResponse> {
	return request('POST', '/api/games', { body: { name, options } });
}
export function setOptions(code: string, token: string, options: GameOptions): Promise<PlayerView> {
	return request('PUT', `${game(code)}/options`, { token, body: options });
}
export function startGame(code: string, token: string): Promise<PlayerView> {
	return request('POST', `${game(code)}/start`, { token });
}
export function chooseFirstPlayer(
	code: string,
	token: string,
	playerId: string
): Promise<PlayerView> {
	return request('POST', `${game(code)}/first-player`, { token, body: { playerId } });
}
export function playCard(
	code: string,
	token: string,
	card: number,
	pile: number
): Promise<PlayerView> {
	return request('POST', `${game(code)}/plays`, { token, body: { card, pile } });
}
export function undoPlay(code: string, token: string): Promise<PlayerView> {
	return request('POST', `${game(code)}/undo`, { token });
}
export function endTurn(code: string, token: string): Promise<PlayerView> {
	return request('POST', `${game(code)}/end-turn`, { token });
}
export function endGame(code: string, token: string): Promise<PlayerView> {
	return request('POST', `${game(code)}/end-game`, { token });
}
