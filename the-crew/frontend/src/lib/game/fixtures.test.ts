import { describe, expect, it } from 'vitest';
import { cardKey, sameCard } from './cards.js';
import {
	FIXTURE_MISSIONS,
	fixtures as planetNineFixtures,
	fullDeck,
	missionDef,
	presets
} from './fixtures.js';
import {
	deepSeaFixtures,
	deepSeaLobbyFixtures,
	deepSeaPresets,
	deepSeaUiFixtures
} from './fixturesDeepSea.js';
import {
	OWNER_CREW,
	OWNER_UNASSIGNED,
	type Card,
	type Phase,
	type PlayerView
} from '#lib/types.js';

// Both editions' fixtures go through the same checks; the Deep Sea ones have more below.
const fixtures = { ...planetNineFixtures, ...deepSeaFixtures, ...deepSeaUiFixtures };

const TOTAL_TRICKS: Record<number, number> = { 2: 13, 3: 13, 4: 10, 5: 8 };

function validCard(c: Card): boolean {
	return c.suit === 'trump' ? c.rank >= 1 && c.rank <= 4 : c.rank >= 1 && c.rank <= 9;
}

/** The cards a client shows face up at once: they must all be different. */
function visibleCards(v: PlayerView): Card[] {
	const a = v.attempt;
	if (!a) return [];
	const cards: Card[] = [...a.hand, ...a.trick.map((p) => p.card)];
	if (a.lastTrick) cards.push(...a.lastTrick.plays.map((p) => p.card));
	for (const m of a.members) {
		// The viewer's own communicated cards are still in the hand.
		if (m.seat === a.seat) continue;
		cards.push(...m.communications.map((c) => c.card));
		// A debrief reveals everything a member still holds, which includes JARVIS's face-up cards.
		if (m.cards) cards.push(...m.cards);
		else cards.push(...m.stacks.flatMap((s) => (s.up ? [s.up] : [])));
	}
	return cards;
}

