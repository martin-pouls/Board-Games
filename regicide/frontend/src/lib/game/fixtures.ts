/**
 * Sample PlayerViews for every phase. Used by tests and handy for layout work
 * (render a component with one of these instead of a live connection).
 */
import type {
	Card,
	Enemy,
	PlayedSet,
	PlayerSummary,
	PlayerView,
	Rules,
	Suit,
	Turn
} from '#lib/types.js';
import { cardValue, enemyStats, sortCards } from './rules.js';

const SUIT_LETTERS: Record<string, Suit> = { H: 'hearts', D: 'diamonds', C: 'clubs', S: 'spades' };

/** Parses "5H", "10D", "KC", "AS"; "X" is the Jester. */
export function card(s: string): Card {
	if (s === 'X') return { suit: '', rank: 0 };
	const suit = SUIT_LETTERS[s.slice(-1)];
	const r = s.slice(0, -1);
	const rank = r === 'A' ? 1 : r === 'J' ? 11 : r === 'Q' ? 12 : r === 'K' ? 13 : Number(r);
	if (!suit || !Number.isInteger(rank) || rank < 1 || rank > 13) throw new Error(`bad card ${s}`);
	return { suit, rank };
}

/** Parses a space-separated list: "5H 5C X". */
export const cards = (s: string): Card[] => (s.trim() ? s.trim().split(/\s+/).map(card) : []);
/** Like `cards`, but sorted like a real hand. */
export const hand = (s: string): Card[] => sortCards(cards(s));

export const rulesFor = (players: number): Rules => ({
	minPlayers: 1,
	maxPlayers: 4,
	maxHandSize: [8, 7, 6, 5][players - 1],
	tavernJesters: [0, 0, 1, 2][players - 1],
	soloJesters: players === 1 ? 2 : 0
});

export function player(
	id: string,
	name: string,
	seat: number,
	o: Partial<PlayerSummary> = {}
): PlayerSummary {
	return {
		id,
		name,
		seat,
		handCount: 7,
		isHost: seat === 0,
		isYou: false,
		isActive: false,
		yieldedLastTurn: false,
		signal: 'none',
		...o
	};
}

export interface PlayInput {
	by: string;
	cards: string;
	/** Suits whose power was blocked by immunity at the time. */
	blocked?: Suit[];
	doubled?: boolean;
}

/** An enemy with derived numbers; shield and damage are computed from `plays` unless overridden. */
export function enemyOf(
	cardStr: string,
	opts: { plays?: PlayInput[]; cancelled?: boolean; damage?: number; shield?: number } = {}
): Enemy {
	const c = card(cardStr);
	const stats = enemyStats(c.rank);
	const plays: PlayedSet[] = (opts.plays ?? []).map((p) => {
		const cs = cards(p.cards);
		const value = cs.reduce((n, x) => n + cardValue(x), 0);
		const jester = cs.length === 1 && cs[0].rank === 0;
		const doubled = p.doubled ?? false;
		return {
			playerId: p.by,
			cards: cs,
			value,
			damage: doubled ? value * 2 : value,
			doubled,
			blockedSuits: p.blocked ?? [],
			jester
		};
	});
	const cancelled = opts.cancelled ?? plays.some((p) => p.jester);
	const damage = opts.damage ?? plays.reduce((n, p) => n + p.damage, 0);
	const spadeSum = plays
		.filter((p) => p.cards.some((x) => x.suit === 'spades'))
		.reduce((n, p) => n + p.value, 0);
	const shield = opts.shield ?? (c.suit === 'spades' && !cancelled ? 0 : spadeSum);
	return {
		card: c,
		attack: stats.attack,
		health: stats.health,
		damage,
		shield,
		effectiveAttack: Math.max(0, stats.attack - shield),
		immunity: c.suit as Suit,
		immunityCancelled: cancelled,
		plays
	};
}

/** Builds a complete view from sensible defaults (2 players, I am Anna, seat 0, my turn). */
export function makeView(o: Partial<PlayerView> = {}): PlayerView {
	const players = o.players ?? [
		player('p1', 'Anna', 0, { isYou: true, isActive: true }),
		player('p2', 'Ben', 1)
	];
	return {
		gameId: 'ABCD',
		revision: 1,
		phase: 'playing',
		you: 'p1',
		hostId: 'p1',
		rules: rulesFor(players.length),
		players,
		hand: [],
		enemy: null,
		castleCount: 11,
		tavernCount: 30,
		discardCount: 0,
		lastTurn: null,
		turn: null,
		soloJestersLeft: 0,
		result: null,
		...o
	};
}

