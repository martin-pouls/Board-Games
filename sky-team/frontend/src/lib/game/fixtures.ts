// Sample views for tests and the dev fixtures route (/dev/fixtures). Not in production builds.
//
// Every game view is built from a small spec by `flight()`, which derives the rest the way the
// server does (aero markers, brake limit, speed, plane supply, reroll tokens, `can`, `legal`,
// per-player counts), so the numbers on one screen always agree. The worst cases (TGU red, KEF
// black, HND red) use the data of layout concept A (sky-team/docs/layout-concepts.html). Every
// function returns a fresh object.
import type {
	AbilityId,
	Condition,
	FlightView,
	LossReason,
	Module,
	Phase,
	Piece,
	Placement,
	PlayerSummary,
	PlayerView,
	Role,
	ScenarioInfo,
	SpaceId,
	Step
} from '#lib/types.js';
import {
	internTargets,
	legalByValue,
	syncTargets,
	windValue,
	type PanelState
} from './fixtureLegal.js';
import { firstPlayerOf, otherRole } from './format.js';
import type { TablePreset } from './tablePreset.js';

// ------------------------------------------------------------------------------- scenarios

type Colour = ScenarioInfo['colour'];

// [id, airport, city, colour, modules, abilities, real-time dropped, approach spaces]
type Row = [string, string, string, Colour, Module[], number, boolean, number];

const ROWS: Row[] = [
	['yul-green', 'YUL', 'Montréal-Trudeau', 'green', [], 0, false, 7],
	['lhr-green', 'LHR', 'Heathrow', 'green', [], 0, false, 6],
	['hnd-green', 'HND', 'Haneda', 'green', [], 0, false, 8],
	['osl-green', 'OSL', 'Gardermoen', 'green', ['kerosene'], 0, false, 8],
	['atl-green', 'ATL', 'Hartsfield-Jackson', 'green', ['intern'], 0, false, 8],
	['prg-green', 'PRG', 'Václav Havel', 'green', ['kerosene'], 2, false, 8],
	['lhr-yellow', 'LHR', 'Heathrow', 'yellow', ['intern'], 0, false, 6],
	['tgu-yellow', 'TGU', 'Toncontín', 'yellow', ['kerosene'], 2, false, 5],
	['gig-yellow', 'GIG', 'Galeão', 'yellow', ['wind'], 1, false, 7],
	['kef-yellow', 'KEF', 'Keflavík', 'yellow', ['iceBrakes'], 1, false, 6],
	['prg-yellow', 'PRG', 'Václav Havel', 'yellow', ['keroseneLeak'], 2, false, 8],
	['kul-yellow', 'KUL', 'Kuala Lumpur', 'yellow', ['kerosene'], 1, false, 8],
	['atl-yellow', 'ATL', 'Hartsfield-Jackson', 'yellow', ['keroseneLeak'], 1, false, 8],
	['pbh-red', 'PBH', 'Paro', 'red', ['kerosene'], 2, true, 6],
	['hnd-red', 'HND', 'Haneda', 'red', ['kerosene', 'intern'], 1, false, 8],
	['gig-red', 'GIG', 'Galeão', 'red', ['wind', 'keroseneLeak'], 2, false, 7],
	['osl-red', 'OSL', 'Gardermoen', 'red', ['keroseneLeak', 'iceBrakes'], 2, false, 8],
	['tgu-red', 'TGU', 'Toncontín', 'red', ['kerosene', 'wind'], 2, false, 5],
	['kef-black', 'KEF', 'Keflavík', 'black', ['wind', 'iceBrakes'], 2, false, 6],
	['kul-black', 'KUL', 'Kuala Lumpur', 'black', ['kerosene'], 2, true, 8],
	['pbh-black', 'PBH', 'Paro', 'black', ['kerosene'], 2, true, 6],
	['dus-yellow', 'DUS', 'Düsseldorf', 'yellow', ['keroseneLeak'], 2, false, 8],
	['dus-black', 'DUS', 'Düsseldorf', 'black', ['keroseneLeak'], 1, true, 8]
];

