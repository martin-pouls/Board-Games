import type { ClientErrorCode, CommonErrorCode } from './wire.js';

/** Messages for the error codes every game has. Games spread these into their own table. */
export const commonErrorMessages: Record<CommonErrorCode | ClientErrorCode, string> = {
	bad_request: 'That request was not valid.',
	unauthorized: 'Your session has expired. Please rejoin the game.',
	game_not_found: 'No game found with that code.',
	not_found: 'That could not be found.',
	conflict: 'Something changed at the same time. Please try again.',
	rate_limited: 'Too many requests. Please wait a moment and try again.',
	internal: 'Something went wrong on the server.',
	login_required: 'Your login has expired. Reloading…',
	network: "Can't reach the server.",
	disconnected: 'Connection lost. Reconnecting…',
	timeout: "The server didn't respond in time."
};
