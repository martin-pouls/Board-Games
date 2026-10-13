// Sample views for tests and UI previews. Not used by the running app.
//
// The views come from a tiny deterministic table simulation (deal, legal cards, tricks, JARVIS's
// stacks, mission 12's draw), so card counts, tricks won, hands and the trick shown always agree.
// The seeds are fixed; every function returns a fresh object.
import { cardKey, sameCard } from './cards.js';
import type {
	Actions,
	Attempt,
	AttemptResult,
	Card,
	CommOption,
	Communication,
	CompletedTrick,
	Exchange,
	LogEntry,
	Member,
	MissionDef,
	Objective,
	ObjectiveKind,
	ObjectiveSpec,
	ObjectiveView,
	OwnerRule,
	Pass,
	PickKind,
	PlayerSummary,
	PlayerView,
	Play,
	Poll,
	Position,
	Stack,
	Status,
	Step,
	Suit,
	Task,
	Token
} from '#lib/types.js';
import { OWNER_CREW, OWNER_UNASSIGNED, LAST_TRICK } from '#lib/types.js';

// ---------------------------------------------------------------------------------- cards

export function card(suit: Suit, rank: number): Card {
	return { suit, rank };
}

const COLOURS: Suit[] = ['pink', 'blue', 'green', 'yellow'];
const SUIT_ORDER: Suit[] = [...COLOURS, 'trump'];

/** The 40 cards: four colours 1–9 and the rockets 1–4. */
export function fullDeck(): Card[] {
	const cards: Card[] = [];
	for (const suit of COLOURS) for (let rank = 1; rank <= 9; rank++) cards.push(card(suit, rank));
	for (let rank = 1; rank <= 4; rank++) cards.push(card('trump', rank));
	return cards;
}

/** Colours in the order pink, blue, green, yellow, rockets; ascending within a colour. */
function sorted(cards: Card[]): Card[] {
	return [...cards].sort(
		(a, b) => SUIT_ORDER.indexOf(a.suit) - SUIT_ORDER.indexOf(b.suit) || a.rank - b.rank
	);
}

const ROCKET_4 = card('trump', 4);

// ---------------------------------------------------------------------------------- builders

const NAMES = ['Alice', 'Bob', 'Cleo', 'Dan', 'Eva'];
const idOf = (seat: number) => NAMES[seat].toLowerCase();

/** A human member at a seat. Override what the test needs. */
export function member(seat: number, over: Partial<Member> = {}): Member {
	return {
		seat,
		playerId: idOf(seat),
		name: NAMES[seat],
		dummy: false,
		cardCount: 0,
		tricksWon: 0,
		communication: null,
		communicated: false,
		stacks: [],
		mustAct: false,
		cards: null,
		...over
	};
}

/** The summaries of `n` players (Alice is the host), seen by the player at `viewer`. */
export function playerList(n: number, viewer = 0): PlayerSummary[] {
	return NAMES.slice(0, n).map((name, seat) => ({
		id: idOf(seat),
		name,
		seat,
		isHost: seat === 0,
		isYou: seat === viewer
	}));
}

/** The logbook of a crew that has finished missions 1–3 and failed mission 4 twice. */
export function logbook(extra: LogEntry[] = []): LogEntry[] {
	const base: LogEntry[] = [
		{ mission: 1, attempts: 1, distress: false, completed: true, total: 1 },
		{ mission: 2, attempts: 3, distress: false, completed: true, total: 3 },
		{ mission: 3, attempts: 2, distress: true, completed: true, total: 3 },
		{ mission: 4, attempts: 2, distress: false, completed: false, total: 2 }
	];
	const merged = [...base.filter((e) => !extra.some((x) => x.mission === e.mission)), ...extra];
	return merged.sort((a, b) => a.mission - b.mission);
}

/** A lobby view of `n` players at mission 4. Override what the test needs. */
export function baseView(n = 1, viewer = 0, over: Partial<PlayerView> = {}): PlayerView {
	return {
		gameId: 'CREW7',
		revision: 5,
		phase: 'lobby',
		you: idOf(viewer),
		hostId: idOf(0),
		players: playerList(n, viewer),
		edition: 'planetNine',
		options: { distress: false },
		mission: 4,
		missionDef: missionDef(4),
		logbook: logbook(),
		attempt: null,
		...over
	};
}