/** The 23 scenarios in rules.md §9 order (the lobby view's catalogue). */
export const catalogue = (): ScenarioInfo[] =>
	ROWS.map(([id, airport, city, colour, modules, abilities, realTimeDropped, spaces]) => ({
		id,
		airport,
		city,
		colour,
		modules: [...modules],
		abilities,
		realTimeDropped,
		spaces
	}));

const scenarioOf = (id: string): ScenarioInfo => {
	const s = catalogue().find((c) => c.id === id);
	if (!s) throw new Error(`unknown scenario ${id}`);
	return s;
};

// ---------------------------------------------------------------------------------- tracks

interface Track {
	/** Traffic-die icons per space. */
	traffic: number[];
	/** Turn masks per space (rules.md §10); null = no icon. */
	turns: (string | null)[];
}

const none = (n: number) => Array<null>(n).fill(null);
const zeros = (n: number) => Array<number>(n).fill(0);

const YUL: Track = { traffic: zeros(7), turns: none(7) };
const TGU_RED: Track = {
	traffic: [3, 0, 2, 0, 0],
	turns: [null, 'ooxxx', 'xooxx', 'ooxxx', null]
};
const KEF_BLACK: Track = {
	traffic: [2, 0, 1, 0, 0, 0],
	turns: [null, 'oooxx', null, 'xxooo', 'xooox', null]
};
const HND_RED: Track = {
	traffic: [3, 0, 0, 1, 1, 1, 0, 0],
	turns: [null, 'xooxx', null, 'ooxxx', 'xooxx', null, 'ooxxx', null]
};

// ----------------------------------------------------------------------------------- people

const SEATS = [
	{ id: 'ann', name: 'Ann', seat: 0, role: 'pilot' as Role },
	{ id: 'ben', name: 'Ben', seat: 1, role: 'copilot' as Role }
];

const idOf = (role: Role) => (role === 'pilot' ? 'ann' : 'ben');

// ----------------------------------------------------------------------------------- lobby

interface LobbySpec {
	players?: 1 | 2;
	scenario?: string;
	abilities?: AbilityId[];
	me?: Role;
}

function lobby({
	players = 2,
	scenario = 'yul-green',
	abilities = [],
	me = 'pilot'
}: LobbySpec = {}): PlayerView {
	const seats: PlayerSummary[] = SEATS.slice(0, players).map((s) => ({
		id: s.id,
		name: s.name,
		seat: s.seat,
		role: s.role,
		isHost: s.id === 'ann',
		isYou: s.role === me,
		mustAct: false,
		diceLeft: 0,
		ready: false,
		revealedDice: null
	}));
	return {
		gameId: 'K7MQ2',
		revision: players,
		phase: 'lobby',
		you: idOf(me),
		hostId: 'ann',
		nextGameId: '',
		players: seats,
		scenario: scenarioOf(scenario),
		abilities,
		catalogue: catalogue(),
		flight: null,
		result: null
	};
}

/** The host alone in the lobby. */
export const lobbyAloneView = () => lobby({ players: 1 });
/** Both seats taken, seen by the host. */
export const lobbyHostView = () => lobby();
/** TGU red (two ability cards), both picked. */
export const lobbyHostAbilitiesView = () =>
	lobby({ scenario: 'tgu-red', abilities: ['synchronisation', 'workingTogether'] });
/** Both seats taken, seen by the guest: PRG green with one of its two cards picked so far. */
export const lobbyGuestView = () =>
	lobby({ scenario: 'prg-green', abilities: ['control'], me: 'copilot' });
/** The host closed the lobby (seen by the guest). */
export const endedView = (): PlayerView => ({ ...lobbyGuestView(), phase: 'ended' });

// ----------------------------------------------------------------------------------- flight

type Dice = (number | null)[];

