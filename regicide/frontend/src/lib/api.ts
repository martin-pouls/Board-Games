// Thin client for the Go REST API: the lobby only. Gameplay goes over the WebSocket (src/lib/game).
// In dev, Vite proxies /api and /ws to the backend.
// Paths are prefixed with the build-time base path (BASE_PATH, e.g. /regicide in production).
// The request helper and the lobby endpoints are shared (shared/frontend api.ts).
import { createRequest, lobbyApi } from '@board-games/shared/api.js';
import type { ApiErrorCode, GameSummary, PlayerView, SessionResponse } from '#lib/types.js';

export { ApiError } from '@board-games/shared/api.js';

const request = createRequest<ApiErrorCode>(__BASE_PATH__);

export const { getHealth, createGame, joinGame, getSummary, getView, leaveGame } = lobbyApi<
	SessionResponse,
	GameSummary,
	PlayerView
>(request);