/** An objective with every field filled in; only the ones its kind uses matter. */
export function objective(kind: ObjectiveKind, over: Partial<Objective> = {}): Objective {
	return {
		kind,
		owner: OWNER_CREW,
		cards: [],
		count: 0,
		tricks: [],
		suit: '',
		rank: 0,
		notWithTrump: false,
		ordered: false,
		...over
	};
}

function objectiveView(o: Objective, status: Status = 'open'): ObjectiveView {
	return { ...o, status };
}

function spec(
	kind: ObjectiveKind,
	ownerRule: OwnerRule,
	over: Partial<ObjectiveSpec> = {}
): ObjectiveSpec {
	return { ...objective(kind), ownerRule, holder: null, ...over };
}

function def(number: number, over: Partial<MissionDef> = {}): MissionDef {
	return {
		number,
		tasks: 0,
		tokens: [],
		distribution: '',
		deadZone: false,
		fromTrick: 0,
		golden: false,
		setup: [],
		poll: '',
		objectives: [],
		...over
	};
}

const open = (number: number, tasks: number, over: Partial<MissionDef> = {}) =>
	def(number, { tasks, distribution: 'open', ...over });

/** The few mission rows the fixtures use, as in rules.md §14. */
const DEFS: Record<number, () => MissionDef> = {
	4: () => open(4, 3),
	5: () =>
		def(5, {
			poll: 'wellIll',
			objectives: [spec('winNoTricks', 'chosen', { owner: OWNER_UNASSIGNED })]
		}),
	6: () => open(6, 3, { tokens: ['arrow1', 'arrow2'], deadZone: true }),
	10: () => open(10, 4),
	12: () => open(12, 4, { tokens: ['omega'], setup: ['drawAfterFirstTrick'] }),
	18: () => open(18, 5, { fromTrick: 2 }),
	22: () => open(22, 5, { tokens: ['arrow1', 'arrow2', 'arrow3', 'arrow4'] }),
	23: () => open(23, 5, { tokens: ['1', '2', '3', '4', '5'], setup: ['swapTokens'] }),
	31: () => open(31, 6, { tokens: ['1', '2', '3'], golden: true }),
	48: () =>
		open(48, 3, {
			tokens: ['omega'],
			golden: true,
			objectives: [spec('cardInLastTrick', 'crew', { taskToken: 'omega' })]
		}),
	50: () =>
		def(50, {
			objectives: [
				spec('winOnlyTricks', 'roleA', { tricks: [1, 2, 3, 4] }),
				spec('winOnlyTricks', 'roleB', { tricks: [LAST_TRICK] })
			]
		})
};

/** The mission numbers `missionDef` knows. */
export const FIXTURE_MISSIONS = Object.keys(DEFS).map(Number);

/** A mission row (a fresh copy). Only the missions the fixtures use exist. */
export function missionDef(n: number): MissionDef {
	const make = DEFS[n];
	if (!make) throw new Error(`fixtures: no mission row for ${n}`);
	return make();
}

// ---------------------------------------------------------------------------------- table simulation

type Rand = () => number;

/** mulberry32: a small seeded generator. */
function seeded(seed: number): Rand {
	let a = seed >>> 0;
	return () => {
		a = (a + 0x6d2b79f5) >>> 0;
		let t = a;
		t = Math.imul(t ^ (t >>> 15), t | 1);
		t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
		return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
	};
}

function shuffled<T>(xs: readonly T[], rand: Rand): T[] {
	const out = [...xs];
	for (let i = out.length - 1; i > 0; i--) {
		const j = Math.floor(rand() * (i + 1));
		[out[i], out[j]] = [out[j], out[i]];
	}
	return out;
}

function pick<T>(xs: readonly T[], rand: Rand): T {
	return xs[Math.floor(rand() * xs.length)];
}

interface Pile {
	up: Card | null;
	down: Card | null;
}

interface Table {
	/** Humans at the table. */
	humans: number;
	/** Seats: the humans, plus JARVIS (seat 2) with two humans. */
	seats: number;
	jarvis: boolean;
	/** Per seat; empty for JARVIS. */
	hands: Card[][];
	/** JARVIS's seven stacks. */
	piles: Pile[];
	commander: number;
	total: number;
	tricks: CompletedTrick[];
	trick: Play[];
	leader: number;
	drawAfterFirstTrick: boolean;
	/** Mission 12: per seat, the card it gave away and the card it got. */
	gave: (Card | null)[];
	got: (Card | null)[];
	rand: Rand;
}

