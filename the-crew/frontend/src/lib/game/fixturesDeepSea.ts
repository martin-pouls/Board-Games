// Sample Deep Sea views for tests and UI previews. Not used by the running app.
//
// Built on the table simulation of fixtures.ts (submarine 4 is the captain's card, Tonoja is the
// dummy of two-player games). Every situation exists for 2 players (with Tonoja), 3 and 5 players:
// `deepSeaFixtures` has it as `<name>2`, `<name>3` and `<name>5`. Task cards come from the 96-card
// table (taskCards.fixture.ts) and are drawn to the mission's difficulty as the server does; the
// mission rows are the ones the server sends (deepSeaMissions.fixture.ts).
import { DEEP_SEA_MISSIONS } from './deepSeaMissions.fixture.js';
import {
	card,
	commOptions,
	communicationOf,
	gameView,
	objective,
	objectiveView,
	seeded,
	shuffled,
	table,
	TOTAL_TRICKS,
	type AttemptOpts,
	type Table
} from './fixtures.js';
import { TASK_CARDS } from './taskCards.fixture.js';
import {
	OWNER_CREW,
	OWNER_UNASSIGNED,
	type Card,
	type Communication,
	type MissionDef,
	type ObjectiveView,
	type PlayerView,
	type Poll,
	type Status,
	type Task
} from '#lib/types.js';

// ---------------------------------------------------------------------------------- helpers

/** The player counts every situation exists for. */
export const DEEP_SEA_PLAYERS = [2, 3, 5] as const;

/** A mission row (a fresh copy). */
export function deepSeaMission(n: number): MissionDef {
	const def = DEEP_SEA_MISSIONS[n - 1];
	if (!def || def.number !== n) throw new Error(`fixtures: no Deep Sea mission ${n}`);
	return structuredClone(def);
}

/** Win-the-tricks-more-than-the-captain cards (4–6): never the captain's. */
const NOT_FOR_CAPTAIN = [4, 5, 6];
/** Prediction cards. */
const PREDICTIONS = [90, 91];

interface DrawOptions {
	/** Cards that are drawn first. */
	must?: number[];
	/** Exactly this many cards (a search over the seeded order); else the server's draw. */
	count?: number;
	/** Allow the captain comparison cards (4–6) and the predictions. */
	extra?: boolean;
}

/**
 * Task card numbers for `seats` crew members with the difficulties adding up to `total`: the
 * `must` cards, then a seeded order of the others, skipping a card that would overshoot (the
 * server's rule). Without the `extra` option the captain comparisons and predictions stay out.
 */
export function drawCards(
	seats: number,
	total: number,
	seed: number,
	o: DrawOptions = {}
): number[] {
	const col = seats - 3;
	const must = o.must ?? [];
	const diff = (id: number) => TASK_CARDS[id - 1].difficulty[col];
	const pool = shuffled(
		TASK_CARDS.map((c) => c.id).filter(
			(id) =>
				!must.includes(id) &&
				(o.extra || !(NOT_FOR_CAPTAIN.includes(id) || PREDICTIONS.includes(id)))
		),
		seeded(seed)
	);
	const base = must.reduce((n, id) => n + diff(id), 0);
	if (base > total) throw new Error(`fixtures: ${must} already exceed difficulty ${total}`);
	if (o.count === undefined) {
		const ids = [...must];
		let sum = base;
		for (const id of pool) {
			if (sum === total) break;
			if (sum + diff(id) <= total) {
				ids.push(id);
				sum += diff(id);
			}
		}
		if (sum !== total) throw new Error(`fixtures: cannot reach difficulty ${total}`);
		return ids;
	}
	const want = o.count - must.length;
	const search = (from: number, left: number, sum: number): number[] | null => {
		if (left === 0) return sum === total ? [] : null;
		for (let i = from; i < pool.length; i++) {
			const d = diff(pool[i]);
			if (sum + d > total) continue;
			const rest = search(i + 1, left - 1, sum + d);
			if (rest) return [pool[i], ...rest];
		}
		return null;
	};
	const found = search(0, want, base);
	if (!found) throw new Error(`fixtures: no ${o.count} cards add up to ${total}`);
	return [...must, ...found];
}

