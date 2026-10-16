/** The host's last lobby options, kept per device (`the-crew:lobby-options`) for new games. */
import { createPref, isRecord } from '@board-games/shared/prefs.js';
import type { GameOptions } from '#lib/types.js';

/** A stored value in today's shape; null when it isn't one. */
export function parseLobbyOptions(v: unknown): GameOptions | null {
	if (!isRecord(v) || typeof v.distress !== 'boolean') return null;
	return { distress: v.distress };
}

export const lobbyOptions = createPref('the-crew:lobby-options', parseLobbyOptions);