interface Spec {
	scenario: string;
	abilities?: AbilityId[];
	me?: Role;
	phase?: Phase;
	round: number;
	step?: Step;
	/** Defaults to the round's first player. */
	turn?: Role;
	track: Track;
	/** Plane tokens per approach space. */
	planes: number[];
	/** Approach space, 1-based. */
	position: number;
	axis?: number;
	gear?: [boolean, boolean, boolean];
	flaps?: number;
	brakes?: number;
	/** Ice marker slot 0…4. */
	ice?: number;
	coffee?: number;
	rerolls?: number;
	fuel?: number;
	/** Wind ring offset from the top space. */
	wind?: number;
	/** Token values in row order and the indices already used. */
	intern?: { values: number[]; taken: number[] };
	placements?: Placement[];
	dice?: Partial<Record<Role, Dice>>;
	ready?: Partial<Record<Role, boolean>>;
	roundStart?: FlightView['roundStart'];
	abilityState?: Partial<FlightView['abilityState']>;
	rerollConfirmed?: Partial<Record<Role, boolean>>;
	swap?: FlightView['swap'];
	sync?: number;
	result?: PlayerView['result'];
}

/** A die of `role`, or `piece` for the intern token / the black die. */
const die = (role: Role, space: SpaceId, value: number, coffee = 0): Placement => ({
	space,
	piece: role,
	value,
	coffee,
	role
});
const token = (role: Role, space: SpaceId, value: number): Placement => ({
	space,
	piece: 'intern' as Piece,
	value,
	coffee: 0,
	role
});
const black = (space: SpaceId, value: number): Placement => ({
	space,
	piece: 'traffic' as Piece,
	value,
	coffee: 0,
	role: null
});

const NO_DICE: Dice = [null, null, null, null];
const count = (d: Dice) => d.filter((v) => v !== null).length;

