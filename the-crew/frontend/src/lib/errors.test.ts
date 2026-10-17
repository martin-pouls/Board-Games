import { describe, expect, it } from 'vitest';
import { errorText } from '#lib/errors.js';
import type { ApiErrorCode } from '#lib/types.js';

const codes: ApiErrorCode[] = [
	'bad_request',
	'unauthorized',
	'not_host',
	'game_not_found',
	'not_found',
	'wrong_phase',
	'game_full',
	'name_taken',
	'cannot_leave',
	'conflict',
	'invalid_name',
	'unknown_player',
	'not_enough_players',
	'rate_limited',
	'internal',
	'network',
	'disconnected',
	'timeout',
	'invalid_prediction',
	'uneven_distribution',
	'invalid_task',
	'invalid_seat'
];

describe('errorText', () => {
	it.each(codes)('returns a non-empty message for %s', (code) => {
		expect(errorText(code).length).toBeGreaterThan(0);
	});
});