describe('fixtures', () => {
	it('cover every phase in each edition (dev route /dev/fixtures/<name>)', () => {
		const phases: Record<Phase, true> = { lobby: true, playing: true, debrief: true, ended: true };
		for (const set of [planetNineFixtures, { ...deepSeaLobbyFixtures, ...deepSeaFixtures }]) {
			const covered = new Set(Object.values(set).map((make) => make().phase));
			expect([...covered].sort()).toEqual(Object.keys(phases).sort());
		}
	});

	it('offers a fresh object every call', () => {
		for (const [name, make] of Object.entries(fixtures)) {
			const first = make();
			expect(make(), name).toEqual(first);
			expect(make(), name).not.toBe(first);
			first.players.length = 0;
			expect(make().players.length, name).toBeGreaterThan(0);
		}
	});

	it('covers the fixture list of the plan', () => {
		expect(Object.keys(fixtures)).toEqual(
			expect.arrayContaining([
				'lobbyView1',
				'lobbyView2',
				'lobbyView4',
				'draftView2',
				'pollView2',
				'pickView2',
				'adjustTokensView2',
				'passView2',
				'playView2',
				'playJarvisView2',
				'debriefWonView2',
				'debriefLostView2',
				'playView3',
				'draftView5',
				'playView5',
				'handoverView5',
				'mission12View',
				'mission50View',
				'mission48View'
			])
		);
	});

	it('starts the table UI only in states the view allows (dev route presets)', () => {
		const all = { ...planetNineFixtures, ...deepSeaUiFixtures };
		for (const [name, preset] of Object.entries({ ...presets, ...deepSeaPresets })) {
			expect(Object.hasOwn(all, name), name).toBe(true);
			const v = all[name]();
			const a = v.attempt!;
			const p = preset(v);
			if (p.selected)
				expect(
					a.playable.some((c) => sameCard(c, p.selected)),
					name
				).toBe(true);
			if (p.communicate) expect(a.actions.communicate.length, name).toBeGreaterThan(0);
			if (p.shown) {
				expect(
					a.actions.communicate.some((o) => sameCard(o.card, p.shown)),
					name
				).toBe(true);
			}
			for (const id of p.tasks ?? []) expect(a.takeable, name).toContain(id);
		}
	});

	it('has stress views: full rows, communications, the last trick', () => {
		const lead = planetNineFixtures.stressView2().attempt!;
		expect(lead.tasks).toHaveLength(10);
		expect(lead.lastTrick).not.toBeNull();
		expect(lead.trick).toEqual([]);
		expect(lead.members[1].communications).toHaveLength(1);
		expect(lead.actions.communicate.length).toBeGreaterThan(0);
		const comm = planetNineFixtures.stressCommView2().attempt!;
		expect(comm.members[0].communications).toHaveLength(1);
		expect(comm.members[1].communications).toHaveLength(1);
		const draft = planetNineFixtures.draftFullView2().attempt!;
		expect(draft.tasks.filter((t) => t.owner === OWNER_UNASSIGNED)).toHaveLength(10);
		for (const n of [2, 3, 5]) {
			const ds = deepSeaFixtures[`dsStress${n}`]().attempt!;
			expect(ds.lastTrick).not.toBeNull();
			expect(ds.members[1].communications).toHaveLength(1);
			expect(ds.actions.communicate.length).toBeGreaterThan(0);
		}
	});

	it('knows the mission rows it needs', () => {
		for (const n of FIXTURE_MISSIONS) expect(missionDef(n).number).toBe(n);
		expect(() => missionDef(1000)).toThrow();
	});

	describe.each(Object.entries(fixtures))('%s', (_name, make) => {
		const v = make();
		const a = v.attempt;

		it('has a consistent viewer, host and mission', () => {
			const me = v.players.find((p) => p.id === v.you);
			expect(me?.isYou).toBe(true);
			expect(v.players.filter((p) => p.isYou)).toHaveLength(1);
			expect(v.players.map((p) => p.seat)).toEqual(v.players.map((_, i) => i));
			expect(v.players.find((p) => p.id === v.hostId)?.isHost).toBe(true);
			expect(v.mission).toBeGreaterThanOrEqual(1);
			expect(v.mission).toBeLessThanOrEqual(50);
			expect(v.missionDef?.number).toBe(v.mission);
			expect(v.logbook.map((e) => e.mission)).toEqual(
				[...v.logbook.map((e) => e.mission)].sort((x, y) => x - y)
			);
			for (const e of v.logbook) expect(e.total).toBe(e.attempts + (e.distress ? 1 : 0));
			if (a) expect(a.seat).toBe(me?.seat);
			else expect(['lobby', 'ended']).toContain(v.phase);
		});

		it('has the right members and counts', () => {
			if (!a) return expect(v.attempt).toBeNull();
			const humans = v.players.length;
			const seats = humans === 2 ? 3 : humans;
			expect(a.members).toHaveLength(seats);
			expect(a.members.map((m) => m.seat)).toEqual(a.members.map((_, i) => i));
			expect(a.totalTricks).toBe(TOTAL_TRICKS[humans]);
			expect(a.members[a.seat].playerId).toBe(v.you);
			// Hands minus the cards already played.
			const played = (a.trickNumber - 1) * seats + a.trick.length;
			expect(a.members.reduce((n, m) => n + m.cardCount, 0)).toBe(40 - played);
			expect(a.members.reduce((n, m) => n + m.tricksWon, 0)).toBe(a.trickNumber - 1);
			expect(a.hand).toHaveLength(a.members[a.seat].cardCount);
			expect(a.hand.every(validCard)).toBe(true);
			// JARVIS is the third seat of a two-player game and nobody else is a dummy.
			a.members.forEach((m) => {
				const dummy = humans === 2 && m.seat === 2;
				expect(m.dummy).toBe(dummy);
				expect(m.playerId === '').toBe(dummy);
				expect(m.name === '').toBe(dummy);
				expect(m.stacks).toHaveLength(dummy ? 7 : 0);
				expect(m.mustAct).toBe(a.actors.includes(m.seat));
			});
		});

		it('shows no card twice', () => {
			const keys = visibleCards(v).map(cardKey);
			expect(new Set(keys).size).toBe(keys.length);
			const deck = new Set(fullDeck().map(cardKey));
			for (const k of keys) expect(deck.has(k)).toBe(true);
		});

		it('keeps playable within the hand or JARVIS face-up cards', () => {
			if (!a) return expect(v.attempt).toBeNull();
			if (a.playable.length === 0) return expect(a.playable).toEqual([]);
			expect(a.step).toBe('play');
			expect(a.actors).toEqual([a.seat]);
			const dummyTurn = a.members[a.turn].dummy;
			const pool = dummyTurn
				? a.members[a.turn].stacks.flatMap((s) => (s.up ? [s.up] : []))
				: a.hand;
			expect(a.playable.every((c) => pool.some((p) => sameCard(p, c)))).toBe(true);
			if (dummyTurn) expect(a.commander).toBe(a.seat);
		});

		it('hides what the viewer must not see', () => {
			if (!a) return expect(v.attempt).toBeNull();
			const over = a.result !== null;
			expect(a.step === '').toBe(over);
			expect(v.phase).toBe(over ? 'debrief' : 'playing');
			for (const m of a.members) expect(m.cards === null).toBe(!over);
			if (over) {
				const mine = a.members[a.seat].cards ?? [];
				expect(mine).toEqual(a.hand);
			}
			for (const t of a.tasks) {
				// Deep Sea task cards have an objective instead of a playing card.
				const deepSea = v.edition === 'deepSea';
				expect(t.card === null).toBe(deepSea || (t.faceDown && !over));
				expect(t.owner === OWNER_UNASSIGNED || (t.owner >= 0 && t.owner < a.members.length)).toBe(
					true
				);
				if (!deepSea) {
					expect([t.cardId, t.difficulty, t.objective, t.predicted]).toEqual([0, 0, null, false]);
				} else if (t.faceDown && !over) {
					expect([t.cardId, t.objective, t.predicted]).toEqual([0, null, false]);
					expect(t.difficulty).toBeGreaterThan(0);
				} else {
					expect(t.cardId).toBeGreaterThanOrEqual(1);
					expect(t.cardId).toBeLessThanOrEqual(96);
					expect(t.difficulty).toBeGreaterThan(0);
					expect(t.objective?.owner).toBe(t.owner);
				}
			}
			for (const o of a.objectives) {
				const ok =
					o.owner === OWNER_UNASSIGNED ||
					o.owner === OWNER_CREW ||
					(o.owner >= 0 && o.owner < a.members.length);
				expect(ok).toBe(true);
			}
			// Turn is -1 outside play and draft.
			if (a.step !== 'play' && a.step !== 'draft') expect(a.turn).toBe(-1);
			else expect(a.turn).toBeGreaterThanOrEqual(0);
			// The last completed trick stays visible during the next one (rules §16).
			expect(a.lastTrick !== null).toBe(a.trickNumber > 1);
			// Communication is not offered mid-trick, over, or outside play.
			if (a.trick.length > 0 || over || a.step !== 'play') {
				expect(a.actions.communicate).toEqual([]);
				expect(a.actions.sendDistress).toBe(false);
				expect(a.actions.handover).toBe(false);
			}
			// Passing and replacing are for the draft (replacing: any step of the selection).
			if (a.actions.pass) expect(a.step).toBe('draft');
			if (a.actions.replaceTasks) {
				expect(['draft', 'poll', 'pick', 'assign']).toContain(a.step);
				expect(a.seat).toBe(a.commander);
			}
		});
	});
});