function flight(s: Spec): PlayerView {
	const me = s.me ?? 'pilot';
	const partner = otherRole(me);
	const phase = s.phase ?? 'playing';
	const step = s.step ?? 'place';
	const info = scenarioOf(s.scenario);
	const abilities = s.abilities ?? [];
	const placements = s.placements ?? [];
	const dice: Record<Role, Dice> = {
		pilot: s.dice?.pilot ?? NO_DICE,
		copilot: s.dice?.copilot ?? NO_DICE
	};
	const first = firstPlayerOf(s.round);
	const turn = s.turn ?? first;
	const playing = phase === 'playing';
	const gear = s.gear ?? [false, false, false];
	const flaps = s.flaps ?? 0;
	const brakes = s.brakes ?? 0;
	const ice = info.modules.includes('iceBrakes') ? (s.ice ?? 0) : null;

	// The intern board: pilot takes from the left end, co-pilot from the right end.
	const tokens = s.intern?.values.map((value, i) => ({
		value,
		taken: s.intern!.taken.includes(i)
	}));
	const free = tokens?.flatMap((t, i) => (t.taken ? [] : [i])) ?? [];
	const nextPilot = free.length ? free[0] : null;
	const nextCopilot = free.length ? free[free.length - 1] : null;
	const intern = tokens ? { tokens, nextPilot, nextCopilot } : null;

	const panel: PanelState = {
		modules: info.modules,
		placements,
		flaps,
		brakes,
		ice,
		internNext: {
			pilot: nextPilot === null || !tokens ? null : tokens[nextPilot].value,
			copilot: nextCopilot === null || !tokens ? null : tokens[nextCopilot].value
		}
	};

	const aero = { blue: 4 + gear.filter(Boolean).length, orange: 8 + flaps };
	const wind = s.wind === undefined ? null : { offset: s.wind, value: windValue(s.wind) };
	const engines = placements.filter((p) => p.space.startsWith('engines'));
	const speed =
		engines.length === 2
			? (() => {
					const sum = engines[0].value + engines[1].value;
					const total = sum + (wind?.value ?? 0);
					const move = s.round === 7 ? null : total <= aero.blue ? 0 : total <= aero.orange ? 1 : 2;
					return { sum, wind: wind?.value ?? 0, total, move };
				})()
			: null;

	const mine = count(dice[me]);
	const theirs = count(dice[partner]);
	const inPlace = playing && step === 'place';
	const has = (a: AbilityId) => abilities.includes(a);
	const abilityState: FlightView['abilityState'] = {
		adaptationUsed: { pilot: false, copilot: false },
		anticipationUsed: false,
		togetherUsed: false,
		syncUsed: false,
		controlTriggered: false,
		masteryTriggered: false,
		...s.abilityState
	};
	const can = {
		spendReroll: inPlace && (s.rerolls ?? 0) > 0,
		adapt: inPlace && has('adaptation') && !abilityState.adaptationUsed[me] && mine > 0,
		anticipate:
			inPlace &&
			has('anticipation') &&
			first === me &&
			turn === me &&
			placements.length === 0 &&
			!abilityState.anticipationUsed &&
			mine > 0,
		offerSwap:
			inPlace && has('workingTogether') && !abilityState.togetherUsed && mine > 0 && theirs > 0
	};

	const yourToken = panel.internNext[me];
	const legal =
		inPlace && turn === me && mine > 0
			? {
					byValue: legalByValue(me, panel),
					internToken: yourToken,
					internTargets: yourToken === null ? [] : internTargets(me, yourToken, panel)
				}
			: null;

	const confirmed = { pilot: false, copilot: false, ...s.rerollConfirmed };
	const ready = { pilot: false, copilot: false, ...s.ready };
	const mustAct = (role: Role): boolean => {
		if (!playing) return false;
		switch (step) {
			case 'discussion':
				return !ready[role];
			case 'place':
				return turn === role && count(dice[role]) > 0;
			case 'reroll':
				return !confirmed[role];
			case 'swap':
				return s.swap?.by !== role;
			case 'sync':
				return role === 'copilot';
		}
	};

	// Reroll tokens on the altitude track: two on the green and yellow side, one on red and black.
	const tokenRounds = info.colour === 'green' || info.colour === 'yellow' ? [1, 5] : [1];
	const view: FlightView = {
		round: s.round,
		firstPlayer: first,
		step,
		turn,
		approach: {
			spaces: s.planes.map((planes, i) => ({
				planes,
				traffic: s.track.traffic[i],
				turn: s.track.turns[i]
			})),
			position: s.position
		},
		rerollRounds: tokenRounds.filter((r) => r > s.round),
		axis: s.axis ?? 0,
		aero,
		gear,
		flaps,
		brakes,
		brakeLimit: ice === null ? [2, 3, 5, 7][brakes] : ice + 2,
		coffee: s.coffee ?? 0,
		rerolls: s.rerolls ?? 0,
		planeSupply: 12 - s.planes.reduce((a, b) => a + b, 0),
		fuel: s.fuel ?? null,
		leak: info.modules.includes('keroseneLeak'),
		wind,
		ice: ice === null ? null : { marker: ice },
		intern,
		placements,
		speed,
		roundStart: s.roundStart ?? { rerollPicked: false, traffic: [] },
		yourDice: [...dice[me]],
		abilityState,
		can,
		legal,
		reroll:
			step === 'reroll'
				? { confirmed: { pilot: confirmed.pilot, copilot: confirmed.copilot } }
				: null,
		swap: step === 'swap' ? (s.swap ?? null) : null,
		sync:
			step === 'sync' && s.sync !== undefined
				? { value: s.sync, targets: me === 'copilot' ? syncTargets(s.sync, panel) : [] }
				: null
	};

	const over = phase === 'won' || phase === 'lost';
	const players: PlayerSummary[] = SEATS.map((seat) => ({
		id: seat.id,
		name: seat.name,
		seat: seat.seat,
		role: seat.role,
		isHost: seat.id === 'ann',
		isYou: seat.role === me,
		mustAct: mustAct(seat.role),
		diceLeft: count(dice[seat.role]),
		ready: step === 'discussion' && playing ? ready[seat.role] : false,
		revealedDice: over ? dice[seat.role].filter((v): v is number => v !== null) : null
	}));

	return {
		gameId: 'K7MQ2',
		revision: 40 + s.round * 8 + placements.length,
		phase,
		you: idOf(me),
		hostId: 'ann',
		nextGameId: '',
		players,
		scenario: info,
		abilities,
		catalogue: null,
		flight: view,
		result: s.result ?? null
	};
}

