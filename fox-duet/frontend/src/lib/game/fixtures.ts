// Sample views for tests and the dev fixtures route (/dev/fixtures). Not in production builds.
import { MOVEMENT } from './format.js';
import type { Ability, Card, Pending, PlayerView, Space, Suit } from '#lib/types.js';

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

const pending = (o: Partial<Pending> & Pick<Pending, 'step' | 'playerIds'>): Pending => ({
	moveOptions: [],
	ends: [],
	yourGift: null,
	...o
});

/** Seen by the other seat: Bob instead of Alice. */
function asBob(v: PlayerView, hand: Card[] = v.hand): PlayerView {
	return {
		...v,
		you: 'bob',
		players: v.players.map((p) => ({ ...p, isYou: p.id === 'bob' })),
		hand,
		playable: v.pending?.playerIds.includes('bob') ? v.playable : []
	};
}

function lobby(players: 1 | 2): PlayerView {
	const v = playingView();
	return {
		...v,
		revision: players,
		phase: 'lobby',
		level: 0,
		players: v.players
			.slice(0, players)
			.map((p) => ({ ...p, handCount: 0, isDealer: false, mustAct: false })),
		hand: [],
		playable: [],
		board: null,
		round: 0,
		decree: null,
		trick: null,
		pending: null
	};
}

/** The host alone in the lobby. */
export const lobbyAloneView = () => lobby(1);
/** Both seats taken, seen by the host. */
export const lobbyHostView = () => lobby(2);
/** Both seats taken, seen by the guest. */
export const lobbyGuestView = () => asBob(lobby(2));
/** The host closed the lobby (seen by the guest). */
export const endedView = (): PlayerView => ({ ...lobbyGuestView(), phase: 'ended' });

/** Trick 4 of round 1: Bob led the 8 of stars, Alice must follow suit. */
export function followView(): PlayerView {
	const v = playingView();
	const hand = v.hand.filter((c) => !(c.suit === 'stars' && c.rank === 3)).slice(0, 8);
	return {
		...v,
		revision: 20,
		players: v.players.map((p) => ({ ...p, handCount: p.id === 'alice' ? 8 : 7 })),
		hand,
		playable: hand.filter((c) => c.suit === 'stars'),
		board: { ...v.board!, traveller: 7 },
		trick: { number: 4, leaderId: 'bob', lead: card('stars', 8), follow: null },
		lastTrick: {
			number: 3,
			leaderId: 'alice',
			lead: card('roses', 2),
			follow: card('roses', 11),
			winnerId: 'bob',
			from: 6,
			to: 7,
			delta: 1,
			beyond: '',
			collected: true
		},
		pending: pending({ step: 'follow', playerIds: ['alice'] })
	};
}

/** Alice led, Bob is to follow: Alice waits. */
export function waitingView(): PlayerView {
	const v = playingView();
	return {
		...v,
		revision: 4,
		players: v.players.map((p) => ({ ...p, handCount: p.id === 'alice' ? 10 : 11 })),
		hand: v.hand.filter((c) => !(c.suit === 'roses' && c.rank === 5)),
		playable: [],
		trick: { number: 1, leaderId: 'alice', lead: card('roses', 5), follow: null },
		pending: pending({ step: 'follow', playerIds: ['bob'] })
	};
}

/** Alice led a 3 (Foxes) and chooses who may exchange the decree card. */
export function foxesChooseView(): PlayerView {
	const v = waitingView();
	return {
		...v,
		hand: playingView().hand.filter((c) => !(c.suit === 'stars' && c.rank === 3)),
		trick: { number: 1, leaderId: 'alice', lead: card('stars', 3), follow: null },
		pending: pending({ step: 'foxesChoose', playerIds: ['alice'] })
	};
}

/** Alice chose herself: she may swap the decree card for a card from her hand. */
export function foxesSwapView(): PlayerView {
	const v = foxesChooseView();
	return {
		...v,
		playable: [...v.hand],
		pending: pending({ step: 'foxesSwap', playerIds: ['alice'], chosenId: 'alice' })
	};
}

/** Alice led a 7 (Gift): both players pick a card to give. */
export function giftView(): PlayerView {
	const v = waitingView();
	const hand = playingView().hand.filter((c) => !(c.suit === 'doves' && c.rank === 7));
	return {
		...v,
		hand,
		playable: [...hand],
		trick: { number: 1, leaderId: 'alice', lead: card('doves', 7), follow: null },
		pending: pending({ step: 'gift', playerIds: ['alice', 'bob'] })
	};
}

/** Alice chose her gift; Bob still chooses. */
export function giftChosenView(): PlayerView {
	const v = giftView();
	return {
		...v,
		playable: [],
		pending: pending({ step: 'gift', playerIds: ['bob'], yourGift: card('stars', 9) })
	};
}

