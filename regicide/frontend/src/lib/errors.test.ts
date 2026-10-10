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
	'wrong_step',
	'game_full',
	'name_taken',
	'not_your_turn',
	'cannot_leave',
	'conflict',
	'invalid_name',
	'unknown_player',
	'card_not_in_hand',
	'illegal_play',
	'illegal_discard',
	'yield_not_allowed',
	'not_solo',
	'no_solo_jester',
	'illegal_signal',
	'rematch_exists',
	'internal',
	'network',
	'disconnected',
	'timeout'
];

describe('errorText', () => {
	it.each(codes)('returns a non-empty message for %s', (code) => {
		expect(errorText(code).length).toBeGreaterThan(0);
	});
});
