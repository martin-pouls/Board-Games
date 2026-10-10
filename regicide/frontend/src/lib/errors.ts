import type { ApiErrorCode } from '#lib/types.js';

const messages: Record<ApiErrorCode, string> = {
	bad_request: 'That request was not valid.',
	unauthorized: 'Your session has expired. Please rejoin the game.',
	not_host: 'Only the host can do that.',
	game_not_found: 'No game found with that code.',
	not_found: 'That could not be found.',
	wrong_step: "That's not possible at this step.",
	cannot_leave: "You're the only player; just head home.",
	illegal_play: "That isn't a legal play.",
	illegal_discard: "Those cards don't cover the damage, or include a card too many.",
	yield_not_allowed: "You can't yield: everyone else yielded on their last turn.",
	not_solo: 'Jester powers exist only in solo games.',
	no_solo_jester: 'No Jester power left.',
	illegal_signal: "You can't signal right now.",
	wrong_phase: 'That action is not possible right now.',
	game_full: 'This game is full.',
	name_taken: 'That name is already taken in this game.',
	not_your_turn: "It's not your turn.",
	conflict: 'Something changed at the same time. Please try again.',
	invalid_name: 'Please enter a valid name.',
	unknown_player: 'That player is not in this game.',
	card_not_in_hand: "That card isn't in your hand.",
	rematch_exists: 'A new game was already created.',
	rate_limited: 'Too many requests. Please wait a moment and try again.',
	internal: 'Something went wrong on the server.',
	network: "Can't reach the server.",
	disconnected: 'Connection lost. Reconnecting…',
	timeout: "The server didn't respond in time."
};

export function errorText(code: ApiErrorCode): string {
	return messages[code];
}
