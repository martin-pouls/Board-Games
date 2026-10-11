import { createSessionStore } from '@board-games/shared/session.js';

export type { StoredSession, SessionStorageLike } from '@board-games/shared/session.js';

/** The Crew's sessions in localStorage (`the-crew:session:<CODE>`, `the-crew:nickname`). */
export const sessions = createSessionStore('the-crew', () => globalThis.localStorage);
export default sessions;
