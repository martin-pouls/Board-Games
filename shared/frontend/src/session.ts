// The player's sessions (token per game code) and last nickname, kept in localStorage under a
// per-game prefix so the games on the shared production origin never read each other's sessions.

export interface StoredSession {
	token: string;
	playerId: string;
	name: string;
}
export type SessionStorageLike = Pick<Storage, 'getItem' | 'setItem' | 'removeItem'> &
	Partial<Pick<Storage, 'key' | 'length'>>;

const norm = (code: string) => code.trim().toUpperCase();

/**
 * `game` namespaces the keys (`<game>:session:<CODE>`, `<game>:nickname`). `storage` may be a
 * getter so that unavailable storage is only discovered lazily; every access is wrapped, so a
 * private window or blocked storage just behaves as empty.
 */
export function createSessionStore(
	game: string,
	storage: SessionStorageLike | (() => SessionStorageLike | undefined)
) {
	const sessionPrefix = `${game}:session:`;
	const nicknameKey = `${game}:nickname`;
	const resolve = typeof storage === 'function' ? storage : () => storage;
	const safe = <T>(fn: (s: SessionStorageLike) => T, fallback: T): T => {
		try {
			const s = resolve();
			return s ? fn(s) : fallback;
		} catch {
			return fallback;
		}
	};
	const get = (code: string): StoredSession | null =>
		safe((s) => {
			const raw = s.getItem(sessionPrefix + norm(code));
			if (!raw) return null;
			const v = JSON.parse(raw);
			if (v && typeof v.token === 'string' && typeof v.playerId === 'string') {
				return { token: v.token, playerId: v.playerId, name: String(v.name ?? '') };
			}
			return null;
		}, null);
	const save = (code: string, session: StoredSession) =>
		safe((s) => s.setItem(sessionPrefix + norm(code), JSON.stringify(session)), undefined);
	return {
		get,
		save,
		clear: (code: string) => safe((s) => s.removeItem(sessionPrefix + norm(code)), undefined),
		list: (): { code: string; session: StoredSession }[] =>
			safe((s) => {
				const codes: string[] = [];
				for (let i = 0; i < (s.length ?? 0); i++) {
					const k = s.key?.(i);
					if (k?.startsWith(sessionPrefix)) codes.push(k.slice(sessionPrefix.length));
				}
				const out: { code: string; session: StoredSession }[] = [];
				for (const code of codes) {
					const session = get(code);
					if (session) out.push({ code, session });
				}
				return out;
			}, []),
		getNickname: (): string => safe((s) => s.getItem(nicknameKey) ?? '', ''),
		setNickname: (name: string) => safe((s) => s.setItem(nicknameKey, name), undefined),
		copy: (fromCode: string, toCode: string) => {
			const v = get(fromCode);
			if (v) save(toCode, v);
		}
	};
}

export type SessionStore = ReturnType<typeof createSessionStore>;
