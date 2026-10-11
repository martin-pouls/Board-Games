import type { BaseView, ServerMessage } from './wire.js';

/** The join-code alphabet (no 0/O/1/I); mirrors shared/backend/idgen. */
export const CODE_ALPHABET = '23456789ABCDEFGHJKLMNPQRSTUVWXYZ';
export const CODE_LENGTH = 5;

/** Uppercases and strips characters outside the join-code alphabet. */
export function normalizeCode(input: string): string {
	return [...input.toUpperCase()]
		.filter((c) => CODE_ALPHABET.includes(c))
		.join('')
		.slice(0, CODE_LENGTH);
}

/** The game's WebSocket URL. basePath is the app's base path (`__BASE_PATH__`, '' in dev). */
export function wsUrl(
	loc: { protocol: string; host: string },
	basePath: string,
	code: string,
	token: string
): string {
	const scheme = loc.protocol === 'https:' ? 'wss:' : 'ws:';
	return `${scheme}//${loc.host}${basePath}/ws/games/${encodeURIComponent(code)}?token=${encodeURIComponent(token)}`;
}

/**
 * Parses a server message, checking the envelope only. The view and events are the game's and
 * are trusted to match its types (the server is authoritative). Garbage gives null.
 */
export function parseServerMessage<V, E, C extends string = string>(
	data: string
): ServerMessage<V, E, C> | null {
	let v: unknown;
	try {
		v = JSON.parse(data);
	} catch {
		return null;
	}
	if (!v || typeof v !== 'object') return null;
	const m = v as Record<string, unknown>;
	switch (m.type) {
		case 'view':
			if (!m.view || typeof m.view !== 'object') return null;
			return {
				type: 'view',
				view: m.view as V,
				events: Array.isArray(m.events) ? (m.events as E[]) : []
			};
		case 'presence':
			return {
				type: 'presence',
				online: Array.isArray(m.online)
					? m.online.filter((x): x is string => typeof x === 'string')
					: []
			};
		case 'ack':
			return typeof m.requestId === 'string' ? { type: 'ack', requestId: m.requestId } : null;
		case 'error': {
			const e = m.error as { code?: unknown; message?: unknown } | undefined;
			if (!e || typeof e.code !== 'string') return null;
			return {
				type: 'error',
				requestId: typeof m.requestId === 'string' ? m.requestId : undefined,
				error: { code: e.code as C, message: String(e.message ?? '') }
			};
		}
		default:
			return null;
	}
}

/** 500 ms * 2^attempt, capped at 10 s, with +-20 % jitter. */
export function nextBackoff(attempt: number, rand: () => number = Math.random): number {
	const base = Math.min(500 * 2 ** Math.max(0, attempt), 10_000);
	return base * (0.8 + 0.4 * rand());
}

/** Views can arrive out of order (broadcast vs. reconnect): never go back to an older revision. */
export function shouldAcceptView(current: BaseView | null, next: BaseView): boolean {
	return current === null || next.revision >= current.revision;
}