const TOTAL_TRICKS: Record<number, number> = { 3: 13, 4: 10, 5: 8 };

const isDummy = (t: Table, seat: number) => t.jarvis && seat === 2;
const nameOf = (t: Table, seat: number) => (isDummy(t, seat) ? '' : NAMES[seat]);
const turnOf = (t: Table) => (t.leader + t.trick.length) % t.seats;

/** The seat that makes the decisions for `seat`: the commander for JARVIS. */
const actorFor = (t: Table, seat: number) => (isDummy(t, seat) ? t.commander : seat);

function deal(humans: number, rand: Rand): Table {
	const jarvis = humans === 2;
	const seats = jarvis ? 3 : humans;
	const hands: Card[][] = Array.from({ length: seats }, () => []);
	const piles: Pile[] = [];
	if (jarvis) {
		// JARVIS: seven face-down cards with a face-up card on each; the rocket 4 goes to the humans.
		const rest = shuffled(
			fullDeck().filter((c) => !sameCard(c, ROCKET_4)),
			rand
		);
		const theirs = rest.splice(0, 14);
		for (let i = 0; i < 7; i++) piles.push({ down: theirs[i], up: theirs[7 + i] });
		const cards = shuffled([...rest, ROCKET_4], rand);
		hands[0] = cards.slice(0, 13);
		hands[1] = cards.slice(13);
	} else {
		const cards = shuffled(fullDeck(), rand);
		const each = Math.floor(40 / seats);
		for (let s = 0; s < seats; s++) hands[s] = cards.slice(s * each, (s + 1) * each);
		// 3 players: the 40th card goes to a random seat.
		for (const c of cards.slice(seats * each)) hands[Math.floor(rand() * seats)].push(c);
	}
	const holder = hands.findIndex((h) => h.some((c) => sameCard(c, ROCKET_4)));
	return {
		humans,
		seats,
		jarvis,
		hands: hands.map(sorted),
		piles,
		commander: holder,
		total: TOTAL_TRICKS[seats],
		tricks: [],
		trick: [],
		leader: holder,
		drawAfterFirstTrick: false,
		gave: Array.from({ length: seats }, () => null),
		got: Array.from({ length: seats }, () => null),
		rand
	};
}

/** Moves the rocket 4 into the hand of `seat`, who becomes the commander and leads. */
function setCommander(t: Table, seat: number) {
	const holder = t.hands.findIndex((h) => h.some((c) => sameCard(c, ROCKET_4)));
	if (holder !== seat) {
		const swapped = t.hands[seat][0];
		t.hands[holder] = sorted(t.hands[holder].map((c) => (sameCard(c, ROCKET_4) ? swapped : c)));
		t.hands[seat] = sorted(t.hands[seat].map((c) => (c === swapped ? ROCKET_4 : c)));
	}
	t.commander = seat;
	t.leader = seat;
}

/** Cards still held by a seat; for JARVIS both layers of its stacks. */
function held(t: Table, seat: number): Card[] {
	if (!isDummy(t, seat)) return t.hands[seat];
	return t.piles.flatMap((p) => [p.up, p.down].filter((c): c is Card => c !== null));
}

/** What the seat may play now: its hand, or JARVIS's face-up cards; follow suit applies. */
function legal(t: Table, seat: number): Card[] {
	const pool = isDummy(t, seat) ? t.piles.flatMap((p) => (p.up ? [p.up] : [])) : t.hands[seat];
	if (t.trick.length === 0) return pool;
	const led = t.trick[0].card.suit;
	const follow = pool.filter((c) => c.suit === led);
	return follow.length ? follow : pool;
}

function trickWinner(plays: Play[]): number {
	const led = plays[0].card.suit;
	const rockets = plays.filter((p) => p.card.suit === 'trump');
	const pool = rockets.length ? rockets : plays.filter((p) => p.card.suit === led);
	return pool.reduce((best, p) => (p.card.rank > best.card.rank ? p : best)).seat;
}