const lost = (reason: LossReason, failed: Condition[] = []): PlayerView['result'] => ({
	won: false,
	reason,
	failed
});

// ------------------------------------------------------------------------- YUL green (base)

/** YUL green, round 3 (the Pilot places first): gear 3–4 and flaps 1–2 are in. */
const yulRound3 = (over: Partial<Spec> = {}): PlayerView =>
	flight({
		scenario: 'yul-green',
		round: 3,
		track: YUL,
		planes: [0, 0, 1, 2, 1, 3, 2],
		position: 3,
		gear: [true, true, false],
		flaps: 1,
		coffee: 1,
		rerolls: 1,
		placements: [die('pilot', 'gear34', 3), die('copilot', 'flaps12', 1)],
		dice: { pilot: [5, null, 6, 2], copilot: [4, 2, null, 6] },
		...over
	});

const yulTurnView = () => yulRound3();
const yulWaitingView = () =>
	yulRound3({
		placements: [
			die('pilot', 'gear34', 3),
			die('copilot', 'flaps12', 1),
			die('pilot', 'axisPilot', 5)
		],
		dice: { pilot: [null, null, 6, 2], copilot: [4, 2, null, 6] },
		turn: 'copilot'
	});
const rerollConfirmedView = () =>
	yulRound3({
		step: 'reroll',
		rerolls: 0,
		rerollConfirmed: { pilot: true }
	});

const yulFinalRoundView = () =>
	flight({
		scenario: 'yul-green',
		round: 7,
		track: YUL,
		planes: zeros(7),
		position: 7,
		gear: [true, true, true],
		flaps: 4,
		brakes: 2,
		coffee: 2,
		placements: [die('pilot', 'enginesPilot', 2), die('copilot', 'flaps56', 5)],
		dice: { pilot: [null, 4, 6, 1], copilot: [3, null, 5, 6] }
	});

// ------------------------------------------------------------------------------- discussion

const discussionView = () =>
	flight({
		scenario: 'yul-green',
		round: 5,
		step: 'discussion',
		track: YUL,
		planes: [0, 0, 0, 2, 1, 3, 2],
		position: 4,
		axis: 1,
		gear: [true, true, false],
		flaps: 2,
		coffee: 2,
		rerolls: 2,
		roundStart: { rerollPicked: true, traffic: [] }
	});

const discussionPartnerReadyView = () =>
	flight({
		scenario: 'yul-green',
		round: 2,
		step: 'discussion',
		track: YUL,
		planes: [0, 0, 1, 2, 1, 3, 2],
		position: 2,
		gear: [true, false, false],
		coffee: 1,
		rerolls: 1,
		ready: { copilot: true }
	});

/** TGU red, round 1: three traffic rolls put planes on the approach, and the reroll is picked up. */
const discussionTrafficView = () =>
	flight({
		scenario: 'tgu-red',
		abilities: ['synchronisation', 'workingTogether'],
		round: 1,
		step: 'discussion',
		track: TGU_RED,
		planes: [0, 2, 2, 2, 2],
		position: 1,
		rerolls: 2,
		fuel: 20,
		wind: 0,
		roundStart: {
			rerollPicked: true,
			traffic: [
				{ roll: 3, space: 3 },
				{ roll: 2, space: 2 },
				{ roll: 4, space: 4 }
			]
		}
	});

// ----------------------------------------------------------------------- TGU red (worst case)