function turn(o: Partial<Turn> = {}): Turn {
	return { number: 1, activePlayerId: 'p1', step: 'play', canYield: true, damageToCover: 0, ...o };
}

export const lobbyView: PlayerView = makeView({
	phase: 'lobby',
	players: [
		player('p1', 'Anna', 0, { isYou: true, handCount: 0 }),
		player('p2', 'Ben', 1, { handCount: 0 })
	],
	castleCount: 0,
	tavernCount: 0
});

/** My turn, J♠ with a Jester and a 4♠ played against it (immunity cancelled, shield 4). */
export const playView: PlayerView = makeView({
	hand: hand('2H 5D 5C 8C 9S KD AH'),
	enemy: enemyOf('JS', {
		plays: [
			{ by: 'p1', cards: 'X' },
			{ by: 'p2', cards: '4S' }
		]
	}),
	tavernCount: 24,
	discardCount: 5,
	lastTurn: { playerId: 'p2', played: cards('4S'), discarded: cards('2D 3C') },
	turn: turn({ number: 4 })
});

/** Ben's turn; I am waiting. */
export const otherTurnView: PlayerView = makeView({
	players: [
		player('p1', 'Anna', 0, { isYou: true }),
		player('p2', 'Ben', 1, { isActive: true, yieldedLastTurn: true })
	],
	hand: hand('2H 5D 5C 8C 9S KD AH'),
	enemy: enemyOf('QH', { plays: [{ by: 'p1', cards: '6D' }] }),
	castleCount: 7,
	tavernCount: 26,
	discardCount: 3,
	lastTurn: { playerId: 'p2', played: null, discarded: cards('KD') },
	turn: turn({ number: 5, activePlayerId: 'p2', canYield: false })
});

/** My discard step: Q♥ strikes for 15. */
export const discardView: PlayerView = makeView({
	hand: hand('3C 6D 7H 8D 9C 10S KS'),
	enemy: enemyOf('QH', { plays: [{ by: 'p1', cards: '8H', blocked: ['hearts'] }] }),
	castleCount: 7,
	tavernCount: 22,
	discardCount: 6,
	lastTurn: { playerId: 'p1', played: cards('8H'), discarded: null },
	turn: turn({ number: 6, step: 'discard', canYield: false, damageToCover: 15 })
});

/** I played a Jester and must choose who goes next. */
export const chooseNextChooserView: PlayerView = makeView({
	players: [
		player('p1', 'Anna', 0, { isYou: true, isActive: true }),
		player('p2', 'Ben', 1, { signal: 'goodPlay' }),
		player('p3', 'Cara', 2, { signal: 'none' })
	],
	hand: hand('2H 5D 5C 8C 9S KD'),
	enemy: enemyOf('KC', { plays: [{ by: 'p1', cards: 'X' }] }),
	castleCount: 3,
	turn: turn({ number: 7, step: 'chooseNext', canYield: false })
});

/** Ben played a Jester; I (Cara) may signal. */
export const chooseNextOtherView: PlayerView = makeView({
	you: 'p3',
	players: [
		player('p1', 'Anna', 0),
		player('p2', 'Ben', 1, { isActive: true }),
		player('p3', 'Cara', 2, { isYou: true, signal: 'ratherNot' })
	],
	hostId: 'p1',
	hand: hand('3D 4H 8S 9C 10H QD'),
	enemy: enemyOf('KC', { plays: [{ by: 'p2', cards: 'X' }] }),
	castleCount: 3,
	turn: turn({ number: 7, activePlayerId: 'p2', step: 'chooseNext', canYield: false })
});

/** Ben just defeated the last Jack, of hearts, exactly with 10♣ and plays on against a Queen. */
export const afterDefeatView: PlayerView = makeView({
	players: [
		player('p1', 'Anna', 0, { isYou: true }),
		player('p2', 'Ben', 1, { isActive: true, handCount: 4 })
	],
	hand: hand('3H 4S 6D 9D 9S 10S'),
	enemy: enemyOf('QH'),
	castleCount: 7,
	tavernCount: 24,
	discardCount: 9,
	lastTurn: {
		playerId: 'p2',
		played: cards('10C'),
		defeated: card('JH'),
		exact: true,
		discarded: null
	},
	turn: turn({ number: 8, activePlayerId: 'p2' })
});

/** Like afterDefeatView, but 7♣ 7♦ (doubled to 28) overshot: the Jack went to the discard pile. */
export const afterOverkillView: PlayerView = makeView({
	...afterDefeatView,
	discardCount: 12,
	lastTurn: {
		playerId: 'p2',
		played: cards('7C 7D'),
		defeated: card('JH'),
		exact: false,
		discarded: null
	}
});