/** Mission 12: everybody takes a random card from the right neighbour; decided first, then moved. */
function drawCards(t: Table) {
	const humans = [...Array(t.seats).keys()].filter((s) => !isDummy(t, s));
	const right = (s: number) => humans[(humans.indexOf(s) - 1 + humans.length) % humans.length];
	const taken = new Map<number, Card>();
	for (const s of humans) taken.set(s, pick(t.hands[right(s)], t.rand));
	for (const s of humans) {
		const c = taken.get(s)!;
		t.got[s] = c;
		t.gave[right(s)] = c;
	}
	for (const s of humans) {
		const lost = t.gave[s]!;
		t.hands[s] = sorted([...t.hands[s].filter((c) => !sameCard(c, lost)), taken.get(s)!]);
	}
}

function finishTrick(t: Table) {
	const winner = trickWinner(t.trick);
	t.tricks.push({ leader: t.leader, plays: t.trick, winner });
	t.trick = [];
	t.leader = winner;
	// The card under a played JARVIS card is turned face up after the trick.
	for (const p of t.piles) {
		if (!p.up && p.down) [p.up, p.down] = [p.down, null];
	}
	if (t.drawAfterFirstTrick && t.tricks.length === 1) drawCards(t);
}

function playCard(t: Table, seat: number, c: Card) {
	if (isDummy(t, seat)) {
		t.piles.find((p) => sameCard(p.up, c))!.up = null;
	} else {
		t.hands[seat] = t.hands[seat].filter((h) => !sameCard(h, c));
	}
	t.trick.push({ seat, card: c });
	if (t.trick.length === t.seats) finishTrick(t);
}

/** Plays random legal cards until `stop`; false if the attempt ends first. */
function run(t: Table, stop: () => boolean): boolean {
	while (!stop()) {
		if (t.tricks.length >= t.total) return false;
		const seat = turnOf(t);
		playCard(t, seat, pick(legal(t, seat), t.rand));
	}
	return true;
}

interface TableSpec {
	humans: number;
	/** The holder of the rocket 4; random if omitted. */
	commander?: number;
	seed: number;
	/** Tricks to play out first (at least; exactly, if no `turn`). */
	tricks?: number;
	/** Cards in the trick in progress (no `turn`). */
	plays?: number;
	/** Then play on until it is this seat's turn with at least `minPlays` cards in the trick. */
	turn?: number;
	minPlays?: number;
	/** Mission 12. */
	draw?: boolean;
	/** Looks for a seed (from `seed` on) that satisfies this. */
	accept?: (t: Table) => boolean;
}

function table(s: TableSpec): Table {
	for (let seed = s.seed; seed < s.seed + 2000; seed++) {
		const t = deal(s.humans, seeded(seed));
		if (s.commander !== undefined) setCommander(t, s.commander);
		t.drawAfterFirstTrick = s.draw ?? false;
		if (!run(t, () => t.tricks.length >= (s.tricks ?? 0))) continue;
		const arrived =
			s.turn === undefined
				? run(t, () => t.trick.length >= (s.plays ?? 0))
				: run(t, () => turnOf(t) === s.turn && t.trick.length >= (s.minPlays ?? 0));
		if (arrived && (s.accept?.(t) ?? true)) return t;
	}
	throw new Error('fixtures: no seed satisfies the table spec');
}

// ---------------------------------------------------------------------------------- views

/** The true statement about a card in a hand, as in the domain (`StatementFor`). */
function statement(hand: Card[], c: Card): Position | null {
	if (c.suit === 'trump') return null;
	const same = hand.filter((h) => h.suit === c.suit);
	if (same.length === 1) return 'only';
	if (c.rank === Math.max(...same.map((h) => h.rank))) return 'highest';
	if (c.rank === Math.min(...same.map((h) => h.rank))) return 'lowest';
	return null;
}

function commOptions(hand: Card[], deadZone: boolean): CommOption[] {
	const out: CommOption[] = [];
	for (const c of hand) {
		const position = statement(hand, c);
		if (position) out.push({ card: { ...c }, position: deadZone ? '' : position });
	}
	return out;
}

/** A communication already made by `seat`: its first communicable card. */
function communicationOf(t: Table, seat: number, deadZone: boolean, trick: number): Communication {
	const [first] = commOptions(t.hands[seat], deadZone);
	return { card: first.card, position: first.position, trick };
}

interface TaskOpts {
	count: number;
	/** Per task, in order; the others have none. */
	tokens?: Token[];
	/** Fulfilled tasks: cards of completed tricks, owned by the winner. */
	fulfilled?: number;
	/** One failed task (the newest trick's card, owned by someone else). */
	failed?: boolean;
	/** Open tasks that have an owner, in draft order from the commander; the rest are unassigned. */
	owned?: number;
}