/** TGU red, round 3, the Pilot's view: the black die went on radio 2, the Pilot is to place. */
const tguRound3 = (over: Partial<Spec> = {}): PlayerView =>
	flight({
		scenario: 'tgu-red',
		abilities: ['synchronisation', 'workingTogether'],
		round: 3,
		track: TGU_RED,
		planes: [0, 1, 2, 0, 2],
		position: 2,
		axis: -1,
		gear: [true, true, false],
		flaps: 2,
		brakes: 1,
		coffee: 2,
		rerolls: 1,
		fuel: 11,
		wind: -3,
		placements: [
			die('pilot', 'gear34', 3),
			die('copilot', 'flaps23', 2),
			black('radioCopilot2', 3),
			die('pilot', 'axisPilot', 4),
			die('copilot', 'radioCopilot1', 5)
		],
		dice: { pilot: [6, null, 2, null], copilot: [null, 3, null, 5] },
		abilityState: { syncUsed: true },
		...over
	});

const tguTurnView = () => tguRound3();
const tguWaitingView = () =>
	tguRound3({
		gear: [true, true, true],
		placements: [
			die('pilot', 'gear34', 3),
			die('copilot', 'flaps23', 2),
			black('radioCopilot2', 3),
			die('pilot', 'axisPilot', 4),
			die('copilot', 'radioCopilot1', 5),
			die('pilot', 'gear56', 6)
		],
		dice: { pilot: [null, null, 2, null], copilot: [null, 3, null, 5] },
		turn: 'copilot'
	});
const tguRerollView = () => tguRound3({ step: 'reroll', rerolls: 0 });

const swapOfferedView = () =>
	tguRound3({
		step: 'swap',
		swap: { by: 'pilot', value: 6, yourSlot: 0 },
		abilityState: { syncUsed: true, togetherUsed: true }
	});
const swapAnswerView = () =>
	tguRound3({
		me: 'copilot',
		step: 'swap',
		swap: { by: 'pilot', value: 6, yourSlot: null },
		abilityState: { syncUsed: true, togetherUsed: true }
	});

/** Gear 3–4 and flaps 2–3 are in: the black die (3) was rolled and the Co-Pilot places it. */
const syncRound = (me: Role): PlayerView =>
	tguRound3({
		me,
		step: 'sync',
		sync: 3,
		planes: [0, 1, 2, 1, 2],
		placements: [die('pilot', 'gear34', 3), die('copilot', 'flaps23', 2)],
		dice: { pilot: [6, null, 2, 5], copilot: [4, null, 5, 6] },
		turn: 'pilot'
	});
const syncPlaceView = () => syncRound('copilot');
const syncWaitingView = () => syncRound('pilot');

// ----------------------------------------------------------------------- KEF black (worst case)

const kefRound5 = (over: Partial<Spec> = {}): PlayerView =>
	flight({
		scenario: 'kef-black',
		abilities: ['adaptation', 'anticipation'],
		round: 5,
		track: KEF_BLACK,
		planes: [0, 0, 0, 0, 1, 0],
		position: 4,
		gear: [true, true, false],
		flaps: 2,
		ice: 1,
		coffee: 1,
		wind: 3,
		...over
	});

/** Round 5 before the first die: the Pilot, first player, may use Anticipation. */
const anticipationView = () => kefRound5({ dice: { pilot: [5, 3, 2, 6], copilot: [1, 4, 4, 6] } });

const kefTurnView = () =>
	kefRound5({
		axis: 1,
		wind: 4,
		placements: [
			die('pilot', 'axisPilot', 2),
			die('copilot', 'axisCopilot', 3),
			die('pilot', 'ice3Top', 3),
			die('copilot', 'enginesCopilot', 4)
		],
		dice: { pilot: [5, 3, null, null], copilot: [null, null, 4, 6] },
		abilityState: { anticipationUsed: true, adaptationUsed: { pilot: false, copilot: true } }
	});

// ----------------------------------------------------------------------- HND red (worst case)