interface TaskSpec {
	owner?: number;
	status?: Status;
	/** A prediction task: the number noted (open, or hidden and shown to its owner). */
	tricks?: number;
	/** A hidden prediction of another member: the number is not sent. */
	masked?: boolean;
}

/** The task as the server sends it. */
function dsTask(t: Table, id: number, cardId: number, spec: TaskSpec = {}): Task {
	const row = TASK_CARDS[cardId - 1];
	const col = t.seats - 3;
	const objective = structuredClone(row.objective);
	const owner = spec.owner ?? OWNER_UNASSIGNED;
	objective.owner = owner;
	if (row.countBy) objective.count = row.countBy[col];
	let predicted = false;
	if (objective.kind === 'predict' && spec.tricks !== undefined) {
		predicted = true;
		objective.count = spec.masked ? -1 : spec.tricks;
	}
	return {
		id,
		card: null,
		token: '',
		owner,
		status: spec.status ?? 'open',
		faceDown: false,
		cardId,
		difficulty: row.difficulty[col],
		objective,
		predicted
	};
}

/** All tasks assigned in draft order from the captain; `over` fixes the ones named by card number. */
function assigned(t: Table, ids: number[], over: Record<number, TaskSpec> = {}, from = 0): Task[] {
	return ids.map((cardId, i) =>
		dsTask(t, i, cardId, { owner: (t.commander + from + i) % t.seats, ...over[cardId] })
	);
}

/** The tasks a seat may take: all unassigned ones, but the captain not cards 4–6. */
function takeable(tasks: Task[], seatIsCaptain: boolean): number[] {
	return tasks
		.filter(
			(t) =>
				t.owner === OWNER_UNASSIGNED &&
				!(seatIsCaptain && t.objective && t.objective.kind === 'cmpCaptain')
		)
		.map((t) => t.id);
}

/** Per task the seats it may go to now, as the server's `Assignable` (free selection, volunteers). */
function assignable(t: Table, tasks: Task[], chosen: number[] | null): number[][] {
	const forbidden = (task: Task, seat: number) =>
		seat === t.commander && task.objective?.kind === 'cmpCaptain';
	return tasks.map((task) => {
		if (task.owner !== OWNER_UNASSIGNED) return [];
		const seats: number[] = [];
		for (let s = 0; s < t.seats; s++) {
			if (forbidden(task, s)) continue;
			if (chosen) {
				if (!chosen.includes(s)) continue;
				// Each volunteer takes at least one task.
				const left = tasks.filter((x) => x.id !== task.id && x.owner === OWNER_UNASSIGNED);
				const empty = chosen.filter((v) => v !== s && !tasks.some((x) => x.owner === v));
				const ok = empty.every((v) => left.some((x) => !forbidden(x, v)));
				if (!ok || left.length < empty.length) continue;
			}
			seats.push(s);
		}
		return seats;
	});
}

/** Yes/no answers of a poll among the humans (the captain and Tonoja never answer). */
function poll(t: Table, yes: Record<number, boolean>): Poll {
	return {
		kind: 'yesNo',
		answers: Array.from({ length: t.seats }, (_, s) => yes[s] ?? null),
		objective: -1,
		task: -1
	};
}

/** The view of a Deep Sea attempt. */
function dsView(t: Table, mission: number, o: Omit<AttemptOpts, 'mission' | 'def' | 'deepSea'>) {
	return gameView(t, { ...o, mission, def: deepSeaMission(mission), deepSea: true });
}

const everySeat = (t: Table) => Array.from({ length: t.seats }, (_, s) => s);

// ---------------------------------------------------------------------------------- draft

/** Mission 3 (draft): the captain took a task, the viewer is next and may pass (two tasks, three or more members). */
function draftPass(humans: number): PlayerView {
	const t = table({ humans, commander: 0, seed: 101 });
	const ids = drawCards(t.seats, 3, 1, { count: 2 });
	const tasks = [dsTask(t, 0, ids[0], { owner: 0 }), dsTask(t, 1, ids[1])];
	return dsView(t, 3, {
		viewer: 1,
		step: 'draft',
		draftTurn: 1,
		number: 2,
		tasks,
		actions: { pass: true },
		extra: { takeable: takeable(tasks, false) }
	});
}

