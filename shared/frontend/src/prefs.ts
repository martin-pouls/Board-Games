// One remembered per-device preference (e.g. the host's last lobby settings), kept as JSON in
// localStorage. Unavailable storage, unparsable or outdated values all read as "nothing stored".

export type PrefStorageLike = Pick<Storage, 'getItem' | 'setItem'>;

/**
 * `key` should carry the game prefix (`<game>:<name>`). `parse` checks a decoded value and returns
 * it in its current shape, or null to drop it. `storage` is a getter so that unavailable storage
 * is only discovered lazily.
 */
export function createPref<T>(
	key: string,
	parse: (raw: unknown) => T | null,
	storage: () => PrefStorageLike | undefined = () => globalThis.localStorage
) {
	return {
		/** The stored value, or null when none (valid) is stored or storage is blocked. */
		load(): T | null {
			try {
				const raw = storage()?.getItem(key);
				return raw == null ? null : parse(JSON.parse(raw));
			} catch {
				return null;
			}
		},
		/** Remembers the value; silently does nothing when storage is blocked. */
		save(value: T): void {
			try {
				storage()?.setItem(key, JSON.stringify(value));
			} catch {
				// Private window or blocked storage: the value is simply not remembered.
			}
		}
	};
}

export type Pref<T> = ReturnType<typeof createPref<T>>;

/** A plain object, for `parse` functions. */
export const isRecord = (v: unknown): v is Record<string, unknown> =>
	typeof v === 'object' && v !== null && !Array.isArray(v);
