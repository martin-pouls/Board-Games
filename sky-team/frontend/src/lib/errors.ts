import { commonErrorMessages } from '@board-games/shared/errors.js';
import type { ApiErrorCode } from '#lib/types.js';

const messages: Record<ApiErrorCode, string> = {
	...commonErrorMessages,
	not_host: 'Only the host can do that.',
	wrong_phase: 'That action is not possible right now.',
	game_full: 'This game already has two players.',
	name_taken: 'That name is already taken in this game.',
	cannot_leave: "You're the only player; just head home.",
	invalid_name: 'Please enter a valid name.',
	unknown_player: 'That player is not in this game.',
	unknown_scenario: 'That scenario does not exist.',
	invalid_abilities: 'That ability selection is not allowed for this scenario.',
	not_enough_players: 'The game needs two players.',
	abilities_incomplete: 'Pick all the ability cards the scenario calls for.',
	wrong_step: "That's not possible at this step.",
	already_ready: 'You are already ready.',
	not_your_turn: "It's not your turn.",
	no_such_die: "That die isn't in your hand.",
	not_enough_coffee: 'Not enough coffee on the panel.',
	value_out_of_range: 'Coffee cannot take a die below 1 or above 6.',
	illegal_space: "That die can't go there.",
	illegal_token_space: "The intern token can't go there.",
	no_reroll: 'No reroll token is left.',
	already_confirmed: 'You already confirmed.',
	ability_not_in_play: 'That ability is not in this game.',
	ability_used: 'That ability has already been used.',
	partner_has_no_dice: 'Your partner has no dice left to swap.',
	rematch_exists: 'A new game was already created.'
};

export function errorText(code: ApiErrorCode): string {
	return messages[code];
}
