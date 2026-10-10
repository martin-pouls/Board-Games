import type { ApiErrorCode } from '#lib/types.js';

const messages: Record<ApiErrorCode, string> = {
	bad_request: 'That request was not valid.',
	unauthorized: 'Your session has expired. Please rejoin the game.',
	not_host: 'Only the host can do that.',
	game_not_found: 'No game found with that code.',
	not_found: 'That could not be found.',
	wrong_phase: 'That action is not possible right now.',
	game_full: 'This game is full.',
	name_taken: 'That name is already taken in this game.',
	not_your_turn: "It's not your turn.",
	conflict: 'Something changed at the same time. Please try again.',
	invalid_name: 'Please enter a valid name.',
	invalid_options: 'Those game options are not valid.',
	unknown_player: 'That player is not in this game.',
	card_not_in_hand: "That card isn't in your hand.",
	invalid_pile: "That pile doesn't exist.",
	illegal_placement: "That card can't go on this pile.",
	minimum_not_reached: 'Play the minimum number of cards first.',
	rematch_exists: 'A new game was already created.',
	nothing_to_undo: 'No card to undo this turn.',
	not_stuck: 'You can still play a card or end your turn.',
	stop_played: 'STOP: no more cards this turn. End your turn.',
	backwards_trick_blocked: 'No backwards trick while that card is visible.',
	single_pile_violation: 'Single pile: all your cards this turn go on the same pile.',
	exact_count_exceeded: '3!: exactly 3 cards this turn, no more.',
	exact_count_unreachable: "You can't play 3!: you started this turn with fewer than 3 cards.",
	exact_count_required: '3!: play exactly 3 cards before ending your turn.',
	stop_too_early: 'After 3!, STOP can only be your third card.',
	skull_uncovered: "Cover the skull first: it can't stay on top when your turn ends.",
	skull_uncoverable: 'After 3! no 4th card is allowed, so the skull could never be covered.',
	rate_limited: 'Too many requests. Please wait a moment and try again.',
	internal: 'Something went wrong on the server.',
	network: "Can't reach the server.",
	disconnected: 'Connection lost. Reconnecting…',
	timeout: "The server didn't respond in time."
};

export function errorText(code: ApiErrorCode): string {
	return messages[code];
}