/** Mission 4 after one automatic re-deal: the captain (the viewer) picks first and may replace tasks. */
function draftStart(humans: number): PlayerView {
	const t = table({ humans, commander: 0, seed: 111 });
	const ids = drawCards(t.seats, 4, 2, { count: 3 });
	const tasks = ids.map((id, i) => dsTask(t, i, id));
	return dsView(t, 4, {
		viewer: 0,
		step: 'draft',
		draftTurn: 0,
		tasks,
		actions: { pass: t.seats > tasks.length, replaceTasks: true },
		extra: { takeable: takeable(tasks, true), redeals: 1 }
	});
}

/** Mission 19: the captain's first task must be one of the hardest. */
function draftHardest(humans: number): PlayerView {
	const t = table({ humans, commander: 0, seed: 121 });
	const ids = drawCards(t.seats, 9, 3, { count: 4 });
	const tasks = ids.map((id, i) => dsTask(t, i, id));
	const open = takeable(tasks, true);
	const hardest = Math.max(...open.map((id) => tasks[id].difficulty));
	return dsView(t, 19, {
		viewer: 0,
		step: 'draft',
		draftTurn: 0,
		tasks,
		actions: { replaceTasks: true },
		extra: { takeable: open.filter((id) => tasks[id].difficulty === hardest) }
	});
}

// ---------------------------------------------------------------------------------- one member, volunteers, free selection

/** Mission 6: the captain (the viewer) picks the one member who takes all the tasks. */
function oneMemberPick(humans: number): PlayerView {
	const t = table({ humans, commander: 0, seed: 131 });
	const tasks = drawCards(t.seats, 5, 4, { count: 3 }).map((id, i) => dsTask(t, i, id));
	return dsView(t, 6, {
		viewer: 0,
		step: 'pick',
		pick: 'oneMember',
		pickable: everySeat(t),
		tasks,
		actions: { replaceTasks: true }
	});
}

/** Mission 14: the viewer is asked whether they volunteer; the others are partly answered. */
function volunteerPoll(humans: number): PlayerView {
	const t = table({ humans, commander: 0, seed: 141 });
	const tasks = drawCards(t.seats, 6, 5, { count: 3 }).map((id, i) => dsTask(t, i, id));
	const answers: Record<number, boolean> = humans >= 3 ? { 2: true } : {};
	if (humans === 5) answers[3] = false;
	return dsView(t, 14, {
		viewer: 1,
		step: 'poll',
		tasks,
		poll: poll(t, answers),
		extra: { comm: 'currents' }
	});
}

/** Mission 14: everybody answered; the captain (the viewer) picks the volunteer, Tonoja included. */
function volunteerPick(humans: number): PlayerView {
	const t = table({ humans, commander: 0, seed: 151 });
	const tasks = drawCards(t.seats, 6, 5, { count: 3 }).map((id, i) => dsTask(t, i, id));
	const answers: Record<number, boolean> = { 1: true };
	if (humans >= 3) answers[2] = false;
	if (humans === 5) {
		answers[3] = true;
		answers[4] = false;
	}
	return dsView(t, 14, {
		viewer: 0,
		step: 'pick',
		pick: 'volunteers',
		pickable: everySeat(t),
		tasks,
		poll: poll(t, answers),
		actions: { replaceTasks: false },
		extra: { comm: 'currents' }
	});
}

/** Mission 17 (free selection): two tasks are assigned; the captain (the viewer) assigns the rest. */
function freeAssign(humans: number): PlayerView {
	const t = table({ humans, commander: 0, seed: 161 });
	const ids = drawCards(t.seats, 9, 6, { must: [4], count: 5, extra: false });
	const tasks = ids.map((id, i) =>
		// The captain comparison (card 4) stays open; tasks 1 and 2 are taken.
		dsTask(t, i, id, i === 1 ? { owner: 1 } : i === 2 ? { owner: 0 } : {})
	);
	return dsView(t, 17, {
		viewer: 0,
		step: 'assign',
		tasks,
		extra: { assignable: assignable(t, tasks, null) }
	});
}

/** Mission 26: the captain (the viewer) and seat 1 volunteered and split the tasks; seat 1 has one. */
function splitAssign(humans: number): PlayerView {
	const t = table({ humans, commander: 0, seed: 171 });
	const ids = drawCards(t.seats, 12, 7, { count: 5 });
	const chosen = [0, 1];
	const tasks = ids.map((id, i) => dsTask(t, i, id, i === 0 ? { owner: 1 } : {}));
	return dsView(t, 26, {
		viewer: 0,
		step: 'assign',
		tasks,
		extra: {
			chosen,
			assignable: assignable(t, tasks, chosen),
			terrain: card('yellow', 2),
			comm: ''
		}
	});
}