/** HND red, round 6, the Co-Pilot's view: four intern tokens gone, the Pilot trained this round. */
const hndRound6 = (): PlayerView =>
	flight({
		scenario: 'hnd-red',
		abilities: ['control'],
		me: 'copilot',
		round: 6,
		track: HND_RED,
		planes: [0, 0, 0, 0, 0, 1, 2, 1],
		position: 6,
		axis: -1,
		gear: [true, true, true],
		flaps: 3,
		brakes: 2,
		coffee: 3,
		fuel: 9,
		intern: { values: [3, 5, 2, 6, 1, 4], taken: [0, 1, 4, 5] },
		placements: [
			die('copilot', 'flaps45', 4),
			die('pilot', 'internPilot', 1),
			token('pilot', 'gear56', 5),
			die('copilot', 'axisCopilot', 3),
			die('pilot', 'axisPilot', 3)
		],
		dice: { pilot: [null, null, 4, 5], copilot: [6, 2, null, null] },
		abilityState: { controlTriggered: true }
	});

const hndTurnView = hndRound6;
const hndTrainingView = hndRound6;

// ----------------------------------------------------------------------------- game over

const wonView = () =>
	flight({
		scenario: 'yul-green',
		phase: 'won',
		round: 7,
		track: YUL,
		planes: zeros(7),
		position: 7,
		gear: [true, true, true],
		flaps: 4,
		brakes: 2,
		coffee: 3,
		placements: [
			die('pilot', 'brakes4', 4),
			die('copilot', 'concentration1', 2),
			die('pilot', 'axisPilot', 3),
			die('copilot', 'axisCopilot', 3),
			die('pilot', 'enginesPilot', 1),
			die('copilot', 'enginesCopilot', 2),
			die('pilot', 'radioPilot', 6),
			die('copilot', 'radioCopilot1', 5)
		],
		result: { won: true, reason: null, failed: [] }
	});

const lostSpinView = () =>
	flight({
		scenario: 'yul-green',
		phase: 'lost',
		round: 4,
		track: YUL,
		planes: [0, 0, 1, 2, 1, 3, 2],
		position: 3,
		axis: -3,
		gear: [true, true, false],
		flaps: 1,
		coffee: 1,
		rerolls: 1,
		turn: 'copilot',
		placements: [die('copilot', 'axisCopilot', 2), die('pilot', 'axisPilot', 4)],
		dice: { pilot: [null, 5, 6, 1], copilot: [3, null, 5, 6] },
		result: lost('spin')
	});

const lostCollisionView = () =>
	flight({
		scenario: 'yul-green',
		phase: 'lost',
		round: 3,
		track: YUL,
		planes: [0, 0, 1, 2, 1, 3, 2],
		position: 3,
		gear: [true, false, false],
		coffee: 1,
		rerolls: 1,
		turn: 'pilot',
		placements: [die('pilot', 'enginesPilot', 3), die('copilot', 'enginesCopilot', 3)],
		dice: { pilot: [null, 2, 5, 6], copilot: [1, null, 4, 6] },
		result: lost('collision')
	});

/** Round 7 ends at the second engine die: speed 9 against a brake limit of 2. */
const lostLandingView = () =>
	flight({
		scenario: 'yul-green',
		phase: 'lost',
		round: 7,
		track: YUL,
		planes: [0, 0, 0, 0, 1, 0, 0],
		position: 7,
		gear: [true, true, false],
		flaps: 3,
		brakes: 1,
		coffee: 1,
		turn: 'pilot',
		placements: [
			die('pilot', 'axisPilot', 3),
			die('copilot', 'axisCopilot', 3),
			die('pilot', 'enginesPilot', 5),
			die('copilot', 'enginesCopilot', 4)
		],
		dice: { pilot: [null, 6, null, 2], copilot: [1, null, 3, null] },
		result: lost('landing', ['A', 'B', 'D'])
	});