function makeTasks(t: Table, o: TaskOpts): Task[] {
	const tasks: Task[] = [];
	const used = new Set<string>();
	const add = (c: Card, owner: number, status: Status) => {
		used.add(cardKey(c));
		tasks.push({
			id: tasks.length,
			card: { ...c },
			token: o.tokens?.[tasks.length] ?? '',
			owner,
			status,
			faceDown: false
		});
	};
	let k = t.tricks.length - 1;
	const fromTrick = (): { card: Card; winner: number } => {
		for (; k >= 0; k--) {
			const tr = t.tricks[k];
			const p = tr.plays.find((p) => p.card.suit !== 'trump' && !used.has(cardKey(p.card)));
			if (p) {
				k--;
				return { card: p.card, winner: tr.winner };
			}
		}
		throw new Error('fixtures: no completed trick with a colour card left for a task');
	};
	if (o.failed) {
		const { card: c, winner } = fromTrick();
		add(c, (winner + 1) % t.seats, 'failed');
	}
	for (let i = 0; i < (o.fulfilled ?? 0); i++) {
		const { card: c, winner } = fromTrick();
		add(c, winner, 'fulfilled');
	}
	const pool = shuffled(
		[...Array(t.seats).keys()]
			.flatMap((s) => held(t, s))
			.filter((c) => c.suit !== 'trump' && !used.has(cardKey(c))),
		t.rand
	);
	for (let i = 0; tasks.length < o.count; i++) {
		const owner = i < (o.owned ?? o.count) ? (t.commander + i) % t.seats : OWNER_UNASSIGNED;
		add(pool.pop()!, owner, 'open');
	}
	return tasks;
}

/** Whether the newest `k` completed tricks each have a colour card (for tasks). */
function hasColourTricks(t: Table, k: number): boolean {
	const recent = t.tricks.slice(-k);
	return recent.length === k && recent.every((tr) => tr.plays.some((p) => p.card.suit !== 'trump'));
}

interface AttemptOpts {
	viewer: number;
	mission: number;
	step?: Step;
	pick?: PickKind;
	number?: number;
	tasks?: Task[];
	objectives?: ObjectiveView[];
	poll?: Poll | null;
	pass?: Pass | null;
	exchange?: Exchange | null;
	/** Draft: the seat to take a task. */
	draftTurn?: number;
	actors?: number[];
	silenced?: number;
	pickable?: number[];
	/** Communications made, by seat. */
	comms?: Record<number, Communication>;
	result?: AttemptResult | null;
	distress?: boolean;
}

/** Who the attempt waits for in a step. */
function defaultActors(t: Table, step: Step, turn: number, o: AttemptOpts): number[] {
	switch (step) {
		case 'adjustTokens':
		case 'pick':
			return [t.commander];
		case 'poll':
			return (o.poll?.answers ?? []).flatMap((a, s) =>
				a === null && s !== t.commander && !isDummy(t, s) ? [s] : []
			);
		case 'pass':
			return o.pass?.waiting ?? [];
		case 'draft':
		case 'play':
			return turn >= 0 ? [actorFor(t, turn)] : [];
		case '':
			return [];
	}
}

