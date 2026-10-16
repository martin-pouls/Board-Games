/** The host's last lobby options, kept per device (`the-game:lobby-options`) for new games. */
import { createPref, isRecord } from '@board-games/shared/prefs.js';
import type { GameOptions } from '#lib/types.js';

/** A stored value in today's shape; null when it isn't one. Reduced needs Pro (as on the server). */
export function parseLobbyOptions(v: unknown): GameOptions | null {
	if (!isRecord(v)) return null;
	const { variant, pro, reducedHandSize } = v;
	if (variant !== 'classic' && variant !== 'extreme') return null;
	if (typeof pro !== 'boolean' || typeof reducedHandSize !== 'boolean') return null;
	return { variant, pro, reducedHandSize: pro && reducedHandSize };
}

export const lobbyOptions = createPref('the-game:lobby-options', parseLobbyOptions);