/** TGU red, end of round 4: no die on the fuel space burned the last 6. */
const lostKeroseneView = () =>
	flight({
		scenario: 'tgu-red',
		abilities: ['adaptation', 'control'],
		phase: 'lost',
		round: 4,
		track: TGU_RED,
		planes: [0, 1, 2, 0, 2],
		position: 3,
		gear: [true, true, true],
		flaps: 3,
		brakes: 1,
		coffee: 3,
		fuel: -2,
		wind: -3,
		turn: 'pilot',
		placements: [
			die('copilot', 'flaps45', 4),
			die('pilot', 'gear56', 5),
			die('copilot', 'axisCopilot', 3),
			die('pilot', 'axisPilot', 3),
			die('copilot', 'enginesCopilot', 2),
			die('pilot', 'enginesPilot', 2),
			die('copilot', 'radioCopilot1', 1),
			die('pilot', 'radioPilot', 6)
		],
		abilityState: { controlTriggered: true },
		result: lost('kerosene')
	});

/** HND red: everything else is in order, but two intern tokens were never used. */
const lostInternView = () =>
	flight({
		scenario: 'hnd-red',
		abilities: ['control'],
		phase: 'lost',
		round: 7,
		track: HND_RED,
		planes: zeros(8),
		position: 8,
		gear: [true, true, true],
		flaps: 4,
		brakes: 3,
		coffee: 3,
		fuel: 9,
		intern: { values: [3, 5, 2, 6, 1, 4], taken: [0, 1, 4, 5] },
		turn: 'pilot',
		placements: [
			die('pilot', 'axisPilot', 3),
			die('copilot', 'axisCopilot', 3),
			die('pilot', 'enginesPilot', 1),
			die('copilot', 'enginesCopilot', 2),
			die('pilot', 'radioPilot', 2),
			die('copilot', 'radioCopilot1', 4),
			die('pilot', 'kerosene', 4),
			die('copilot', 'concentration1', 5)
		],
		abilityState: { controlTriggered: true },
		result: lost('intern', ['intern'])
	});

// -------------------------------------------------------------------------------- registry

/**
 * Every named fixture; each call returns a fresh view. The dev route `/dev/fixtures/<name>`
 * shows them on the game screen.
 */
export const fixtures: Record<string, () => PlayerView> = {
	lobbyHost: lobbyHostView,
	lobbyHostAbilities: lobbyHostAbilitiesView,
	lobbyGuest: lobbyGuestView,
	lobbyAlone: lobbyAloneView,
	discussion: discussionView,
	discussionTraffic: discussionTrafficView,
	discussionPartnerReady: discussionPartnerReadyView,
	yulTurn: yulTurnView,
	yulWaiting: yulWaitingView,
	yulFinalRound: yulFinalRoundView,
	anticipation: anticipationView,
	tguTurn: tguTurnView,
	tguWaiting: tguWaitingView,
	tguReroll: tguRerollView,
	kefTurn: kefTurnView,
	hndTurn: hndTurnView,
	hndTraining: hndTrainingView,
	rerollConfirmed: rerollConfirmedView,
	swapOffered: swapOfferedView,
	swapAnswer: swapAnswerView,
	syncPlace: syncPlaceView,
	syncWaiting: syncWaitingView,
	won: wonView,
	lostSpin: lostSpinView,
	lostCollision: lostCollisionView,
	lostLanding: lostLandingView,
	lostKerosene: lostKeroseneView,
	lostIntern: lostInternView,
	ended: endedView
};

/** The table's UI state each fixture starts with: a die picked, a coffee modifier, a space. */
export const presets: Record<string, TablePreset> = {
	yulTurn: { slot: 0 },
	anticipation: { slot: 1 },
	tguTurn: { slot: 0 },
	tguReroll: { rerollSlots: [2] },
	kefTurn: { slot: 0 },
	hndTurn: { slot: 1, coffee: -1 },
	hndTraining: { slot: 1, space: 'internCopilot' },
	swapAnswer: { slot: 1 }
};
