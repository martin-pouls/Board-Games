/**
 * Sample PlayerViews for every phase. Used by tests and handy for layout work
 * (render a component with one of these instead of a live connection).
 */
import type {
	ApiErrorCode,
	Card,
	CommandName,
	HandCard,
	Pile,
	PlayerSummary,
	PlayerView,
	Turn,
	TurnConstraints
} from '#lib/types.js';

export const card = (value: number, command?: CommandName): Card =>
	command ? { value, kind: 'number', command } : { value, kind: 'number' };

/** Extreme deck: the card values that carry a command (docs/rules.md §2). */
const extremeValues: Record<CommandName, number[]> = {
	stop: [6, 35, 57, 95],
	skull: [17, 41, 61, 87],
	exactlyThree: [19, 28, 53, 81],
	noCommunication: [26, 65, 78, 92],
	noBackwardsTrick: [21, 32, 59, 70],
	singlePile: [9, 47, 68, 84],
	drawOnlyOne: [13, 38, 43, 76]
};

/** The card with this value in an Extreme deck (with its command, if any). */
function xcard(value: number): Card {
	const entry = Object.entries(extremeValues).find(([, values]) => values.includes(value));
	return card(value, entry?.[0] as CommandName | undefined);
}

/** Turn constraints when no command applies. */
export const noConstraints: TurnConstraints = {
	exactCount: null,
	singlePile: false,
	boundPile: null,
	skullPiles: [],
	stopPlayed: false
};

