// Sample views for tests and UI previews. Not used by the running app.
import { MOVEMENT } from './format.js';
import type { Ability, Card, PlayerView, Space, Suit } from '#lib/types.js';

const ABILITY_BY_RANK: Record<number, Ability> = {
	1: 'musician',
	3: 'foxes',
	5: 'gazelle',
	7: 'gift',
	9: 'royalHeir'
};

export function card(suit: Suit, rank: number): Card {
	return { suit, rank, movement: MOVEMENT[rank] ?? 0, ability: ABILITY_BY_RANK[rank] ?? '' };
}

/** Side A at the start: gems per space and pluses. */
export function sideA(): Space[] {
	const gems = [1, 1, 2, 1, 1, 0, 1, 1, 2, 1, 1];
	return gems.map((g, i) => ({ gems: g, plus: i % 2 === 1, covered: false }));
}

/**
 * Round 1, trick 1 of a level-1 game: the viewer Alice (host, seat 0, left end) leads; Bob
 * dealt. Every call returns a fresh object.
 */
export function playingView(): PlayerView {
	const hand = [
		card('doves', 2),
		card('doves', 7),
		card('doves', 10),
		card('roses', 1),
		card('roses', 5),
		card('roses', 8),
		card('stars', 3),
		card('stars', 4),
		card('stars', 6),
		card('stars', 9),
		card('roses', 3)
	];
	return {
		gameId: 'K7MQ2',
		revision: 3,
		phase: 'playing',
		you: 'alice',
		hostId: 'alice',
		level: 1,
		players: [
			{
				id: 'alice',
				name: 'Alice',
				seat: 0,
				end: 'left',
				handCount: 11,
				isHost: true,
				isYou: true,
				isDealer: false,
				mustAct: true
			},
			{
				id: 'bob',
				name: 'Bob',
				seat: 1,
				end: 'right',
				handCount: 11,
				isHost: false,
				isYou: false,
				isDealer: true,
				mustAct: false
			}
		],
		hand,
		playable: [...hand],
		board: {
			side: 'A',
			start: 5,
			spaces: sideA(),
			traveller: 5,
			forestTokens: 4,
			forestTokensTotal: 4,
			gemsLeft: 12
		},
		round: 1,
		decree: card('doves', 6),
		trick: { number: 1, leaderId: 'alice', lead: null, follow: null },
		lastTrick: null,
		pending: {
			step: 'lead',
			playerIds: ['alice'],
			moveOptions: [],
			ends: [],
			yourGift: null
		},
		lastGift: null,
		result: null
	};
}
