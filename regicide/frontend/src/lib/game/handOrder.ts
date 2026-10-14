/** The player's hand order, kept per device in localStorage (`regicide:hand-order`). */
import { DEFAULT_HAND_ORDER, HAND_ORDERS, type HandOrder } from './rules.js';

export const HAND_ORDER_KEY = 'regicide:hand-order';

type StorageLike = Pick<Storage, 'getItem' | 'setItem'>;
const defaultStorage = (): StorageLike | undefined => globalThis.localStorage;

/** The stored order; the default when none is stored, it is unknown or storage is blocked. */
export function loadHandOrder(storage: () => StorageLike | undefined = defaultStorage): HandOrder {
	try {
		const v = storage()?.getItem(HAND_ORDER_KEY);
		return HAND_ORDERS.find((o) => o === v) ?? DEFAULT_HAND_ORDER;
	} catch {
		return DEFAULT_HAND_ORDER;
	}
}

/** Remembers the order; silently does nothing when storage is blocked. */
export function saveHandOrder(
	order: HandOrder,
	storage: () => StorageLike | undefined = defaultStorage
): void {
	try {
		storage()?.setItem(HAND_ORDER_KEY, order);
	} catch {
		// Private window or blocked storage: the order then lasts until the page reloads.
	}
}
