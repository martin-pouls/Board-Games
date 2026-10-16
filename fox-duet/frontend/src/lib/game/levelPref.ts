/** The host's last difficulty level, kept per device (`fox-duet:level`) for the next lobby. */
import { createPref } from '@board-games/shared/prefs.js';

export const LEVEL_NUMBERS = [1, 2, 3] as const;

/** A stored level; null when it isn't one of today's levels. */
export const parseLevel = (v: unknown): number | null => LEVEL_NUMBERS.find((l) => l === v) ?? null;

export const levelPref = createPref('fox-duet:level', parseLevel);