/** Four players with longer names: the header and the plays against the enemy at their widest. */
export const fourPlayerView: PlayerView = makeView({
	players: [
		player('p1', 'Anna', 0, { isYou: true, handCount: 5 }),
		player('p2', 'Benedikt', 1, { handCount: 4, yieldedLastTurn: true }),
		player('p3', 'Charlotte', 2, { isActive: true, handCount: 5 }),
		player('p4', 'Dan', 3, { handCount: 3 })
	],
	hand: hand('3H 4S 6D 9D 10S'),
	enemy: enemyOf('KS', {
		plays: [
			{ by: 'p1', cards: '5C 5D' },
			{ by: 'p2', cards: '2H 2D 2S' },
			{ by: 'p3', cards: 'AC 9H' },
			{ by: 'p4', cards: '7D' }
		]
	}),
	castleCount: 2,
	tavernCount: 12,
	discardCount: 31,
	lastTurn: { playerId: 'p2', played: null, discarded: cards('2C 3D 4H 5S 6C 7H') },
	turn: turn({ number: 33, activePlayerId: 'p3' })
});

export const soloPlayView: PlayerView = makeView({
	players: [player('p1', 'Anna', 0, { isYou: true, isActive: true, handCount: 8 })],
	hand: hand('2H 3D 4C 5S 7H 9D 10C QS'),
	enemy: enemyOf('JD'),
	soloJestersLeft: 2,
	turn: turn({ number: 1 })
});

/** Solo discard: K♥ strikes for 20 but my hand is worth 14; one Jester power left. */
export const soloDiscardShortView: PlayerView = makeView({
	players: [player('p1', 'Anna', 0, { isYou: true, isActive: true, handCount: 4 })],
	hand: hand('2H 3D 4C 5S'),
	enemy: enemyOf('KH'),
	castleCount: 3,
	soloJestersLeft: 1,
	turn: turn({ number: 30, step: 'discard', canYield: true, damageToCover: 20 })
});

export const wonSoloView: PlayerView = makeView({
	phase: 'won',
	players: [
		player('p1', 'Anna', 0, { isYou: true, handCount: 5, revealedHand: hand('2H 3D 4C 5S 6H') })
	],
	hand: hand('2H 3D 4C 5S 6H'),
	castleCount: 0,
	soloJestersLeft: 1,
	result: { outcome: 'won', grade: 'silver', enemiesDefeated: 12, soloJestersUsed: 1 }
});

export const lostView: PlayerView = makeView({
	phase: 'lost',
	players: [
		player('p1', 'Anna', 0, { isYou: true, handCount: 2, revealedHand: hand('2H 3D') }),
		player('p2', 'Ben', 1, { handCount: 3, revealedHand: hand('4C 5S 6H') })
	],
	hand: hand('2H 3D'),
	enemy: enemyOf('QS'),
	castleCount: 6,
	turn: turn({ number: 40, step: '', canYield: false }),
	result: {
		outcome: 'lost',
		reason: 'cannotCoverDamage',
		playerId: 'p1',
		enemiesDefeated: 5,
		soloJestersUsed: 0
	}
});

/** A full lobby seen by a guest (Cleo, seat 2). */
export const lobbyGuestView: PlayerView = makeView({
	phase: 'lobby',
	you: 'p3',
	players: [
		player('p1', 'Anna', 0, { handCount: 0 }),
		player('p2', 'Ben', 1, { handCount: 0 }),
		player('p3', 'Cleo', 2, { isYou: true, handCount: 0 }),
		player('p4', 'Dan', 3, { handCount: 0 })
	],
	castleCount: 0,
	tavernCount: 0
});

/** The host closed the lobby before the game started. */
export const endedView: PlayerView = makeView({
	...lobbyView,
	phase: 'ended',
	you: 'p2',
	players: lobbyView.players.map((p) => ({ ...p, isYou: p.id === 'p2' }))
});

/**
 * Every named fixture; each call returns a fresh copy. The dev route `/dev/fixtures/<name>` shows
 * them on the game screen.
 */
export const fixtures = Object.fromEntries(
	Object.entries({
		lobbyView,
		lobbyGuestView,
		playView,
		otherTurnView,
		discardView,
		chooseNextChooserView,
		chooseNextOtherView,
		afterDefeatView,
		afterOverkillView,
		fourPlayerView,
		soloPlayView,
		soloDiscardShortView,
		wonSoloView,
		lostView,
		endedView
	}).map(([name, v]) => [name, () => structuredClone(v)])
) as Record<string, () => PlayerView>;
