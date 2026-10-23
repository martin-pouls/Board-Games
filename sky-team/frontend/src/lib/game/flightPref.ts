/** The host's last scenario and abilities, kept per device (`sky-team:flight`) for new games. */
import { createPref, isRecord } from '@board-games/shared/prefs.js';
import type { AbilityId, FlightOptions } from '#lib/types.js';

export const ABILITY_IDS: AbilityId[] = [
	'adaptation',
	'anticipation',
	'control',
	'mastery',
	'synchronisation',
	'workingTogether'
];

/** A stored value in today's shape; null when it isn't one. Unknown abilities are dropped. */
export function parseFlight(v: unknown): FlightOptions | null {
	if (!isRecord(v) || typeof v.scenario !== 'string' || !Array.isArray(v.abilities)) return null;
	const abilities = ABILITY_IDS.filter((id) => (v.abilities as unknown[]).includes(id));
	return { scenario: v.scenario, abilities };
}

export const flightPref = createPref('sky-team:flight', parseFlight);
