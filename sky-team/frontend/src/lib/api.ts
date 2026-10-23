// Thin client for the Go REST API. In dev, Vite proxies /api and /ws to the backend.
// Paths are prefixed with the build-time base path (BASE_PATH, e.g. /sky-team in production).
// The request helper and the lobby endpoints are shared (shared/frontend api.ts). Gameplay runs
// over the WebSocket (game/connection.svelte.ts).
import { createRequest, lobbyApi } from '@board-games/shared/api.js';
import type {
	ApiErrorCode,
	FlightOptions,
	GameSummary,
	PlayerView,
	SessionResponse
} from '#lib/types.js';

export { ApiError } from '@board-games/shared/api.js';

const request = createRequest<ApiErrorCode>(__BASE_PATH__);

export const { getHealth, joinGame, getSummary, getView, leaveGame } = lobbyApi<
	SessionResponse,
	GameSummary,
	PlayerView
>(request);

/** Creates a game; without a flight the server uses the defaults (YUL green, no abilities). */
export function createGame(name: string, flight?: FlightOptions): Promise<SessionResponse> {
	return request('POST', '/api/games', { body: { name, flight } });
}
