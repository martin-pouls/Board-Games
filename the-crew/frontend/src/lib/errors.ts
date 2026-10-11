import { commonErrorMessages } from '@board-games/shared/errors.js';
import type { ApiErrorCode } from '#lib/types.js';

const messages: Record<ApiErrorCode, string> = {
	...commonErrorMessages,
	not_host: 'Only the host can do that.',
	wrong_phase: 'That action is not possible right now.',
	game_full: 'This game already has five players.',
	name_taken: 'That name is already taken in this game.',
	cannot_leave: "You're the only player; just head home.",
	invalid_name: 'Please enter a valid name.',
	unknown_player: 'That player is not in this game.',
	not_enough_players: 'The crew needs at least two players.',
	not_commander: 'Only the commander can do that.',
	wrong_step: 'That is not possible at this point of the mission.',
	duplicate_player: "You're already in this game.",
	not_your_turn: "It's not your turn.",
	invalid_mission: 'There is no such mission.',
	card_not_in_hand: "That card isn't in the hand.",
	must_follow_suit: 'You must follow the colour that was led.',
	cannot_communicate: "You can't communicate right now.",
	already_communicated: 'You have already communicated in this attempt.',
	invalid_position: "That statement isn't true for this card.",
	invalid_seat: "That crew member can't be chosen.",
	invalid_task: "That task can't be chosen.",
	uneven_distribution: 'The tasks must be shared out as evenly as possible.',
	cannot_pass_trump: "Rockets can't be passed.",
	already_answered: 'You have already answered.',
	not_allowed: "That isn't allowed in this mission.",
	invalid_direction: 'Choose left or right.',
	invalid_token_move: "That token can't be moved there.",
	invalid_result_choice: "That isn't possible after this result."
};

export function errorText(code: ApiErrorCode): string {
	return messages[code];
}
