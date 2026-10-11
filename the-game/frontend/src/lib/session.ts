import { createSessionStore } from '@board-games/shared/session.js';

export type { StoredSession, SessionStorageLike } from '@board-games/shared/session.js';

/** The Game's sessions in localStorage (`the-game:session:<CODE>`, `the-game:nickname`). */
export const sessions = createSessionStore('the-game', () => globalThis.localStorage);
export default sessions;
