import { describe, expect, it } from 'vitest';
import type { Phase, Step } from '#lib/types.js';
import { fixtures, presets } from './fixtures.js';
import { firstPlayerOf } from './format.js';

// The dev route /dev/fixtures/<name> shows each fixture on the game screen (screenshots, visual
// review), so every phase and every step with a screen needs at least one. A new phase or step
// fails to compile here until it is listed; `false` marks one whose screen is not built yet
// (its work package adds the fixtures and flips the flag).
const phases: Record<Phase, boolean> = {
	lobby: true,
	playing: true,
	won: true,
	lost: true,
	ended: true
};

const steps: Record<Step, boolean> = {
	discussion: true,
	place: true,
	reroll: true,
	swap: true,
	sync: true
};

const all = () => Object.entries(fixtures).map(([name, make]) => [name, make()] as const);

describe('fixtures', () => {
	it('cover every phase that has a screen', () => {
		const covered = new Set(Object.values(fixtures).map((make) => make().phase));
		const expected = Object.entries(phases)
			.filter(([, built]) => built)
			.map(([phase]) => phase);
		expect([...covered].sort()).toEqual(expected.sort());
	});

	it('cover every step of a running game', () => {
		const covered = new Set(
			all()
				.filter(([, v]) => v.phase === 'playing')
				.map(([, v]) => v.flight!.step)
		);
		const expected = Object.entries(steps)
			.filter(([, built]) => built)
			.map(([step]) => step);
		expect([...covered].sort()).toEqual(expected.sort());
	});

	it('return a fresh view on every call', () => {
		for (const [name, make] of Object.entries(fixtures)) {
			const first = make();
			expect(make(), name).toEqual(first);
			expect(make(), name).not.toBe(first);
		}
	});

	it('have exactly one viewer, who is a player', () => {
		for (const [name, v] of all()) {
			expect(
				v.players.filter((p) => p.isYou).map((p) => p.id),
				name
			).toEqual([v.you]);
		}
	});

	it('carry the flight exactly when the game has started', () => {
		for (const [name, v] of all()) {
			const started = v.phase === 'playing' || v.phase === 'won' || v.phase === 'lost';
			expect(v.flight !== null, name).toBe(started);
			expect(v.catalogue !== null, name).toBe(v.phase === 'lobby' || v.phase === 'ended');
		}
	});

	it('have a result exactly when the game is over', () => {
		for (const [name, v] of all()) {
			expect(v.result !== null, name).toBe(v.phase === 'won' || v.phase === 'lost');
			if (v.result) expect(v.result.won, name).toBe(v.phase === 'won');
		}
	});

	it('hold the scenario catalogue of rules.md §9 in the lobby', () => {
		const v = fixtures.lobbyHost();
		expect(v.catalogue).toHaveLength(23);
		expect(v.catalogue!.map((s) => s.id)).toContain(v.scenario.id);
		expect(v.catalogue!.filter((s) => s.realTimeDropped).map((s) => s.id)).toEqual([
			'pbh-red',
			'kul-black',
			'pbh-black',
			'dus-black'
		]);
	});

	it('pick no more ability cards than the scenario has', () => {
		for (const [name, v] of all()) {
			expect(v.abilities.length, name).toBeLessThanOrEqual(v.scenario.abilities);
			expect(new Set(v.abilities).size, name).toBe(v.abilities.length);
		}
	});
});