function pile(
	index: number,
	top?: number,
	count = 0,
	opts: { extreme?: boolean; commandActive?: boolean } = {}
): Pile {
	const direction = index < 2 ? 'up' : 'down';
	const startValue = direction === 'up' ? 1 : 100;
	const topCard = top === undefined ? null : opts.extreme ? xcard(top) : card(top);
	const cards: Card[] = [];
	if (top !== undefined && topCard) {
		// Plausible history: evenly spaced values from the start card towards the top.
		for (let i = 1; i < count; i++) {
			const v = Math.round(startValue + ((top - startValue) * i) / count);
			cards.push(opts.extreme ? xcard(v) : card(v));
		}
		cards.push(topCard);
	}
	return {
		index,
		direction,
		startValue,
		topValue: top ?? startValue,
		topCard,
		cardCount: cards.length,
		cards,
		commandActive: opts.commandActive ?? false
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

/** Extreme hand with explicit legality (the command rules are the server's job). */
function xhand(cards: [value: number, legalPiles: number[], blocked?: ApiErrorCode][]): HandCard[] {
	return cards.map(([v, legal, blocked]) => ({
		...xcard(v),
		legalPiles: legal,
		...(blocked ? { blocked } : {})
	}));
}

const ME = 'p_cleo';

function players(
	opts: {
		active?: string;
		host?: string;
		counts?: number[];
		reveal?: number[][];
		extreme?: boolean;
	} = {}
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
			revealedHand: opts.reveal?.[seat]?.length
				? opts.reveal[seat].map(opts.extreme ? xcard : (v) => card(v))
				: undefined
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
	communicationRestricted: false,
	activeCommands: []
} satisfies Partial<PlayerView>;

const extremeBase = {
	...base,
	options: { variant: 'extreme', pro: false, reducedHandSize: false }
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

export const lobbyHostExtreme: PlayerView = {
	...lobbyHost,
	options: { variant: 'extreme', pro: true, reducedHandSize: false },
	rules: { ...base.rules, minCardsPerTurn: 3 }
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
		stuck: false,
		plays: [{ card: card(34), pile: 0 }],
		constraints: noConstraints
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
		stuck: false,
		plays: [
			{ card: card(77), pile: 2 },
			{ card: card(13), pile: 1 }
		],
		constraints: noConstraints
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

// ---- The Game Extreme -------------------------------------------------------

/**
 * My turn, 1 card played (34 on pile 0). Single pile (47) on pile 1 binds the
 * turn to pile 0, No backwards trick (59) on pile 2, an earlier turn's Skull
 * (87) on pile 3 has no effect any more. Hand: STOP 35, 3! 53; 3 is blocked by
 * Single pile and 24 (trick on pile 0) by No backwards trick.
 */
const extremeMidPiles = [
	pile(0, 34, 6, { extreme: true }),
	pile(1, 47, 4, { extreme: true, commandActive: true }),
	pile(2, 59, 5, { extreme: true, commandActive: true }),
	pile(3, 87, 2, { extreme: true })
];

export const extremeMyTurn: PlayerView = {
	...extremeBase,
	revision: 42,
	phase: 'playing',
	players: players({ active: ME, counts: [6, 6, 5, 6] }),
	hand: xhand([
		[3, [], 'single_pile_violation'],
		[24, [], 'backwards_trick_blocked'],
		[35, [0]],
		[53, [0]],
		[66, [0]]
	]),
	piles: extremeMidPiles,
	drawPileCount: 64,
	turn: {
		number: 9,
		activePlayerId: ME,
		cardsPlayed: 1,
		requiredMinimum: 2,
		canEndTurn: false,
		stuck: false,
		plays: [{ card: card(34), pile: 0 }],
		constraints: { ...noConstraints, singlePile: true, boundPile: 0 }
	},
	activeCommands: [
		{ command: 'singlePile', pile: 1 },
		{ command: 'noBackwardsTrick', pile: 2 }
	],
	result: null
};

/**
 * My turn after 3! (28 on pile 0) and a Skull (41 on pile 3): the third and
 * last card must cover the skull. STOP 95 would leave it showing.
 */
const skullTurn: Turn = {
	number: 12,
	activePlayerId: ME,
	cardsPlayed: 2,
	requiredMinimum: 2,
	canEndTurn: false,
	stuck: false,
	plays: [
		{ card: card(28, 'exactlyThree'), pile: 0 },
		{ card: card(41, 'skull'), pile: 3 }
	],
	constraints: { ...noConstraints, exactCount: 3, skullPiles: [3] }
};

export const extremeSkull: PlayerView = {
	...extremeBase,
	revision: 57,
	phase: 'playing',
	players: players({ active: ME, counts: [6, 6, 4, 6] }),
	hand: xhand([
		[7, [3]],
		[33, [3]],
		[39, [3]],
		[95, [], 'skull_uncovered']
	]),
	piles: [
		pile(0, 28, 7, { extreme: true, commandActive: true }),
		pile(1, 12, 2, { extreme: true }),
		pile(2, 71, 5, { extreme: true }),
		pile(3, 41, 6, { extreme: true, commandActive: true })
	],
	drawPileCount: 52,
	turn: skullTurn,
	result: null
};

/** Ben's turn; he just played No communication (65) on pile 2. Draw only one (43) lies on pile 1. */
export const extremeOthersNoComm: PlayerView = {
	...extremeBase,
	revision: 63,
	phase: 'playing',
	communicationRestricted: true,
	players: players({ active: 'p_ben', counts: [6, 5, 6, 6] }),
	hand: xhand([
		[6, []],
		[19, []],
		[50, []],
		[61, []],
		[84, []],
		[92, []]
	]),
	piles: [
		pile(0, 40, 9, { extreme: true }),
		pile(1, 43, 7, { extreme: true, commandActive: true }),
		pile(2, 65, 8, { extreme: true, commandActive: true }),
		pile(3, 80, 4, { extreme: true })
	],
	drawPileCount: 44,
	turn: {
		number: 14,
		activePlayerId: 'p_ben',
		cardsPlayed: 1,
		requiredMinimum: 2,
		canEndTurn: false,
		stuck: false,
		plays: [{ card: card(65, 'noCommunication'), pile: 2 }],
		constraints: noConstraints
	},
	activeCommands: [
		{ command: 'drawOnlyOne', pile: 1 },
		{ command: 'noCommunication', pile: 2 }
	],
	result: null
};

/** Solo hand of 8 command cards: the widest hand with an icon on every card. */
export const extremeEightCommands: PlayerView = {
	...extremeMyTurn,
	rules: { ...base.rules, handSize: 8 },
	players: players({ active: ME, counts: [8, 8, 8, 8] }),
	hand: xhand([
		[9, [2, 3]],
		[13, [2, 3]],
		[17, [2, 3]],
		[21, [2, 3]],
		[26, [2, 3]],
		[53, [0, 1, 2, 3]],
		[78, [0, 1, 3]],
		[95, [0, 1]]
	]),
	turn: {
		...extremeMyTurn.turn!,
		cardsPlayed: 0,
		plays: [],
		constraints: { ...noConstraints, singlePile: true }
	},
	activeCommands: [
		{ command: 'singlePile', pile: 1 },
		{ command: 'noBackwardsTrick', pile: 2 }
	]
};

/** Lost with score 0: Dan's last card was a Skull (87 on pile 1) that nobody could cover. */
export const finishedLostSkull: PlayerView = {
	...extremeBase,
	revision: 230,
	phase: 'finished',
	players: players({ counts: [0, 0, 0, 0], extreme: true }),
	hand: [],
	piles: [
		pile(0, 99, 26, { extreme: true }),
		pile(1, 87, 23, { extreme: true, commandActive: true }),
		pile(2, 2, 25, { extreme: true }),
		pile(3, 4, 24, { extreme: true })
	],
	drawPileCount: 0,
	turn: {
		number: 61,
		activePlayerId: 'p_dan',
		cardsPlayed: 2,
		requiredMinimum: 1,
		canEndTurn: false,
		stuck: false,
		plays: [
			{ card: card(4), pile: 3 },
			{ card: card(87, 'skull'), pile: 1 }
		],
		constraints: { ...noConstraints, skullPiles: [1] }
	},
	result: {
		outcome: 'lost',
		score: 0,
		excellent: true,
		cardsLeftInHands: 0,
		cardsLeftInDrawPile: 0,
		stuckPlayerId: 'p_dan'
	}
};

export const fixtures = {
	lobbyGuest,
	lobbyHost,
	lobbyHostExtreme,
	choosingGuest,
	choosingHost,
	playingMyTurn,
	playingEightCards,
	playingOthersTurn,
	finishedWon,
	finishedLost,
	extremeMyTurn,
	extremeSkull,
	extremeOthersNoComm,
	extremeEightCommands,
	finishedLostSkull
} satisfies Record<string, PlayerView>;
