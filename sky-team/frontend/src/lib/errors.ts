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
	unknown_player: 'That player is not in this game.'
};

export function errorText(code: ApiErrorCode): string {
	return messages[code];
}
