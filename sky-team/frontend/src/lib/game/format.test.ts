import { describe, expect, it } from 'vitest';
import {
	ABILITY,
	CONDITION,
	LOSS,
	MODULE,
	altitudeLabel,
	failedConditions,
	firstPlayerOf,
	moduleList,
	myRole,
	otherRole,
	partnerOf,
	plural,
	resultHeadline,
	scenarioGroups,
	scenarioName,
	scenarioShort,
	spaceLabel,
	spaceValues,
	turnAllows
} from './format.js';
import { catalogue, fixtures } from './fixtures.js';
import type { AbilityId, Condition, LossReason, Module } from '#lib/types.js';

describe('roles', () => {
	it('swap and alternate', () => {
		expect(otherRole('pilot')).toBe('copilot');
		expect(otherRole('copilot')).toBe('pilot');
		expect([1, 2, 3, 4, 5, 6, 7].map(firstPlayerOf)).toEqual([
			'pilot',
			'copilot',
			'pilot',
			'copilot',
			'pilot',
			'copilot',
			'pilot'
		]);
	});
	it('find the viewer and the partner', () => {
		const v = fixtures.hndTurn();
		expect(myRole(v)).toBe('copilot');
		expect(partnerOf(v)?.role).toBe('pilot');
		expect(partnerOf(fixtures.lobbyAlone())).toBeUndefined();
	});
});

describe('texts', () => {
	it('exist for every module, ability, loss and condition', () => {
		const modules: Module[] = ['kerosene', 'keroseneLeak', 'intern', 'wind', 'iceBrakes'];
		const abilities: AbilityId[] = [
			'adaptation',
			'anticipation',
			'control',
			'mastery',
			'synchronisation',
			'workingTogether'
		];
		const losses: LossReason[] = [
			'spin',
			'collision',
			'overshoot',
			'turn',
			'missedAirport',
			'mandatory',
			'kerosene',
			'landing',
			'intern',
			'iceBrakes',
			'gaveUp'
		];
		const conditions: Condition[] = ['A', 'B', 'C', 'D', 'intern', 'iceBrakes'];
		for (const m of modules) expect(MODULE[m].name && MODULE[m].text, m).toBeTruthy();
		for (const a of abilities) expect(ABILITY[a].name && ABILITY[a].text, a).toBeTruthy();
		for (const l of losses) expect(LOSS[l].title && LOSS[l].text, l).toBeTruthy();
		for (const c of conditions) expect(CONDITION[c], c).toBeTruthy();
	});
	it('list the failed conditions in words', () => {
		expect(failedConditions(['A', 'D'])).toEqual([CONDITION.A, CONDITION.D]);
		expect(failedConditions([])).toEqual([]);
	});
	it('headline a result', () => {
		expect(resultHeadline({ won: true, reason: null, failed: [] })).toBe('Safe landing');
		expect(resultHeadline({ won: false, reason: 'spin', failed: [] })).toBe('The plane spun');
		expect(resultHeadline({ won: false, reason: null, failed: [] })).toBe('Flight lost');
	});
	it('pluralise', () => {
		expect(plural(1, 'die', 'dice')).toBe('1 die');
		expect(plural(3, 'die', 'dice')).toBe('3 dice');
		expect(plural(2, 'plane')).toBe('2 planes');
	});
});

describe('scenarios', () => {
	const all = catalogue();
	it('name and abbreviate', () => {
		const tgu = all.find((s) => s.id === 'tgu-red')!;
		expect(scenarioShort(tgu)).toBe('TGU red');
		expect(scenarioName(tgu)).toBe('TGU Toncontín, red');
		expect(moduleList(tgu)).toBe('Kerosene, Wind');
		expect(moduleList(all[0])).toBe('No modules');
	});
	it('group by colour in difficulty order', () => {
		const groups = scenarioGroups(all);
		expect(groups.map((g) => g.colour)).toEqual(['green', 'yellow', 'red', 'black']);
		expect(groups.map((g) => g.scenarios.length)).toEqual([6, 8, 5, 4]);
		expect(groups.flatMap((g) => g.scenarios)).toHaveLength(23);
		expect(groups[1].scenarios.map((s) => s.id)).toContain('dus-yellow');
	});
});

describe('altitude', () => {
	it('labels the seven rounds', () => {
		expect([1, 2, 3, 4, 5, 6, 7].map(altitudeLabel)).toEqual([
			'6000',
			'5000',
			'4000',
			'3000',
			'2000',
			'1000',
			'Landing'
		]);
		expect(altitudeLabel(8)).toBe('');
	});
});

describe('spaces', () => {
	it('label and constrain', () => {
		expect(spaceLabel('flaps23')).toBe('Flaps 2–3');
		expect(spaceLabel('radioCopilot2')).toBe('Radio');
		expect(spaceLabel('ice4Bottom')).toBe('Ice 4');
		expect(spaceValues('gear56')).toEqual([5, 6]);
		expect(spaceValues('brakes4')).toEqual([4]);
		expect(spaceValues('ice3Top')).toEqual([3]);
		expect(spaceValues('axisPilot')).toBeNull();
		expect(spaceValues('kerosene')).toBeNull();
	});
	it('read turn masks by axis position', () => {
		expect(turnAllows('xooxx', -2)).toBe(false);
		expect(turnAllows('xooxx', -1)).toBe(true);
		expect(turnAllows('xooxx', 0)).toBe(true);
		expect(turnAllows('xooxx', 1)).toBe(false);
		expect(turnAllows(null, 2)).toBe(true);
	});
});
