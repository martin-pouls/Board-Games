import { createSessionStore } from '@board-games/shared/session.js';
import { storagePrefix } from './edition.js';

export type { StoredSession, SessionStorageLike } from '@board-games/shared/session.js';

/** The sessions in localStorage (`<storagePrefix>:session:<CODE>`, `<storagePrefix>:nickname`). */
export const sessions = createSessionStore(storagePrefix, () => globalThis.localStorage);
export default sessions;