describe('named fixtures', () => {
	it('has lobbies of 1, 2 and 4 players with a logbook', () => {
		for (const [name, n] of [
			['lobbyView1', 1],
			['lobbyView2', 2],
			['lobbyView4', 4]
		] as const) {
			const v = fixtures[name]();
			expect(v.phase, name).toBe('lobby');
			expect(v.players, name).toHaveLength(n);
			expect(v.attempt, name).toBeNull();
			expect(v.logbook.length, name).toBeGreaterThanOrEqual(3);
			expect(v.mission, name).toBe(4);
			expect(v.logbook.find((e) => e.mission === 4)?.completed).toBe(false);
			// The distress signal is a lobby option and off by default.
			expect(v.options, name).toEqual({ distress: false });
		}
		expect(fixtures.lobbyView4().players.find((p) => p.isYou)?.isHost).toBe(false);
	});

	it('has one two-player fixture for each step', () => {
		const steps = (
			[
				['draftView2', 'draft', 4],
				['pollView2', 'poll', 5],
				['pickView2', 'pick', 5],
				['adjustTokensView2', 'adjustTokens', 23],
				['passView2', 'pass', 10],
				['playView2', 'play', 10],
				['playJarvisView2', 'play', 10],
				['debriefWonView2', '', 10],
				['debriefLostView2', '', 10]
			] as const
		).map(([name, step, mission]) => ({ name, step, mission, v: fixtures[name]() }));
		for (const { name, step, mission, v } of steps) {
			expect(v.attempt?.step, name).toBe(step);
			expect(v.mission, name).toBe(mission);
			expect(v.players, name).toHaveLength(2);
			expect(v.attempt?.members[2].dummy, name).toBe(true);
		}
		expect(fixtures.draftView2().attempt?.turn).toBe(0);
		expect(fixtures.pollView2().attempt?.poll?.kind).toBe('wellIll');
		expect(fixtures.pickView2().attempt?.commander).toBe(0);
		expect(fixtures.pickView2().attempt?.pick).toBe('objective');
		expect(fixtures.passView2().attempt?.distress).toBe(true);
		expect(fixtures.passView2().attempt?.pass?.direction).toBeDefined();
	});

	it('shows JARVIS with hidden stacks and a played one', () => {
		const stacks = fixtures.playView2().attempt!.members[2].stacks;
		expect(stacks.some((s) => s.hidden)).toBe(true);
		expect(stacks.some((s) => s.up === null)).toBe(true);
	});

	it('gives the player a strict subset to play, or JARVIS its legal cards', () => {
		const mine = fixtures.playView2().attempt!;
		expect(mine.turn).toBe(mine.seat);
		expect(mine.trick.length).toBeGreaterThan(0);
		expect(mine.playable.length).toBeGreaterThan(0);
		expect(mine.playable.length).toBeLessThan(mine.hand.length);
		const led = mine.trick[0].card.suit;
		expect(mine.playable.every((c) => c.suit === led)).toBe(true);

		const jarvis = fixtures.playJarvisView2().attempt!;
		expect(jarvis.turn).toBe(2);
		expect(jarvis.commander).toBe(jarvis.seat);
		expect(jarvis.actors).toEqual([jarvis.seat]);
		expect(jarvis.members[2].mustAct).toBe(false);
		expect(jarvis.playable.length).toBeGreaterThan(0);
	});

	it('ends won and lost debriefs with their reasons', () => {
		const won = fixtures.debriefWonView2();
		expect(won.attempt?.result?.success).toBe(true);
		expect(won.attempt?.tasks.every((t) => t.status === 'fulfilled')).toBe(true);
		expect(won.logbook.find((e) => e.mission === 10)?.completed).toBe(true);
		const lost = fixtures.debriefLostView2().attempt!;
		expect(lost.result).toMatchObject({ success: false, reason: 'taskFailed', task: 0 });
		expect(lost.tasks[0].status).toBe('failed');
		expect(lost.tasks[0].owner).not.toBe(lost.result?.seat);
		expect(lost.tasks.filter((t) => t.status === 'open')).toHaveLength(2);
	});

	it('has a dead-zone three-player game with a communication and a last trick', () => {
		const v = fixtures.playView3();
		const a = v.attempt!;
		expect(v.players).toHaveLength(3);
		expect(v.missionDef?.deadZone).toBe(true);
		const other = a.members.find((m) => m.communication);
		expect(other).toBeDefined();
		expect(other?.seat).not.toBe(a.seat);
		expect(other?.communication?.position).toBe('');
		expect(a.lastTrick).not.toBeNull();
		expect(a.actions.communicate.length).toBeGreaterThan(0);
		expect(a.actions.communicate.every((o) => o.position === '')).toBe(true);
	});

	it('has five-player fixtures with 8 tricks', () => {
		for (const name of ['draftView5', 'playView5', 'handoverView5'] as const) {
			const a = fixtures[name]().attempt!;
			expect(a.members, name).toHaveLength(5);
			expect(a.totalTricks, name).toBe(8);
		}
		const draft = fixtures.draftView5().attempt!;
		expect(draft.step).toBe('draft');
		expect(draft.turn).toBe(2);
		expect(draft.tasks.filter((t) => t.owner === OWNER_UNASSIGNED)).toHaveLength(3);
		const handover = fixtures.handoverView5();
		expect(handover.missionDef?.golden).toBe(true);
		expect(handover.attempt?.actions.handover).toBe(true);
		expect(handover.attempt?.trickNumber).toBe(1);
		expect(handover.attempt?.trick).toEqual([]);
		expect(fixtures.playView5().attempt?.actions.handover).toBe(false);
	});

	it('has the special missions', () => {
		const m12 = fixtures.mission12View().attempt!;
		expect(m12.exchange?.kind).toBe('draw');
		expect(m12.exchange?.gave).not.toBeNull();
		expect(m12.exchange?.got).not.toBeNull();
		expect(m12.hand.some((c) => sameCard(c, m12.exchange?.got))).toBe(true);
		expect(m12.hand.some((c) => sameCard(c, m12.exchange?.gave))).toBe(false);
		expect(m12.tasks.some((t) => t.token === 'omega')).toBe(true);

		const m50 = fixtures.mission50View().attempt!;
		expect(m50.objectives).toHaveLength(2);
		expect(m50.objectives.map((o) => o.owner)).toEqual([1, 2]);
		expect(m50.objectives[1].tricks).toEqual([-1]);
		expect(m50.tasks).toEqual([]);

		const m48 = fixtures.mission48View().attempt!;
		const omega = m48.tasks.find((t) => t.token === 'omega');
		expect(omega).toBeDefined();
		expect(m48.objectives[0].kind).toBe('cardInLastTrick');
		expect(sameCard(m48.objectives[0].cards[0], omega?.card)).toBe(true);
	});
});