/** Alice won a trick with a Musician (1) in it: she picks the direction. */
export function moveView(): PlayerView {
	const v = playingView();
	const hand = v.hand.filter((c) => !(c.suit === 'roses' && c.rank === 1));
	return {
		...v,
		revision: 6,
		players: v.players.map((p) => ({ ...p, handCount: 10 })),
		hand,
		playable: [],
		trick: {
			number: 1,
			leaderId: 'alice',
			lead: card('roses', 1),
			follow: card('roses', 4),
			winnerId: 'alice'
		},
		pending: pending({
			step: 'move',
			playerIds: ['alice'],
			moveOptions: [
				{ delta: -2, to: 3, beyond: '', gem: true, lost: false },
				{ delta: 2, to: 7, beyond: '', gem: true, lost: false }
			]
		})
	};
}

/** End of round 1: the host chooses the end for the forest token. */
export function forestEndView(): PlayerView {
	const v = playingView();
	return {
		...v,
		revision: 60,
		players: v.players.map((p) => ({ ...p, handCount: 0 })),
		hand: [],
		playable: [],
		board: { ...v.board!, traveller: 4, gemsLeft: 7 },
		trick: null,
		lastTrick: {
			number: 11,
			leaderId: 'bob',
			lead: card('doves', 4),
			follow: card('doves', 9),
			winnerId: 'alice',
			from: 6,
			to: 4,
			delta: -2,
			beyond: '',
			collected: true
		},
		pending: pending({ step: 'forestEnd', playerIds: ['alice'], ends: ['left', 'right'] })
	};
}

/** Bob won trick 5 and leads trick 6; the completed trick stays on the table until he leads. */
export function trickWonView(): PlayerView {
	const v = playingView();
	const hand = [
		card('doves', 2),
		card('doves', 10),
		card('roses', 5),
		card('roses', 8),
		card('stars', 4),
		card('stars', 11)
	];
	const spaces = sideA();
	spaces[3].gems = 0;
	spaces[7].gems = 0;
	return {
		...v,
		revision: 30,
		players: v.players.map((p) => ({ ...p, handCount: 6, mustAct: p.id === 'bob' })),
		hand,
		playable: [],
		board: { ...v.board!, spaces, traveller: 8, gemsLeft: 10 },
		trick: { number: 6, leaderId: 'bob', lead: null, follow: null },
		lastTrick: {
			number: 5,
			leaderId: 'alice',
			lead: card('doves', 7),
			follow: card('doves', 9),
			winnerId: 'bob',
			from: 7,
			to: 8,
			delta: 1,
			beyond: '',
			collected: false
		},
		pending: pending({ step: 'lead', playerIds: ['bob'] })
	};
}

/** Round 2 has just been dealt: gems were added and a forest token covers Bob's last space. */
export function newRoundView(): PlayerView {
	const v = playingView();
	const spaces = sideA().map((s) => ({ ...s, gems: s.plus ? 1 : 0 }));
	spaces[2].gems = 1;
	spaces[9].gems = 2;
	spaces[10].covered = true;
	const hand = [
		card('doves', 1),
		card('doves', 4),
		card('doves', 8),
		card('roses', 2),
		card('roses', 6),
		card('roses', 7),
		card('roses', 9),
		card('stars', 1),
		card('stars', 3),
		card('stars', 5),
		card('stars', 10)
	];
	return {
		...v,
		revision: 70,
		round: 2,
		players: v.players.map((p) => ({ ...p, isDealer: p.id === 'alice', mustAct: p.id === 'bob' })),
		hand,
		playable: [],
		board: {
			...v.board!,
			spaces,
			traveller: 5,
			forestTokens: 3,
			gemsLeft: spaces.reduce((n, s) => n + s.gems, 0)
		},
		decree: card('roses', 4),
		trick: { number: 1, leaderId: 'bob', lead: null, follow: null },
		pending: pending({ step: 'lead', playerIds: ['bob'] })
	};
}

/** Won in round 2 at level 1. */
export function wonView(): PlayerView {
	const v = forestEndView();
	const left = [card('stars', 2), card('roses', 6)];
	return {
		...v,
		phase: 'won',
		round: 2,
		players: v.players.map((p, i) => ({
			...p,
			handCount: i === 0 ? 2 : 1,
			mustAct: false,
			revealedHand: i === 0 ? left : [card('doves', 11)]
		})),
		hand: left,
		board: { ...v.board!, gemsLeft: 0, forestTokens: 3 },
		pending: null,
		result: {
			outcome: 'won',
			round: 2,
			gemsLeft: 0,
			score: { level: 10, cards: 3, earlyWin: 10, forestTokens: 9, total: 32 }
		}
	};
}

/** Lost in the forest in round 3. */
export function lostView(): PlayerView {
	const v = wonView();
	return {
		...v,
		phase: 'lost',
		round: 3,
		board: { ...v.board!, gemsLeft: 4, forestTokens: 0 },
		result: { outcome: 'lost', reason: 'lostInForest', round: 3, gemsLeft: 4, score: null }
	};
}

/**
 * Every named fixture; each call returns a fresh view. The dev route `/dev/fixtures/<name>`
 * shows them on the game screen.
 */
export const fixtures: Record<string, () => PlayerView> = {
	lobbyAloneView,
	lobbyHostView,
	lobbyGuestView,
	playingView,
	waitingView,
	followView,
	foxesChooseView,
	foxesSwapView,
	giftView,
	giftChosenView,
	moveView,
	trickWonView,
	forestEndView,
	newRoundView,
	wonView,
	lostView,
	endedView
};
