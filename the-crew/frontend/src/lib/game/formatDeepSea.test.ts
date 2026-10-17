import { describe, expect, it } from 'vitest';
import { terms } from '#lib/edition.js';
import type { Member, Objective } from '#lib/types.js';
import { DEEP_SEA_MISSIONS } from './deepSeaMissions.fixture.js';
import { member as fixtureMember, objective } from './fixtures.js';
import { deepSeaBriefing, objectiveText, resultText, specText, stepText } from './format.js';
import { TASK_CARDS } from './taskCards.fixture.js';
import { taskText } from './taskText.js';

// The words differ per edition (submarine / rocket, captain / commander): the expectations use the
// terms. The briefings are tested with the mission rows as the server sends them.
const T = terms.trump;
const C = terms.commander;
const cap = (s: string) => s.charAt(0).toUpperCase() + s.slice(1);
const mission = (n: number) => DEEP_SEA_MISSIONS[n - 1];
const brief = (n: number) => deepSeaBriefing(mission(n));

describe('the mission objectives', () => {
	it.each<[string, number, number, string]>([
		['8: two 9s more', 8, 0, 'The crew: nobody may ever have won two more 9s than another member'],
		[
			'21: two 1s more',
			21,
			0,
			'The crew: nobody may ever have won two more 1s than another member'
		],
		[
			'12: never lead pink or a submarine',
			12,
			0,
			`The crew: don't open a trick with a pink card or a ${T}`
		],
		[
			'23: the leader stays ahead',
			23,
			0,
			'The crew: whoever wins the first trick must always have won more tricks than every other member'
		],
		[
			'27: yellow 5 last',
			27,
			0,
			'The crew: the yellow 5 must be the last card played in the last trick'
		]
	])('%s', (_what, n, i, want) => {
		expect(specText(mission(n).objectives[i])).toBe(want);
	});

	it('words a task objective like its card, without the capital', () => {
		const card = TASK_CARDS[58 - 1].objective;
		expect(objectiveText(card)).toBe(`win the pink 7 with a ${T}`);
		expect(objectiveText(TASK_CARDS[36 - 1].objective)).toBe('win the green 2 in the last trick');
		expect(taskText(card, 3)).toBe(cap(objectiveText(card)));
	});

	it('keeps Planet Nine objectives as they were', () => {
		const balance = objective('crewBalance', { count: 1 });
		expect(objectiveText(balance)).toBe('nobody may ever have 2 tricks more than another member');
	});
});

