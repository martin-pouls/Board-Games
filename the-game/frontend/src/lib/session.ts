export interface StoredSession {
	token: string;
	playerId: string;
	name: string;
}
export type SessionStorageLike = Pick<Storage, 'getItem' | 'setItem' | 'removeItem'> &
	Partial<Pick<Storage, 'key' | 'length'>>;

const SESSION_PREFIX = 'the-game:session:';
const NICKNAME_KEY = 'the-game:nickname';

const norm = (code: string) => code.trim().toUpperCase();

/** `storage` may be a getter so that unavailable storage is only discovered lazily. */
export function createSessionStore(
	storage: SessionStorageLike | (() => SessionStorageLike | undefined)
) {
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
			const raw = s.getItem(SESSION_PREFIX + norm(code));
			if (!raw) return null;
			const v = JSON.parse(raw);
			if (v && typeof v.token === 'string' && typeof v.playerId === 'string') {
				return { token: v.token, playerId: v.playerId, name: String(v.name ?? '') };
			}
			return null;
		}, null);
	const save = (code: string, session: StoredSession) =>
		safe((s) => s.setItem(SESSION_PREFIX + norm(code), JSON.stringify(session)), undefined);
	return {
		get,
		save,
		clear: (code: string) => safe((s) => s.removeItem(SESSION_PREFIX + norm(code)), undefined),
		list: (): { code: string; session: StoredSession }[] =>
			safe((s) => {
				const codes: string[] = [];
				for (let i = 0; i < (s.length ?? 0); i++) {
					const k = s.key?.(i);
					if (k?.startsWith(SESSION_PREFIX)) codes.push(k.slice(SESSION_PREFIX.length));
				}
				const out: { code: string; session: StoredSession }[] = [];
				for (const code of codes) {
					const session = get(code);
					if (session) out.push({ code, session });
				}
				return out;
			}, []),
		getNickname: (): string => safe((s) => s.getItem(NICKNAME_KEY) ?? '', ''),
		setNickname: (name: string) => safe((s) => s.setItem(NICKNAME_KEY, name), undefined),
		copy: (fromCode: string, toCode: string) => {
			const v = get(fromCode);
			if (v) save(toCode, v);
		}
	};
}

export const sessions = createSessionStore(() => globalThis.localStorage);
export default sessions;
