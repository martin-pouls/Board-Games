// Thin client for the Go REST API. In dev, Vite proxies /api and /ws to the backend.
// Paths are prefixed with the build-time base path (BASE_PATH, e.g. /regicide in production).
import type {
	ApiErrorCode,
	ErrorBody,
	Card,
	JesterSignal,
	GameSummary,
	PlayerView,
	SessionResponse
} from '#lib/types.js';

export class ApiError extends Error {
	status: number;
	code: ApiErrorCode;
	constructor(status: number, code: ApiErrorCode, message: string) {
		super(message);
		this.name = 'ApiError';
		this.status = status;
		this.code = code;
	}
}

export async function getHealth(): Promise<{ status: string }> {
	const res = await fetch(`${__BASE_PATH__}/api/health`);
	if (!res.ok) throw new Error(`health check failed: ${res.status}`);
	return res.json();
}

async function request<T>(
	method: string,
	path: string,
	opts: { token?: string; body?: unknown } = {}
): Promise<T> {
	const headers: Record<string, string> = {};
	if (opts.token) headers['Authorization'] = `Bearer ${opts.token}`;
	let body: string | undefined;
	if (opts.body !== undefined) {
		headers['Content-Type'] = 'application/json';
		body = JSON.stringify(opts.body);
	}
	let res: Response;
	try {
		res = await fetch(__BASE_PATH__ + path, { method, headers, body });
	} catch {
		throw new ApiError(0, 'network', 'network error');
	}
	if (!res.ok) {
		let err: Partial<ErrorBody> = {};
		try {
			err = (await res.json()).error ?? {};
		} catch {
			// non-JSON error body
		}
		throw new ApiError(res.status, err.code ?? 'internal', err.message ?? res.statusText);
	}
	if (res.status === 204) return undefined as T;
	try {
		return (await res.json()) as T;
	} catch {
		throw new ApiError(res.status, 'internal', 'invalid response');
	}
}

const game = (code: string) => `/api/games/${encodeURIComponent(code.trim().toUpperCase())}`;

export function createGame(name: string): Promise<SessionResponse> {
	return request('POST', '/api/games', { body: { name } });
}
export function joinGame(code: string, name: string): Promise<SessionResponse> {
	return request('POST', `${game(code)}/players`, { body: { name } });
}
export function getSummary(code: string): Promise<GameSummary> {
	return request('GET', game(code));
}
export function getView(code: string, token: string): Promise<PlayerView> {
	return request('GET', `${game(code)}/view`, { token });
}
export function leaveGame(code: string, token: string): Promise<void> {
	return request('POST', `${game(code)}/leave`, { token });
}
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
export function giveUp(code: string, token: string): Promise<PlayerView> {
	return request('POST', `${game(code)}/give-up`, { token });
}
export function rematch(code: string, token: string): Promise<PlayerView> {
	return request('POST', `${game(code)}/rematch`, { token });
}