function attemptOf(t: Table, o: AttemptOpts): Attempt {
	const m = missionDef(o.mission);
	const step: Step = o.step ?? (o.result ? '' : 'play');
	const over = !!o.result;
	const turn = step === 'play' ? turnOf(t) : step === 'draft' ? (o.draftTurn ?? -1) : -1;
	const actors = o.actors ?? defaultActors(t, step, turn, o);
	const hand = t.hands[o.viewer];
	const prePlay = step === 'play' && t.tricks.length === 0 && t.trick.length === 0;
	const comms = o.comms ?? {};

	const members: Member[] = [...Array(t.seats).keys()].map((seat) => {
		const dummy = isDummy(t, seat);
		const stacks: Stack[] = dummy
			? t.piles.map((p) => ({ up: p.up ? { ...p.up } : null, hidden: p.down !== null }))
			: [];
		return member(seat, {
			playerId: dummy ? '' : idOf(seat),
			name: nameOf(t, seat),
			dummy,
			cardCount: held(t, seat).length,
			tricksWon: t.tricks.filter((tr) => tr.winner === seat).length,
			communication: comms[seat] ?? null,
			communicated: seat in comms,
			stacks,
			mustAct: actors.includes(seat),
			cards: over ? sorted(held(t, seat)) : null
		});
	});

	const actions: Actions = { communicate: [], sendDistress: false, handover: false };
	if (!over && step === 'play') {
		const canCommunicate =
			t.trick.length === 0 &&
			!(o.viewer in comms) &&
			o.silenced !== o.viewer &&
			t.tricks.length + 1 >= m.fromTrick;
		if (canCommunicate) actions.communicate = commOptions(hand, m.deadZone);
		const commander = o.viewer === t.commander;
		actions.sendDistress = commander && prePlay && Object.keys(comms).length === 0;
		actions.handover = commander && prePlay && t.seats === 5 && m.golden;
	}

	// The last completed trick stays visible during the next one (rules §16).
	const lastTrick = t.tricks.length > 0 ? { ...t.tricks[t.tricks.length - 1] } : null;

	return {
		number: o.number ?? 3,
		seat: o.viewer,
		commander: t.commander,
		step,
		pick: o.pick ?? '',
		actors,
		turn,
		trickNumber: t.tricks.length + 1,
		totalTricks: t.total,
		members,
		hand: hand.map((c) => ({ ...c })),
		playable:
			step === 'play' && !over && turn >= 0 && actorFor(t, turn) === o.viewer
				? legal(t, turn).map((c) => ({ ...c }))
				: [],
		tasks: o.tasks ?? [],
		objectives: o.objectives ?? [],
		silenced: o.silenced ?? -1,
		pickable: o.pickable ?? [],
		poll: o.poll ?? null,
		pass: o.pass ?? null,
		exchange: o.exchange ?? null,
		trick: t.trick.map((p) => ({ seat: p.seat, card: { ...p.card } })),
		trickLeader: t.leader,
		lastTrick: lastTrick && {
			...lastTrick,
			plays: lastTrick.plays.map((p) => ({ seat: p.seat, card: { ...p.card } }))
		},
		distress: o.distress ?? false,
		actions,
		result: o.result ?? null
	};
}

/** A running attempt or its debrief, as seen by `o.viewer`. */
function gameView(t: Table, o: AttemptOpts): PlayerView {
	const attempt = attemptOf(t, o);
	const won = !!o.result?.success;
	const entry: LogEntry = {
		mission: o.mission,
		attempts: attempt.number,
		distress: attempt.distress,
		completed: won,
		total: attempt.number + (attempt.distress ? 1 : 0)
	};
	return baseView(t.humans, o.viewer, {
		phase: o.result ? 'debrief' : 'playing',
		mission: o.mission,
		missionDef: missionDef(o.mission),
		logbook: logbook([entry]),
		attempt
	});
}

// ---------------------------------------------------------------------------------- lobby

/** The lobby with `n` players and a logbook of a few entries; mission 4 is selected. */
export function lobbyView(n: number, viewer = 0): PlayerView {
	return baseView(n, viewer);
}
export const lobbyView1 = () => lobbyView(1);
export const lobbyView2 = () => lobbyView(2);
/** Seen by a guest (Cleo), who cannot pick the mission. */
export const lobbyView4 = () => lobbyView(4, 2);

// ---------------------------------------------------------------------------------- 2 players + JARVIS

/** Mission 4, draft: the viewer is the commander and takes the first task. */
export function draftView2(): PlayerView {
	const t = table({ humans: 2, commander: 0, seed: 11 });
	return gameView(t, {
		viewer: 0,
		mission: 4,
		step: 'draft',
		draftTurn: 0,
		tasks: makeTasks(t, { count: 3, owned: 0 })
	});
}

/** Mission 5, poll: Bob is the commander and asks the viewer whether they are well or ill. */
export function pollView2(): PlayerView {
	const t = table({ humans: 2, commander: 1, seed: 12 });
	return gameView(t, {
		viewer: 0,
		mission: 5,
		step: 'poll',
		objectives: [objectiveView(objective('winNoTricks', { owner: OWNER_UNASSIGNED }))],
		poll: { kind: 'wellIll', answers: [null, null, null], objective: 0, task: -1 }
	});
}

