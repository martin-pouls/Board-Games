// REST plumbing every game shares: the request helper with bearer auth and wire errors, and the
// lobby endpoints (same routes in every game). Games add their own endpoints on top of request.
import type { ClientErrorCode, CommonErrorCode, ErrorBody } from './wire.js';

export class ApiError<C extends string = string> extends Error {
	status: number;
	code: C;
	constructor(status: number, code: C, message: string) {
		super(message);
		this.name = 'ApiError';
		this.status = status;
		this.code = code;
	}
}

export type RequestOptions = { token?: string; body?: unknown };
export type Request = <T>(method: string, path: string, opts?: RequestOptions) => Promise<T>;

/**
 * Returns the request function for an app served under basePath (`__BASE_PATH__`, '' in dev;
 * Vite proxies /api in dev). Failures throw ApiError with the server's code, 'network' when the
 * server can't be reached and 'internal' for unreadable responses. C is the game's error codes.
 */
export function createRequest<C extends string>(
	basePath: string,
	fetchFn: typeof fetch = (...args) => fetch(...args)
): Request {
	return async function request<T>(
		method: string,
		path: string,
		opts: RequestOptions = {}
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
			res = await fetchFn(basePath + path, { method, headers, body });
		} catch {
			throw new ApiError<C | ClientErrorCode>(0, 'network', 'network error');
		}
		if (!res.ok) {
			let err: Partial<ErrorBody<C>> = {};
			try {
				err = (await res.json()).error ?? {};
			} catch {
				// non-JSON error body
			}
			throw new ApiError<C | CommonErrorCode>(
				res.status,
				err.code ?? 'internal',
				err.message ?? res.statusText
			);
		}
		if (res.status === 204) return undefined as T;
		try {
			return (await res.json()) as T;
		} catch {
			throw new ApiError<C | CommonErrorCode>(res.status, 'internal', 'invalid response');
		}
	};
}

/** The path of a game's REST resource, from a (possibly untidy) join code. */
export const gamePath = (code: string) =>
	`/api/games/${encodeURIComponent(code.trim().toUpperCase())}`;

/**
 * The lobby endpoints every game has. Session, Summary and View are the game's response types.
 */
export function lobbyApi<Session, Summary, View>(request: Request) {
	return {
		getHealth: () => request<{ status: string }>('GET', '/api/health'),
		createGame: (name: string) => request<Session>('POST', '/api/games', { body: { name } }),
		joinGame: (code: string, name: string) =>
			request<Session>('POST', `${gamePath(code)}/players`, { body: { name } }),
		getSummary: (code: string) => request<Summary>('GET', gamePath(code)),
		getView: (code: string, token: string) =>
			request<View>('GET', `${gamePath(code)}/view`, { token }),
		leaveGame: (code: string, token: string) =>
			request<void>('POST', `${gamePath(code)}/leave`, { token }),
		giveUp: (code: string, token: string) =>
			request<View>('POST', `${gamePath(code)}/give-up`, { token }),
		rematch: (code: string, token: string) =>
			request<View>('POST', `${gamePath(code)}/rematch`, { token })
	};
}
