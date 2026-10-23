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
	'unknown_scenario',
	'invalid_abilities',
	'not_enough_players',
	'abilities_incomplete',
	'wrong_step',
	'already_ready',
	'not_your_turn',
	'no_such_die',
	'not_enough_coffee',
	'value_out_of_range',
	'illegal_space',
	'illegal_token_space',
	'no_reroll',
	'already_confirmed',
	'ability_not_in_play',
	'ability_used',
	'partner_has_no_dice',
	'rematch_exists',
	'rate_limited',
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