/** Mission 5, pick: Bob said he is well; the viewer, the commander, chooses who wins no trick. */
export function pickView2(): PlayerView {
	const t = table({ humans: 2, commander: 0, seed: 13 });
	return gameView(t, {
		viewer: 0,
		mission: 5,
		step: 'pick',
		pick: 'objective',
		pickable: [1, 2],
		objectives: [objectiveView(objective('winNoTricks', { owner: OWNER_UNASSIGNED }))],
		poll: { kind: 'wellIll', answers: [null, true, null], objective: 0, task: -1 }
	});
}

/** Mission 23: the commander (the viewer) may swap the tokens of two tasks. */
export function adjustTokensView2(): PlayerView {
	const t = table({ humans: 2, commander: 0, seed: 14 });
	return gameView(t, {
		viewer: 0,
		mission: 23,
		step: 'adjustTokens',
		tasks: makeTasks(t, { count: 5, tokens: ['1', '2', '3', '4', '5'], owned: 0 })
	});
}

/** Mission 10, distress signal sent to the left: both players choose a card to pass. */
export function passView2(): PlayerView {
	const t = table({ humans: 2, commander: 0, seed: 15 });
	return gameView(t, {
		viewer: 0,
		mission: 10,
		step: 'pass',
		distress: true,
		tasks: makeTasks(t, { count: 4 }),
		pass: { direction: 'left', chosen: null, waiting: [0, 1] }
	});
}

/** The same pass after the viewer chose; Bob has not. */
export function passChosenView2(): PlayerView {
	const t = table({ humans: 2, commander: 0, seed: 15 });
	const chosen = t.hands[0].find((c) => c.suit !== 'trump')!;
	return gameView(t, {
		viewer: 0,
		mission: 10,
		step: 'pass',
		distress: true,
		tasks: makeTasks(t, { count: 4 }),
		pass: { direction: 'left', chosen: { ...chosen }, waiting: [1] }
	});
}

/**
 * Mission 10, trick 4 or later: JARVIS has played, Bob has communicated, and it is the viewer's
 * turn. The playable cards (must follow suit) are a strict subset of the hand.
 */
export function playView2(): PlayerView {
	const t = table({
		humans: 2,
		commander: 0,
		seed: 21,
		tricks: 3,
		turn: 0,
		minPlays: 1,
		accept: (t) => legal(t, 0).length >= 1 && legal(t, 0).length < t.hands[0].length
	});
	return gameView(t, {
		viewer: 0,
		mission: 10,
		tasks: makeTasks(t, { count: 4, fulfilled: 1 }),
		comms: { 1: communicationOf(t, 1, false, 2) }
	});
}

/** Mission 10: it is JARVIS's turn and the viewer, the commander, picks one of its face-up cards. */
export function playJarvisView2(): PlayerView {
	const t = table({
		humans: 2,
		commander: 0,
		seed: 22,
		tricks: 4,
		turn: 2,
		minPlays: 1,
		accept: (t) => {
			const ups = t.piles.filter((p) => p.up).length;
			return legal(t, 2).length >= 1 && legal(t, 2).length < ups;
		}
	});
	return gameView(t, {
		viewer: 0,
		mission: 10,
		tasks: makeTasks(t, { count: 4, fulfilled: 1 })
	});
}

/** Mission 10 won in trick 9: all four tasks fulfilled, the hands revealed. */
export function debriefWonView2(): PlayerView {
	const t = table({
		humans: 2,
		commander: 0,
		seed: 31,
		tricks: 9,
		accept: (t) => hasColourTricks(t, 4)
	});
	return gameView(t, {
		viewer: 0,
		mission: 10,
		step: '',
		number: 3,
		tasks: makeTasks(t, { count: 4, fulfilled: 4 }),
		result: { success: true, reason: 'completed', task: -1, objective: -1, trick: -1, seat: -1 }
	});
}

/** Mission 10 lost in trick 6: the first task was won by the wrong member. */
export function debriefLostView2(): PlayerView {
	const t = table({
		humans: 2,
		commander: 0,
		seed: 32,
		tricks: 6,
		accept: (t) => hasColourTricks(t, 2)
	});
	const tasks = makeTasks(t, { count: 4, failed: true, fulfilled: 1 });
	return gameView(t, {
		viewer: 0,
		mission: 10,
		step: '',
		number: 3,
		tasks,
		result: {
			success: false,
			reason: 'taskFailed',
			task: 0,
			objective: -1,
			trick: t.tricks.length,
			seat: t.tricks[t.tricks.length - 1].winner
		}
	});
}

