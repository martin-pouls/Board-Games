import { createSessionStore } from '@board-games/shared/session.js';

export type { StoredSession, SessionStorageLike } from '@board-games/shared/session.js';

/** Fox Duet's sessions in localStorage (`fox-duet:session:<CODE>`, `fox-duet:nickname`). */
export const sessions = createSessionStore('fox-duet', () => globalThis.localStorage);
export default sessions;