/** Mission 10, passed on: seat 1 took all tasks, so communication ends with the first trick. */
function playFirstTrickOnly(humans: number): PlayerView {
	const t = table({ humans, commander: 0, seed: 181 });
	const ids = drawCards(t.seats, 4, 8, { count: 3 });
	const tasks = ids.map((id, i) => dsTask(t, i, id, { owner: 1 }));
	return dsView(t, 10, {
		viewer: 0,
		tasks,
		extra: { chosen: [1], commFirstTrickOnly: true }
	});
}

// ---------------------------------------------------------------------------------- predictions

/** Mission 7: the viewer is the captain and predicts for the hidden prediction task (Tonoja's with two players). */
function captainPredict(humans: number): PlayerView {
	const t = table({ humans, commander: 0, seed: 191 });
	const ids = drawCards(t.seats, 6, 9, { must: [91], count: 3 });
	const owner = t.jarvis ? 2 : 0;
	const tasks = assigned(t, ids, { 91: { owner } }, 1);
	return dsView(t, 7, {
		viewer: 0,
		step: 'predict',
		tasks,
		actors: [0],
		extra: { predict: tasks.findIndex((x) => x.cardId === 91) }
	});
}

/** Mission 9: the open prediction (90) of seat 1 is made; the viewer must predict the hidden one (91). */
function predictOwn(humans: number): PlayerView {
	const t = table({ humans, commander: 0, seed: 201 });
	const ids = drawCards(t.seats, 7, 10, { must: [90, 91] });
	const tasks = assigned(t, ids, { 90: { owner: 1, tricks: 2 }, 91: { owner: 0 } }, 2);
	return dsView(t, 9, {
		viewer: 0,
		step: 'predict',
		tasks,
		actors: [0],
		extra: { predict: tasks.findIndex((x) => x.cardId === 91), comm: 'currents' }
	});
}

/** Mission 9: the viewer's open prediction is made; seat 1 is predicting the hidden one. */
function predictOther(humans: number): PlayerView {
	const t = table({ humans, commander: 0, seed: 211 });
	const ids = drawCards(t.seats, 7, 10, { must: [90, 91] });
	const tasks = assigned(t, ids, { 90: { owner: 0, tricks: 3 }, 91: { owner: 1 } }, 2);
	return dsView(t, 9, {
		viewer: 0,
		step: 'predict',
		tasks,
		actors: [1],
		extra: { predict: tasks.findIndex((x) => x.cardId === 91), comm: 'currents' }
	});
}

/** Mission 9, trick 4: the viewer sees their own hidden prediction (1) and the open one of seat 1 (2). */
function playPredictOwn(humans: number): PlayerView {
	const t = table({ humans, commander: 0, seed: 221, tricks: 3, plays: 0 });
	const ids = drawCards(t.seats, 7, 10, { must: [90, 91] });
	const tasks = assigned(t, ids, { 90: { owner: 1, tricks: 2 }, 91: { owner: 0, tricks: 1 } }, 2);
	return dsView(t, 9, {
		viewer: 0,
		tasks,
		extra: { comm: 'currents' },
		comms: { 1: communicationOf(t, 1, true, 2) },
		actions: { communicate: commOptions(t.hands[0], true) }
	});
}

/** Mission 9, trick 4: seat 1's hidden prediction is a question mark to the viewer, who has the open one. */
function playPredictOther(humans: number): PlayerView {
	const t = table({ humans, commander: 0, seed: 231, tricks: 3, plays: 0 });
	const ids = drawCards(t.seats, 7, 10, { must: [90, 91] });
	const tasks = assigned(
		t,
		ids,
		{ 90: { owner: 0, tricks: 2 }, 91: { owner: 1, tricks: 1, masked: true } },
		2
	);
	return dsView(t, 9, {
		viewer: 0,
		tasks,
		extra: { comm: 'currents' },
		comms: { 1: communicationOf(t, 1, true, 2) },
		actions: { communicate: commOptions(t.hands[0], true) }
	});
}