describe('game fixtures', () => {
	const games = () => all().filter(([, v]) => v.flight !== null);

	it('keep the dice, planes, switches and tokens within their limits', () => {
		for (const [name, v] of games()) {
			const f = v.flight!;
			const placed = f.placements.filter((p) => p.piece === 'pilot' || p.piece === 'copilot');
			expect(f.yourDice, name).toHaveLength(4);
			for (const d of f.yourDice) expect(d === null || (d >= 1 && d <= 6), name).toBe(true);
			expect(f.approach.spaces, name).toHaveLength(v.scenario.spaces);
			expect(f.approach.position, name).toBeGreaterThanOrEqual(1);
			expect(f.approach.position, name).toBeLessThanOrEqual(v.scenario.spaces);
			expect(f.planeSupply, name).toBeGreaterThanOrEqual(0);
			expect(f.coffee, name).toBeGreaterThanOrEqual(0);
			expect(f.coffee, name).toBeLessThanOrEqual(3);
			expect(Math.abs(f.axis), name).toBeLessThanOrEqual(3);
			expect(f.flaps, name).toBeLessThanOrEqual(4);
			expect(f.brakes, name).toBeLessThanOrEqual(3);
			expect(f.rerolls + f.rerollRounds.length, name).toBeLessThanOrEqual(2);
			// Every die is in a hand, on the panel or still to be rolled: 4 per player.
			for (const p of v.players) {
				const onPanel = placed.filter((pl) => pl.piece === p.role).length;
				const inHand = p.diceLeft;
				expect(onPanel + inHand, `${name} ${p.role}`).toBe(f.step === 'discussion' ? onPanel : 4);
			}
		}
	});

	it('place at most one piece on each space', () => {
		for (const [name, v] of games()) {
			const spaces = v.flight!.placements.map((p) => p.space);
			expect(new Set(spaces).size, name).toBe(spaces.length);
		}
	});

	it('agree between the viewer, their dice and the player list', () => {
		for (const [name, v] of games()) {
			const me = v.players.find((p) => p.isYou)!;
			expect(v.flight!.yourDice.filter((d) => d !== null).length, name).toBe(me.diceLeft);
		}
	});

	it('hide the partner dice while the game runs and reveal both hands when it is over', () => {
		for (const [name, v] of games()) {
			for (const p of v.players) {
				if (v.phase === 'playing') expect(p.revealedDice, name).toBeNull();
				else expect(p.revealedDice, name).toHaveLength(p.diceLeft);
			}
		}
	});

	it('offer legal spaces only on the viewer’s turn in the place step', () => {
		for (const [name, v] of games()) {
			const f = v.flight!;
			const me = v.players.find((p) => p.isYou)!;
			const expected =
				v.phase === 'playing' && f.step === 'place' && f.turn === me.role && me.diceLeft > 0;
			expect(f.legal !== null, name).toBe(expected);
			if (!f.legal) continue;
			expect(f.legal.byValue, name).toHaveLength(6);
			const taken = new Set(f.placements.map((p) => p.space));
			for (const list of [...f.legal.byValue, f.legal.internTargets]) {
				for (const space of list) expect(taken.has(space), `${name} ${space}`).toBe(false);
			}
		}
	});

	it('name the first player by the altitude track', () => {
		for (const [name, v] of games()) {
			expect(v.flight!.firstPlayer, name).toBe(firstPlayerOf(v.flight!.round));
		}
	});

	it('keep each step’s own data and nothing else', () => {
		for (const [name, v] of games()) {
			const f = v.flight!;
			expect(f.reroll !== null, name).toBe(f.step === 'reroll');
			expect(f.swap !== null, name).toBe(f.step === 'swap');
			expect(f.sync !== null, name).toBe(f.step === 'sync');
		}
	});

	it('show the wind, fuel, ice and intern board only when the scenario has the module', () => {
		for (const [name, v] of games()) {
			const f = v.flight!;
			const has = (m: string) => v.scenario.modules.includes(m as never);
			expect(f.wind !== null, name).toBe(has('wind'));
			expect(f.ice !== null, name).toBe(has('iceBrakes'));
			expect(f.intern !== null, name).toBe(has('intern'));
			expect(f.fuel !== null, name).toBe(has('kerosene') || has('keroseneLeak'));
			expect(f.leak, name).toBe(has('keroseneLeak'));
		}
	});

	it('only show the Working Together die value, never the initiator’s slot to the partner', () => {
		const offered = fixtures.swapOffered().flight!.swap!;
		const answer = fixtures.swapAnswer().flight!.swap!;
		expect(offered).toMatchObject({ by: 'pilot', value: 6, yourSlot: 0 });
		expect(answer).toMatchObject({ by: 'pilot', value: 6, yourSlot: null });
	});

	it('give only the Co-Pilot the Synchronisation targets', () => {
		expect(fixtures.syncPlace().flight!.sync!.targets.length).toBeGreaterThan(0);
		expect(fixtures.syncWaiting().flight!.sync!.targets).toEqual([]);
	});

	it('have presets that point at a die in hand', () => {
		for (const [name, preset] of Object.entries(presets)) {
			expect(Object.hasOwn(fixtures, name), name).toBe(true);
			const f = fixtures[name]().flight!;
			if (preset.slot !== undefined) expect(f.yourDice[preset.slot], name).not.toBeNull();
			for (const slot of preset.rerollSlots ?? []) expect(f.yourDice[slot], name).not.toBeNull();
		}
	});
});