// ---------------------------------------------------------------------------------- 3 players

/**
 * Mission 6 (dead zone) after trick 4: the last trick is shown, Bob has communicated a card
 * without a position, and the viewer may communicate now.
 */
export function playView3(): PlayerView {
	const t = table({ humans: 3, commander: 1, seed: 41, tricks: 4, plays: 0 });
	return gameView(t, {
		viewer: 0,
		mission: 6,
		tasks: makeTasks(t, { count: 3, tokens: ['arrow1', 'arrow2'] }),
		comms: { 1: communicationOf(t, 1, true, 3) }
	});
}

// ---------------------------------------------------------------------------------- 5 players

/** Mission 18, draft: Alice and Bob took a task; the viewer (Cleo) is next. */
export function draftView5(): PlayerView {
	const t = table({ humans: 5, commander: 0, seed: 51 });
	return gameView(t, {
		viewer: 2,
		mission: 18,
		step: 'draft',
		draftTurn: 2,
		tasks: makeTasks(t, { count: 5, owned: 2 })
	});
}

/** Mission 22, trick 4 (of 8): the viewer (Dan) is to play. */
export function playView5(): PlayerView {
	const t = table({ humans: 5, commander: 0, seed: 52, tricks: 3, turn: 3, minPlays: 1 });
	return gameView(t, {
		viewer: 3,
		mission: 22,
		tasks: makeTasks(t, {
			count: 5,
			tokens: ['arrow1', 'arrow2', 'arrow3', 'arrow4'],
			fulfilled: 1
		})
	});
}

/** Mission 31 (golden frame) before trick 1: the commander may hand a task to another member. */
export function handoverView5(): PlayerView {
	const t = table({ humans: 5, commander: 0, seed: 53 });
	return gameView(t, {
		viewer: 0,
		mission: 31,
		tasks: makeTasks(t, { count: 6, tokens: ['1', '2', '3'] })
	});
}

// ---------------------------------------------------------------------------------- special missions

/** Mission 12 right after the first trick and the draw: the viewer gave one card and got one. */
export function mission12View(): PlayerView {
	const t = table({ humans: 4, commander: 1, seed: 61, tricks: 1, draw: true });
	return gameView(t, {
		viewer: 0,
		mission: 12,
		tasks: makeTasks(t, { count: 4, tokens: ['', '', '', 'omega'] }),
		exchange: { kind: 'draw', gave: { ...t.gave[0]! }, got: { ...t.got[0]! } }
	});
}

/** Mission 50 after trick 2: Bob (A) won tricks 1 and 2 as he should; Cleo (B) waits for the last. */
export function mission50View(): PlayerView {
	const t = table({
		humans: 3,
		commander: 0,
		seed: 71,
		tricks: 2,
		plays: 0,
		accept: (t) => t.tricks.every((tr) => tr.winner === 1)
	});
	return gameView(t, {
		viewer: 0,
		mission: 50,
		objectives: [
			objectiveView(objective('winOnlyTricks', { owner: 1, tricks: [1, 2, 3, 4] })),
			objectiveView(objective('winOnlyTricks', { owner: 2, tricks: [LAST_TRICK] }))
		]
	});
}

/** Mission 48 after trick 2: the Ω task must be won in the last trick. */
export function mission48View(): PlayerView {
	const t = table({ humans: 4, commander: 2, seed: 81, tricks: 2, plays: 1 });
	const tasks = makeTasks(t, { count: 3, tokens: ['', '', 'omega'] });
	return gameView(t, {
		viewer: 0,
		mission: 48,
		tasks,
		objectives: [objectiveView(objective('cardInLastTrick', { cards: [tasks[2].card!] }))]
	});
}

/** Every fixture by name, for tests and the UI preview. */
export const fixtures: Record<string, () => PlayerView> = {
	lobbyView1,
	lobbyView2,
	lobbyView4,
	draftView2,
	pollView2,
	pickView2,
	adjustTokensView2,
	passView2,
	passChosenView2,
	playView2,
	playJarvisView2,
	debriefWonView2,
	debriefLostView2,
	playView3,
	draftView5,
	playView5,
	handoverView5,
	mission12View,
	mission50View,
	mission48View
};