// ---------------------------------------------------------------------------------- communication modes

/** Mission 11 (rapture of the deep) in trick 2: sonar tokens are shared; with five players seat 1 used two. */
function playRapture(humans: number): PlayerView {
	const t = table({
		humans,
		commander: 0,
		seed: 241,
		tricks: 1,
		plays: 0,
		accept: (x) =>
			commOptions(x.hands[1], false).length >= 2 && commOptions(x.hands[0], false).length >= 1
	});
	const tasks = assigned(t, drawCards(t.seats, 8, 11, {}));
	const tokens = t.seats - 2;
	const options = commOptions(t.hands[1], false);
	const made = (o: (typeof options)[number]): Communication => ({ ...o, trick: 1 });
	// Two communications on one member while a token is left; with three members one token was used.
	const comms: Record<number, Communication[]> = {
		1: t.seats === 5 ? [made(options[0]), made(options[1])] : [made(options[0])]
	};
	const tokensLeft = tokens - Object.values(comms).reduce((n, c) => n + c.length, 0);
	return dsView(t, 11, {
		viewer: 0,
		tasks,
		comms,
		actions: tokensLeft > 0 ? { communicate: commOptions(t.hands[0], false) } : { communicate: [] },
		extra: { comm: 'rapture', tokensLeft }
	});
}

/** Mission 9 (currents) in trick 3: seat 1 communicated a card without a position. */
function playCurrents(humans: number): PlayerView {
	const t = table({ humans, commander: 0, seed: 251, tricks: 2, plays: 0 });
	const tasks = assigned(t, drawCards(t.seats, 7, 12, {}));
	return dsView(t, 9, {
		viewer: 0,
		tasks,
		comms: { 1: communicationOf(t, 1, true, 2) },
		actions: { communicate: commOptions(t.hands[0], true) },
		extra: { comm: 'currents' }
	});
}

/** Mission 20 (unfamiliar terrain) before trick 1: the terrain card has picked the mode. */
function playTerrain(humans: number): PlayerView {
	const t = table({ humans, commander: 0, seed: 261 });
	const tasks = assigned(t, drawCards(t.seats, 10, 13, {}));
	const terrain: Card =
		humans === 2 ? card('yellow', 2) : humans === 3 ? card('green', 5) : card('pink', 8);
	const comm = humans === 2 ? '' : humans === 3 ? 'currents' : 'rapture';
	return dsView(t, 20, {
		viewer: 0,
		tasks,
		actions: comm === 'currents' ? { communicate: commOptions(t.hands[0], true) } : {},
		extra: { terrain, comm, tokensLeft: comm === 'rapture' ? t.seats - 2 : -1 }
	});
}

// ---------------------------------------------------------------------------------- side quests

/** Q1, stage B (a new deal): the holder of submarine 1 is the captain and must win exactly three tricks. */
function q1StageB(humans: number): PlayerView {
	const t = table({
		humans,
		commander: 0,
		captainCard: card('trump', 1),
		seed: 271,
		tricks: 2,
		plays: 0
	});
	const stage = deepSeaMission(42).stageObjectives[1][0];
	const objectives: ObjectiveView[] = [
		objectiveView(
			objective(stage.kind, { count: stage.count, without: stage.without, owner: t.commander })
		)
	];
	return dsView(t, 42, {
		viewer: 0,
		number: 3,
		objectives,
		extra: { stage: 1, stages: 2 }
	});
}

/** Q2: the task cards stay face down, only their difficulty shows; the quota adds up to the tricks. */
function q2FaceDown(humans: number): PlayerView {
	const t = table({ humans, commander: 0, seed: 281, tricks: 2, plays: 1 });
	const total = deepSeaMission(43).difficultyBy[t.seats - 3];
	const ids = drawCards(t.seats, total, 14, { count: 5 });
	const tasks: Task[] = ids.map((cardId, i) => ({
		id: i,
		card: null,
		token: '',
		owner: (t.commander + i) % t.seats,
		status: 'open',
		faceDown: true,
		cardId: 0,
		difficulty: TASK_CARDS[cardId - 1].difficulty[t.seats - 3],
		objective: null,
		predicted: false
	}));
	const targets = Array.from({ length: t.seats }, (_, s) =>
		tasks.filter((x) => x.owner === s).reduce((n, x) => n + x.difficulty, 0)
	);
	return dsView(t, 43, {
		viewer: 0,
		tasks,
		objectives: [objectiveView(objective('trickQuota', { owner: OWNER_CREW, targets }))]
	});
}

