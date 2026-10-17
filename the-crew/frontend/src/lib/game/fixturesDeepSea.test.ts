import { describe, expect, it } from 'vitest';
import { deepSeaBriefing, objectiveText, stepText } from './format.js';
import { DEEP_SEA_PLAYERS, DEEP_SEA_SITUATIONS, deepSeaFixtures } from './fixturesDeepSea.js';
import { TASK_CARDS } from './taskCards.fixture.js';
import { taskTextOf } from './taskText.js';
import { OWNER_UNASSIGNED } from '#lib/types.js';

// The general checks (cards shown once, hands, hiding) run over these views in fixtures.test.ts;
// these are the Deep Sea ones.
const at = (name: string) => deepSeaFixtures[name]().attempt!;

describe('Deep Sea fixtures', () => {
	const names = Object.keys(deepSeaFixtures);

	it('has every situation for 2, 3 and 5 players', () => {
		expect(DEEP_SEA_SITUATIONS.length).toBeGreaterThanOrEqual(20);
		expect(names).toHaveLength(DEEP_SEA_SITUATIONS.length * DEEP_SEA_PLAYERS.length);
		for (const situation of DEEP_SEA_SITUATIONS) {
			for (const n of DEEP_SEA_PLAYERS) {
				const v = deepSeaFixtures[`${situation}${n}`]();
				expect(v.players, `${situation}${n}`).toHaveLength(n);
				// Two players play with Tonoja as a third crew member.
				expect(
					v.attempt?.members.some((m) => m.dummy),
					`${situation}${n}`
				).toBe(n === 2);
			}
		}
	});

	it('covers the steps and situations of the plan', () => {
		// Draft with pass available, replacing tasks, the hardest task first.
		expect(at('dsDraftPass3').actions.pass).toBe(true);
		expect(at('dsDraftStart5').actions).toMatchObject({ pass: true, replaceTasks: true });
		expect(at('dsDraftStart3').redeals).toBe(1);
		expect(at('dsDraftHardest3').takeable.length).toBeGreaterThan(0);
		// Mission 6, volunteers, free selection, the split of 26, a passed-on mission 10.
		expect(at('dsOneMemberPick2')).toMatchObject({ step: 'pick', pick: 'oneMember' });
		expect(at('dsOneMemberPick2').pickable).toEqual([0, 1, 2]);
		expect(at('dsVolunteerPoll3')).toMatchObject({ step: 'poll', actors: [1] });
		expect(at('dsVolunteerPick5')).toMatchObject({ step: 'pick', pick: 'volunteers' });
		expect(at('dsFreeAssign3').step).toBe('assign');
		expect(at('dsSplitAssign5')).toMatchObject({ step: 'assign', chosen: [0, 1] });
		expect(at('dsPlayFirstTrickOnly3').commFirstTrickOnly).toBe(true);
		// Predictions: the captain predicts for Tonoja; own pending, the other's pending.
		const captain = at('dsCaptainPredict2');
		expect(captain).toMatchObject({ step: 'predict', actors: [0] });
		expect(captain.tasks[captain.predict].owner).toBe(2);
		expect(at('dsPredictOwn3')).toMatchObject({ step: 'predict', actors: [0] });
		expect(at('dsPredictOther3')).toMatchObject({ step: 'predict', actors: [1] });
		// Communication modes.
		const rapture = at('dsPlayRapture5');
		expect(rapture.comm).toBe('rapture');
		expect(rapture.members.some((m) => m.communications.length === 2)).toBe(true);
		expect(rapture.tokensLeft).toBe(1);
		expect(at('dsPlayRapture3').tokensLeft).toBe(0);
		expect(at('dsPlayCurrents3').members[1].communication?.position).toBe('');
		expect(at('dsPlayTerrain3').terrain).toEqual({ suit: 'green', rank: 5 });
		// Debriefs.
		expect(at('dsDebriefLost3').result).toMatchObject({ success: false, reason: 'taskFailed' });
		const epilogue = deepSeaFixtures.dsDebriefLostEpilogue3();
		expect(epilogue.missionDef?.uncounted).toBe(true);
		expect(epilogue.logbook.find((e) => e.mission === 33)?.attempts).toBe(0);
	});

	describe.each(Object.entries(deepSeaFixtures))('%s', (_name, make) => {
		const v = make();
		const a = v.attempt!;
		const def = v.missionDef!;
		const seats = a.members.length;

		it('is a Deep Sea view of a Deep Sea mission', () => {
			expect(v.edition).toBe('deepSea');
			expect(v.mission).toBeLessThanOrEqual(41);
			expect(def.number).toBe(v.mission);
			expect(a.members[a.commander].dummy).toBe(false);
			// Before the first card is played, the captain holds the captain's card.
			if (a.trickNumber === 1 && a.trick.length === 0) {
				expect(a.hand.some((c) => c.suit === 'trump' && c.rank === 4)).toBe(a.seat === a.commander);
			}
		});

		it('has tasks whose difficulties add up to the mission', () => {
			const sum = a.tasks.reduce((n, t) => n + t.difficulty, 0);
			if (def.difficulty > 0) expect(sum).toBe(def.difficulty);
			else expect(a.tasks).toHaveLength(0);
			expect(a.tasks.some((t) => t.faceDown)).toBe(false);
			// A card is never laid out twice.
			const ids = a.tasks.filter((t) => t.cardId > 0).map((t) => t.cardId);
			expect(new Set(ids).size).toBe(ids.length);
			for (const t of a.tasks) {
				if (!t.objective) continue;
				expect(taskTextOf(t, seats).length).toBeGreaterThan(0);
				expect(t.difficulty).toBe(TASK_CARDS[t.cardId - 1].difficulty[seats - 3]);
				// The captain never owns a captain comparison.
				if (t.objective.kind === 'cmpCaptain') expect(t.owner).not.toBe(a.commander);
			}
		});

		it('shows a hidden prediction only to its owner', () => {
			const over = a.result !== null;
			const actsFor = (owner: number) =>
				owner === a.seat || (a.members[owner]?.dummy && a.seat === a.commander);
			for (const t of a.tasks) {
				if (t.objective?.kind !== 'predict') continue;
				if (!t.predicted) {
					expect(t.objective.count).toBe(-1);
				} else if (t.objective.hidden && !over && !actsFor(t.owner)) {
					expect(t.objective.count).toBe(-1);
				} else {
					expect(t.objective.count).toBeGreaterThanOrEqual(0);
					expect(t.objective.count).toBeLessThanOrEqual(a.totalTricks);
				}
				if (t.objective.count >= 0) expect(t.predicted).toBe(true);
			}
			if (a.step === 'predict') {
				const pending = a.tasks[a.predict];
				expect(pending.objective?.kind).toBe('predict');
				expect(pending.predicted).toBe(false);
				expect(a.actors).toHaveLength(1);
				const owner = a.members[pending.owner];
				expect(a.actors[0]).toBe(owner.dummy ? a.commander : owner.seat);
			} else {
				expect(a.predict).toBe(-1);
			}
		});

		it('has a consistent communication mode', () => {
			const rank = a.terrain?.rank ?? 0;
			if (def.commMode === 'unfamiliar') {
				expect(a.terrain).not.toBeNull();
				expect(a.terrain?.suit).not.toBe('trump');
				expect(a.comm).toBe(rank <= 3 ? '' : rank <= 6 ? 'currents' : 'rapture');
			} else {
				expect(a.terrain).toBeNull();
				expect(a.comm).toBe(def.commMode);
			}
			const used = a.members.reduce((n, m) => n + m.communications.length, 0);
			if (a.comm === 'rapture') {
				expect(a.tokensLeft).toBe(Math.max(seats - 2 - used, 0));
				expect(used).toBeLessThanOrEqual(seats - 2);
			} else {
				expect(a.tokensLeft).toBe(-1);
			}
			for (const m of a.members) {
				expect(m.communication).toEqual(m.communications[0] ?? null);
				expect(m.communicated).toBe(m.communications.length > 0);
				for (const c of m.communications) expect(c.position === '').toBe(a.comm === 'currents');
			}
			for (const o of a.actions.communicate) expect(o.position === '').toBe(a.comm === 'currents');
			if (a.comm === 'none') expect(a.actions.communicate).toEqual([]);
			expect(a.commFirstTrickOnly).toBe(a.chosen.length === 1 && a.chosen[0] !== a.commander);
		});

		it('has consistent selection fields', () => {
			expect(a.chosen.every((s) => s >= 0 && s < seats)).toBe(true);
			const free = a.tasks.filter((t) => t.owner === OWNER_UNASSIGNED).map((t) => t.id);
			// Takeable: only for the one picking in the draft, among the unassigned tasks.
			if (a.step === 'draft') {
				const picker = a.members[a.turn].dummy ? a.commander : a.turn;
				expect(a.takeable.length > 0).toBe(picker === a.seat);
				expect(a.takeable.every((id) => free.includes(id))).toBe(true);
				// Passing is for fewer tasks than members, and only on one's own turn.
				if (a.actions.pass) {
					expect(picker).toBe(a.seat);
					expect(a.tasks.length).toBeLessThan(seats);
				}
			} else {
				expect(a.takeable).toEqual([]);
			}
			// Assignable: per task, for the captain, in the assign step only.
			if (a.step === 'assign' && a.seat === a.commander) {
				expect(a.assignable).toHaveLength(a.tasks.length);
				a.assignable.forEach((seatsOf, id) => {
					expect(seatsOf.length > 0).toBe(free.includes(id));
					expect(seatsOf.every((s) => s >= 0 && s < seats)).toBe(true);
					if (a.chosen.length > 0) expect(seatsOf.every((s) => a.chosen.includes(s))).toBe(true);
				});
			} else {
				expect(a.assignable).toEqual([]);
			}
			// Polls and picks belong to the mission's distribution.
			if (a.step === 'poll') {
				expect(['volunteers', 'captainOrVolunteer']).toContain(def.distribution);
			}
			if (a.pick === 'oneMember') expect(def.distribution).toBe('oneMember');
			if (a.pick === 'volunteers') {
				expect(['volunteers', 'captainOrVolunteer']).toContain(def.distribution);
			}
			if (a.step === 'assign') expect(['free', 'volunteers']).toContain(def.distribution);
		});

		it('reads in words', () => {
			if (a.step !== 'play' && a.step !== '') expect(stepText(a.step, a.pick, def)).not.toBe('');
			expect(deepSeaBriefing(def).length).toBeGreaterThan(0);
			for (const o of a.objectives) expect(objectiveText(o, seats).length).toBeGreaterThan(0);
		});
	});
});
