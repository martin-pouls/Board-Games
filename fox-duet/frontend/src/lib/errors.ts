import { commonErrorMessages } from '@board-games/shared/errors.js';
import type { ApiErrorCode } from '#lib/types.js';

const messages: Record<ApiErrorCode, string> = {
	...commonErrorMessages,
	not_host: 'Only the host can do that.',
	wrong_phase: 'That action is not possible right now.',
	wrong_step: "That's not possible at this step.",
	game_full: 'This game already has two players.',
	name_taken: 'That name is already taken in this game.',
	not_your_turn: "It's not your turn.",
	cannot_leave: "You're the only player; just head home.",
	invalid_name: 'Please enter a valid name.',
	unknown_player: 'That player is not in this game.',
	not_enough_players: 'The game needs two players.',
	invalid_level: 'Pick level 1, 2 or 3.',
	card_not_in_hand: "That card isn't in your hand.",
	must_follow_suit: 'You must follow the lead suit.',
	gift_already_chosen: 'You already chose your gift.',
	illegal_move: "The traveller can't move that way.",
	illegal_end: "The forest token can't go there.",
	rematch_exists: 'A new game was already created.'
};

export function errorText(code: ApiErrorCode): string {
	return messages[code];
}
