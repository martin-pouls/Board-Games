// Thin client for the Go REST API (lobby only; gameplay goes over the WebSocket).
// In dev, Vite proxies /api and /ws to the backend.
// Paths are prefixed with the build-time base path (BASE_PATH, e.g. /the-game in production).
import { createRequest, lobbyApi } from '@board-games/shared/api.js';
import type {
	ApiErrorCode,
	GameOptions,
	GameSummary,
	PlayerView,
	SessionResponse
} from '#lib/types.js';

export { ApiError } from '@board-games/shared/api.js';

const request = createRequest<ApiErrorCode>(__BASE_PATH__);

export const { getHealth, joinGame, getSummary, getView } = lobbyApi<
	SessionResponse,
	GameSummary,
	PlayerView
>(request);

export function createGame(name: string, options?: GameOptions): Promise<SessionResponse> {
	return request('POST', '/api/games', { body: { name, options } });
}
