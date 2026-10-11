import { createSessionStore } from '@board-games/shared/session.js';

export type { StoredSession, SessionStorageLike } from '@board-games/shared/session.js';

/** Regicide's sessions in localStorage (`regicide:session:<CODE>`, `regicide:nickname`). */
export const sessions = createSessionStore('regicide', () => globalThis.localStorage);
export default sessions;
