import type { GameEvent, PlayerView, ServerMessage } from '#lib/types.js';

export const CODE_ALPHABET = '23456789ABCDEFGHJKLMNPQRSTUVWXYZ';
export const CODE_LENGTH = 5;

/** Uppercases and strips characters outside the join-code alphabet. */
export function normalizeCode(input: string): string {
	return [...input.toUpperCase()]
		.filter((c) => CODE_ALPHABET.includes(c))
		.join('')
		.slice(0, CODE_LENGTH);
}

export function wsUrl(
	loc: { protocol: string; host: string },
	code: string,
	token: string
): string {
	const scheme = loc.protocol === 'https:' ? 'wss:' : 'ws:';
	return `${scheme}//${loc.host}${__BASE_PATH__}/ws/games/${encodeURIComponent(code)}?token=${encodeURIComponent(token)}`;
}

export function parseServerMessage(data: string): ServerMessage | null {
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
				view: m.view as PlayerView,
				events: Array.isArray(m.events) ? (m.events as GameEvent[]) : []
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
				error: {
					code: e.code as ServerMessageErrorCode,
					message: String(e.message ?? '')
				}
			};
		}
		default:
			return null;
	}
}

type ServerMessageErrorCode = Extract<ServerMessage, { type: 'error' }>['error']['code'];

/** 500 ms * 2^attempt, capped at 10 s, with +-20 % jitter. */
export function nextBackoff(attempt: number, rand: () => number = Math.random): number {
	const base = Math.min(500 * 2 ** Math.max(0, attempt), 10_000);
	return base * (0.8 + 0.4 * rand());
}

export function shouldAcceptView(current: PlayerView | null, next: PlayerView): boolean {
	return current === null || next.revision >= current.revision;
}
