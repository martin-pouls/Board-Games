import { createSessionStore } from '@board-games/shared/session.js';

export type { StoredSession, SessionStorageLike } from '@board-games/shared/session.js';

/** Sky Team's sessions in localStorage (`sky-team:session:<CODE>`, `sky-team:nickname`). */
export const sessions = createSessionStore('sky-team', () => globalThis.localStorage);
export default sessions;
