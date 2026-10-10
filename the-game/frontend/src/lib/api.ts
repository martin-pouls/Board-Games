// Thin client for the Go REST API. In dev, Vite proxies /api and /ws to the backend.
import type {
	ApiErrorCode,
	ErrorBody,
	GameOptions,
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
	const res = await fetch('/api/health');
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
		res = await fetch(path, { method, headers, body });
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
	try {
		return (await res.json()) as T;
	} catch {
		throw new ApiError(res.status, 'internal', 'invalid response');
	}
}

const game = (code: string) => `/api/games/${encodeURIComponent(code.trim().toUpperCase())}`;

export function createGame(name: string, options?: GameOptions): Promise<SessionResponse> {
	return request('POST', '/api/games', { body: { name, options } });
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
export function endTurn(code: string, token: string): Promise<PlayerView> {
	return request('POST', `${game(code)}/end-turn`, { token });
}
export function endGame(code: string, token: string): Promise<PlayerView> {
	return request('POST', `${game(code)}/end-game`, { token });
}

export function rematch(code: string, token: string): Promise<PlayerView> {
	return request('POST', `${game(code)}/rematch`, { token });
}