describe('the worst-case fixtures', () => {
	it('match layout concept A for TGU red', () => {
		const v = fixtures.tguTurn();
		const f = v.flight!;
		expect(v.scenario.id).toBe('tgu-red');
		expect(f.aero).toEqual({ blue: 6, orange: 10 });
		expect(f.wind).toEqual({ offset: -3, value: 2 });
		expect(f.fuel).toBe(11);
		expect(f.yourDice).toEqual([6, null, 2, null]);
		expect(f.placements.find((p) => p.piece === 'traffic')).toMatchObject({
			space: 'radioCopilot2',
			value: 3
		});
		// Brake 6 needs brake 4 first; gear 5–6 is open to the 6.
		expect(f.legal!.byValue[5]).toContain('gear56');
		expect(f.legal!.byValue[5]).not.toContain('brakes6');
	});

	it('match layout concept A for KEF black', () => {
		const f = fixtures.kefTurn().flight!;
		expect(f.ice).toEqual({ marker: 1 });
		expect(f.wind).toEqual({ offset: 4, value: 1 });
		expect(f.placements).toHaveLength(4);
		// Engines 4 + 5 + wind 1 = 10: one space.
		expect(f.legal!.byValue[4]).toContain('enginesPilot');
		expect(f.brakeLimit).toBe(3);
		expect(f.legal!.byValue.flat().some((s) => s.startsWith('brakes'))).toBe(false);
	});

	it('match layout concept A for HND red, seen by the Co-Pilot', () => {
		const v = fixtures.hndTurn();
		const f = v.flight!;
		expect(v.players.find((p) => p.isYou)!.role).toBe('copilot');
		expect(f.intern).toMatchObject({ nextPilot: 2, nextCopilot: 3 });
		expect(f.legal!.internToken).toBe(6);
		// The die 2 with coffee −1 is a 1: radio, the open flaps, training.
		expect(f.legal!.byValue[0]).toEqual(
			expect.arrayContaining(['radioCopilot1', 'enginesCopilot', 'internCopilot', 'flaps12'])
		);
		expect(f.legal!.internTargets).toContain('flaps56');
		expect(f.legal!.internTargets.some((s) => s.startsWith('concentration'))).toBe(false);
		expect(presets.hndTurn).toMatchObject({ slot: 1, coffee: -1 });
	});

	it('show the loss reasons and failed conditions of the game over fixtures', () => {
		expect(fixtures.lostLanding().result).toEqual({
			won: false,
			reason: 'landing',
			failed: ['A', 'B', 'D']
		});
		expect(fixtures.lostSpin().flight!.axis).toBe(-3);
		expect(fixtures.lostKerosene().flight!.fuel).toBeLessThan(0);
		expect(fixtures.lostIntern().flight!.intern!.tokens.filter((t) => !t.taken)).toHaveLength(2);
	});
});
