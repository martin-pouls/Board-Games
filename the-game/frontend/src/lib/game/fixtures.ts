/**
 * Sample PlayerViews for every phase. Used by tests and handy for layout work
 * (render a component with one of these instead of a live connection).
 */
import type { Card, HandCard, Pile, PlayerSummary, PlayerView } from '#lib/types.js';

const card = (value: number): Card => ({ value, kind: 'number' });

function pile(index: number, top?: number, count = 0): Pile {
	const direction = index < 2 ? 'up' : 'down';
	const startValue = direction === 'up' ? 1 : 100;
	return {
		index,
		direction,
		startValue,
		topValue: top ?? startValue,
		topCard: top === undefined ? null : card(top),
		cardCount: top === undefined ? 0 : count
	};
}

/** Base-game placement (incl. backwards trick), good enough for fixtures. */
function legalPiles(value: number, piles: Pile[]): number[] {
	return piles
		.filter((p) =>
			p.direction === 'up'
				? value > p.topValue || value === p.topValue - 10
				: value < p.topValue || value === p.topValue + 10
		)
		.map((p) => p.index);
}

function hand(values: number[], piles: Pile[], interactive: boolean): HandCard[] {
	return [...values]
		.sort((a, b) => a - b)
		.map((v) => ({ ...card(v), legalPiles: interactive ? legalPiles(v, piles) : [] }));
}

const ME = 'p_cleo';

function players(
	opts: { active?: string; host?: string; counts?: number[]; reveal?: number[][] } = {}
): PlayerSummary[] {
	const names = ['Anna', 'Ben', 'Cleo', 'Dan'];
	return names.map((name, seat) => {
		const id = `p_${name.toLowerCase()}`;
		return {
			id,
			name,
			seat,
			handCount: opts.counts?.[seat] ?? 0,
			isHost: id === (opts.host ?? 'p_anna'),
			isYou: id === ME,
			isActive: id === opts.active,
			revealedHand: opts.reveal?.[seat]?.length ? opts.reveal[seat].map(card) : undefined
		};
	});
}

const base = {
	gameId: 'K7MQ2',
	you: ME,
	hostId: 'p_anna',
	options: { variant: 'classic', pro: false, reducedHandSize: false },
	rules: {
		minCardsPerTurn: 2,
		minCardsDrawPileEmpty: 1,
		handSize: 6,
		minPlayers: 1,
		maxPlayers: 5
	},
	communicationRestricted: false
} satisfies Partial<PlayerView>;

const freshPiles = [pile(0), pile(1), pile(2), pile(3)];

export const lobbyGuest: PlayerView = {
	...base,
	revision: 3,
	phase: 'lobby',
	players: players().slice(0, 3),
	hand: [],
	piles: freshPiles,
	drawPileCount: 98,
	turn: null,
	result: null
};

export const lobbyHost: PlayerView = {
	...lobbyGuest,
	hostId: ME,
	players: players({ host: ME }).slice(0, 3)
};

const startHand = [7, 15, 23, 41, 56, 62];

export const choosingGuest: PlayerView = {
	...base,
	revision: 5,
	phase: 'choosingFirstPlayer',
	players: players({ counts: [6, 6, 6, 6] }),
	hand: hand(startHand, freshPiles, false),
	piles: freshPiles,
	drawPileCount: 74,
	turn: null,
	result: null
};

export const choosingHost: PlayerView = {
	...choosingGuest,
	hostId: ME,
	players: players({ host: ME, counts: [6, 6, 6, 6] })
};

const midPiles = [pile(0, 34, 6), pile(1, 12, 2), pile(2, 71, 5), pile(3, 88, 3)];

export const playingMyTurn: PlayerView = {
	...base,
	revision: 42,
	phase: 'playing',
	players: players({ active: ME, counts: [6, 6, 5, 6] }),
	hand: hand([3, 24, 40, 61, 78], midPiles, true),
	piles: midPiles,
	drawPileCount: 64,
	turn: {
		number: 9,
		activePlayerId: ME,
		cardsPlayed: 1,
		requiredMinimum: 2,
		canEndTurn: false,
		plays: [{ card: card(34), pile: 0 }]
	},
	result: null
};

/** Solo-sized hand of 8 to check that the widest hand fits. */
export const playingEightCards: PlayerView = {
	...playingMyTurn,
	rules: { ...base.rules, handSize: 8 },
	players: players({ active: ME, counts: [8, 8, 8, 8] }),
	hand: hand([3, 22, 24, 40, 61, 78, 89, 98], midPiles, true),
	turn: { ...playingMyTurn.turn!, cardsPlayed: 0, plays: [] }
};

export const playingOthersTurn: PlayerView = {
	...playingMyTurn,
	players: players({ active: 'p_ben', counts: [6, 4, 6, 6] }),
	hand: hand([3, 24, 40, 61, 78, 95], midPiles, false),
	turn: {
		number: 10,
		activePlayerId: 'p_ben',
		cardsPlayed: 2,
		requiredMinimum: 2,
		canEndTurn: true,
		plays: [
			{ card: card(77), pile: 2 },
			{ card: card(13), pile: 1 }
		]
	}
};

const endPiles = [pile(0, 97, 30), pile(1, 99, 22), pile(2, 2, 25), pile(3, 5, 21)];

export const finishedWon: PlayerView = {
	...base,
	revision: 210,
	phase: 'finished',
	players: players({ counts: [0, 0, 0, 0] }),
	hand: [],
	piles: endPiles,
	drawPileCount: 0,
	turn: null,
	result: {
		outcome: 'won',
		score: 0,
		excellent: true,
		cardsLeftInHands: 0,
		cardsLeftInDrawPile: 0
	}
};

const lostPiles = [pile(0, 88, 20), pile(1, 93, 18), pile(2, 9, 17), pile(3, 14, 15)];

export const finishedLost: PlayerView = {
	...base,
	revision: 180,
	phase: 'finished',
	players: players({
		counts: [2, 4, 3, 0],
		reveal: [[45, 67], [30, 52, 60, 75], [19, 33, 48], []]
	}),
	hand: hand([19, 33, 48], lostPiles, false),
	piles: lostPiles,
	drawPileCount: 8,
	turn: null,
	result: {
		outcome: 'lost',
		score: 17,
		excellent: false,
		cardsLeftInHands: 9,
		cardsLeftInDrawPile: 8,
		stuckPlayerId: 'p_ben'
	},
	nextGameId: 'R3MAT'
};

export const fixtures = {
	lobbyGuest,
	lobbyHost,
	choosingGuest,
	choosingHost,
	playingMyTurn,
	playingEightCards,
	playingOthersTurn,
	finishedWon,
	finishedLost
} satisfies Record<string, PlayerView>;