describe('the briefing of a Deep Sea mission', () => {
	it('gives difficulty and communication mode first', () => {
		expect(brief(1)[0]).toBe('Difficulty 1');
		expect(brief(9)[0]).toBe('Difficulty 7 · Currents');
		expect(brief(11)[0]).toBe('Difficulty 8 · Rapture of the deep');
		expect(brief(20)[0]).toBe('Difficulty 10 · Unfamiliar terrain');
		expect(brief(16)[0]).toBe('Difficulty 6 · No communication');
	});

	it('explains the communication modes', () => {
		expect(brief(9)).toContain(
			'Currents: communicate a card without saying whether it is highest, only or lowest.'
		);
		expect(brief(11).join(' ')).toContain(
			`${terms.commToken}s, two fewer than there are crew members`
		);
		expect(brief(20).join(' ')).toContain('1–3: normal, 4–6: currents, 7–9: rapture of the deep.');
		expect(brief(16)).toContain('Nobody may communicate.');
		expect(brief(23)).toContain('Nobody may communicate before trick 2.');
	});

	it('explains how the tasks are taken', () => {
		expect(brief(3)).toEqual([
			'Difficulty 3',
			`Tasks are drafted, the ${C} first, then clockwise. A member may pass only while there are fewer tasks than members; every task must be taken after one round.`
		]);
		expect(brief(6)[1]).toBe(
			`The ${C} gives all tasks to one crew member. Decide together, without talking about your hands.`
		);
		expect(brief(10)[1]).toBe(
			`The ${C} takes all tasks or passes them to a willing crew member; the crew answers yes or no first. If they are passed on, all communication must happen before the first trick.`
		);
		expect(brief(14)[1]).toBe(
			`One volunteer takes all tasks: everyone answers yes or no, then the ${C} picks the volunteer. Played untimed.`
		);
		expect(brief(26)[1]).toBe(
			`Two volunteers share the tasks, each taking at least one: everyone answers yes or no, then the ${C} picks them and assigns the tasks. Played untimed.`
		);
		for (const n of [17, 28, 29, 30, 31, 33, 41]) {
			expect(brief(n)).toContain(
				`Free selection: decide together who takes which task, without talking about your hands. The ${C} assigns the tasks in the app.`
			);
		}
	});

	it('names the special rules', () => {
		expect(brief(19)).toContain(
			`The ${C}'s first task must be one of the displayed tasks with the highest difficulty.`
		);
		expect(brief(25)).toContain(`The ${C} is skipped in the selection and takes no task.`);
		expect(brief(8)).toEqual([
			'No task cards.',
			'The crew: nobody may ever have won two more 9s than another member.'
		]);
		expect(brief(12)).toEqual([
			'No task cards.',
			`The crew: don't open a trick with a pink card or a ${T}.`
		]);
		expect(brief(27)).toContain(
			'The crew: the yellow 5 must be the last card played in the last trick.'
		);
		expect(brief(41)).toContain(
			'Epilogue: attempts are not counted in the logbook; the distress signal still is.'
		);
	});

	it('lists the fixed tasks of mission 32 as their cards read', () => {
		const texts = [74, 88, 85, 80].map((id) => {
			const o = TASK_CARDS[id - 1].objective;
			return taskText(o, 3).replace(/^./, (c) => c.toLowerCase());
		});
		expect(brief(32)[0]).toBe(`Four fixed tasks, taken as usual: ${texts.join('; ')}.`);
	});

	it('has a briefing for every mission', () => {
		expect(DEEP_SEA_MISSIONS).toHaveLength(41);
		for (const m of DEEP_SEA_MISSIONS) {
			const lines = deepSeaBriefing(m);
			expect(lines.length, `mission ${m.number}`).toBeGreaterThan(0);
			for (const l of lines.slice(1)) expect(l.endsWith('.'), l).toBe(true);
		}
	});
});

describe('stepText', () => {
	it('words the Deep Sea steps', () => {
		const free = mission(17);
		const split = mission(26);
		expect(stepText('assign', '', free)).toBe(`The ${C} assigns every task to a crew member.`);
		expect(stepText('assign', '', split)).toBe(
			`The ${C} splits the tasks between the volunteers, each taking at least one.`
		);
		expect(stepText('predict', '', null)).toBe(
			'The owners of the prediction tasks note how many tricks they will win.'
		);
		expect(stepText('pick', 'oneMember', mission(6))).toBe(
			`The ${C} picks who takes all the tasks.`
		);
		expect(stepText('pick', 'volunteers', mission(14))).toBe(
			`The ${C} picks the volunteer who takes all the tasks.`
		);
		expect(stepText('pick', 'volunteers', split)).toBe(
			`The ${C} picks the two volunteers who share the tasks.`
		);
		expect(stepText('poll', '', mission(14))).toBe(
			'Who is willing to take the tasks? Everyone answers yes or no.'
		);
		expect(stepText('poll', '', null)).toBe(`The ${C} asks the crew.`);
	});
});

describe('resultText', () => {
	const members: Member[] = [fixtureMember(0), fixtureMember(1)];
	const failed = {
		success: false,
		reason: 'taskFailed' as const,
		task: 0,
		objective: -1,
		trick: 5,
		seat: -1
	};
	const task = (objective: Objective, owner: number) => ({ card: null, objective, owner });

	it('names the owner and the task of a failed Deep Sea task', () => {
		expect(resultText(failed, members, 0, [task(TASK_CARDS[16 - 1].objective, 1)], [])).toBe(
			'Failed in trick 5: Bob: win the pink 3.'
		);
		expect(resultText(failed, members, 0, [task(TASK_CARDS[83 - 1].objective, 0)], [])).toBe(
			'Failed in trick 5: You: win exactly one trick.'
		);
	});

	it('words a failed crew objective of Deep Sea', () => {
		const o = { ...mission(8).objectives[0], owner: -1 };
		expect(
			resultText(
				{ ...failed, reason: 'objectiveFailed', task: -1, objective: 0, trick: 3 },
				members,
				0,
				[],
				[o]
			)
		).toBe(
			'Failed in trick 3: The crew: nobody may ever have won two more 9s than another member.'
		);
	});
});
