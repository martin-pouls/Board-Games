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
	'not_your_turn',
	'conflict',
	'invalid_name',
	'invalid_options',
	'unknown_player',
	'card_not_in_hand',
	'invalid_pile',
	'illegal_placement',
	'minimum_not_reached',
	'rematch_exists',
	'backwards_trick_blocked',
	'single_pile_violation',
	'exact_count_exceeded',
	'exact_count_unreachable',
	'exact_count_required',
	'stop_too_early',
	'skull_uncovered',
	'skull_uncoverable',
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