// ---------------------------------------------------------------------------------- debrief

/** Mission 9 lost in trick 5: a task can no longer be met. The host picks same or new tasks. */
function debriefLost(humans: number): PlayerView {
	const t = table({ humans, commander: 0, seed: 291, tricks: 5 });
	const ids = drawCards(t.seats, 7, 15, { count: 4 });
	const tasks = assigned(t, ids).map((x, i) => ({
		...x,
		status: (i === 0 ? 'failed' : 'open') as Status
	}));
	return dsView(t, 9, {
		viewer: 0,
		step: '',
		number: 2,
		tasks,
		result: { success: false, reason: 'taskFailed', task: 0, objective: -1, trick: 5, seat: -1 },
		extra: { comm: 'currents' }
	});
}

/** Epilogue mission 33 lost: attempts are not counted, so the logbook stays at the earlier tries. */
function debriefLostEpilogue(humans: number): PlayerView {
	const t = table({ humans, commander: 0, seed: 301, tricks: 7 });
	const ids = drawCards(t.seats, 18, 16, {});
	const tasks = assigned(t, ids).map((x, i) => ({
		...x,
		status: (i === 1 ? 'failed' : 'open') as Status
	}));
	return dsView(t, 33, {
		viewer: 0,
		step: '',
		number: 3,
		tasks,
		result: { success: false, reason: 'taskFailed', task: 1, objective: -1, trick: 7, seat: -1 },
		logbookEntry: { attempts: 0, total: 0 }
	});
}

/** Mission 7 won: every task is fulfilled and the hidden prediction is shown to everybody. */
function debriefWon(humans: number): PlayerView {
	const t = table({
		humans,
		commander: 0,
		seed: 311,
		tricks: TOTAL_TRICKS[humans === 2 ? 3 : humans]
	});
	const ids = drawCards(t.seats, 6, 17, { must: [91], count: 3 });
	const tasks = assigned(t, ids, { 91: { owner: 1, tricks: 2 } }, 1).map((x) => ({
		...x,
		status: 'fulfilled' as Status
	}));
	return dsView(t, 7, {
		viewer: 0,
		step: '',
		number: 2,
		tasks,
		result: {
			success: true,
			reason: 'completed',
			task: -1,
			objective: -1,
			trick: t.total,
			seat: -1
		}
	});
}

// ---------------------------------------------------------------------------------- the list

const SITUATIONS: Record<string, (humans: number) => PlayerView> = {
	dsDraftPass: draftPass,
	dsDraftStart: draftStart,
	dsDraftHardest: draftHardest,
	dsOneMemberPick: oneMemberPick,
	dsVolunteerPoll: volunteerPoll,
	dsVolunteerPick: volunteerPick,
	dsFreeAssign: freeAssign,
	dsSplitAssign: splitAssign,
	dsPlayFirstTrickOnly: playFirstTrickOnly,
	dsCaptainPredict: captainPredict,
	dsPredictOwn: predictOwn,
	dsPredictOther: predictOther,
	dsPlayPredictOwn: playPredictOwn,
	dsPlayPredictOther: playPredictOther,
	dsPlayRapture: playRapture,
	dsPlayCurrents: playCurrents,
	dsPlayTerrain: playTerrain,
	dsQ1StageB: q1StageB,
	dsQ2FaceDown: q2FaceDown,
	dsDebriefLost: debriefLost,
	dsDebriefLostEpilogue: debriefLostEpilogue,
	dsDebriefWon: debriefWon
};

/** The Deep Sea situations by name. */
export const DEEP_SEA_SITUATIONS = Object.keys(SITUATIONS);

/** Every Deep Sea fixture by name: each situation with 2, 3 and 5 players (`dsDraftPass2`, …). */
export const deepSeaFixtures: Record<string, () => PlayerView> = Object.fromEntries(
	Object.entries(SITUATIONS).flatMap(([name, make]) =>
		DEEP_SEA_PLAYERS.map((n) => [`${name}${n}`, () => make(n)] as const)
	)
);
