import { describe, expect, it } from 'vitest';
import { terms } from '#lib/edition.js';
import type { Objective, Task } from '#lib/types.js';
import { TASK_CARDS } from './taskCards.fixture.js';
import { taskText, taskTextOf } from './taskText.js';

// The words of the UI differ per edition (rocket / submarine, commander / captain), so the
// expectations use the terms; run with `--mode deepSea` to see the Deep Sea words.
const T = terms.trump;
const Ts = terms.trumps;
const C = terms.commander;

/** The "Task" column of deep-sea/rules.md §9, in the app's words, by card number. */
const EXPECTED: Record<number, string> = {
	1: 'Win more tricks than everyone else',
	2: 'Win more tricks than everyone else together',
	3: 'Win fewer tricks than everyone else',
	4: `Win more tricks than the ${C}`,
	5: `Win fewer tricks than the ${C}`,
	6: `Win the same number of tricks as the ${C}`,
	7: 'Win a trick where all cards are lower than 7',
	8: 'Win a trick where all cards are greater than 5',
	9: 'Win a trick with a 6',
	10: 'Win a trick with a 5',
	11: 'Win a trick with a 3',
	12: 'Win a 5 with a 7',
	13: 'Win an 8 with a 4',
	14: 'Win any 6 with another 6',
	15: 'Win a trick with a 2',
	16: 'Win the pink 3',
	17: 'Win the yellow 1',
	18: 'Win the blue 4',
	19: 'Win the green 6',
	20: 'Win all four 3s',
	21: 'Win at least three 5s',
	22: 'Win at least three 9s',
	23: 'Win at least two 7s',
	24: 'Win all four 9s',
	25: 'Win exactly three 6s',
	26: 'Win exactly two 9s',
	27: 'Win the blue 1, 2 and 3',
	28: 'Win the blue 6 and the yellow 7',
	29: 'Win the pink 5 and the yellow 6',
	30: 'Win the green 5 and the blue 8',
	31: 'Win the blue 5 and the pink 8',
	32: 'Win the pink 9 and the yellow 8',
	33: 'Win the pink 1 and the green 7',
	34: 'Win the yellow 9 and the blue 7',
	35: 'Win the green 3, the yellow 4 and the yellow 5',
	36: 'Win the green 2 in the last trick',
	37: 'Win exactly one pink and one green card',
	38: 'Win at least seven yellow cards',
	39: 'Win at least five pink cards',
	40: 'Win exactly two green cards',
	41: 'Win exactly two blue cards',
	42: 'Win exactly one pink card',
	43: "Don't win any pink cards",
	44: 'Win at least one card of each colour',
	45: 'Win all cards of at least one colour',
	46: 'Win a trick that has only even numbers',
	47: 'Win a trick that has only odd numbers',
	48: 'Win a trick with a total value of at least 23',
	49: 'Win a trick with a total value lower than 8',
	50: 'Win a trick with a total value of 22 or 23',
	51: `Win exactly one ${T}`,
	52: `Win ${T} 1 and no other ${T}`,
	53: `Win ${T} 2 and no other ${T}`,
	54: `Win ${T} 3`,
	55: `Win exactly two ${Ts}`,
	56: `Win exactly three ${Ts}`,
	57: `Don't win any ${Ts}`,
	58: `Win the pink 7 with a ${T}`,
	59: `Win the green 9 with a ${T}`,
	60: "Don't open a trick with a pink, yellow or blue card",
	61: "Don't open a trick with a pink or green card",
	62: "Don't win any green cards",
	63: "Don't win any yellow cards",
	64: "Don't win any pink or blue cards",
	65: "Don't win any yellow or green cards",
	66: "Don't win any 8s or 9s",
	67: "Don't win any 9s",
	68: "Don't win any 5s",
	69: "Don't win any 1s",
	70: "Don't win any 1s, 2s or 3s",
	71: "Don't win any of the first four tricks",
	72: "Don't win any of the first three tricks",
	73: "Don't win any of the first five tricks",
	74: "Don't win any tricks",
	75: "Don't win two consecutive tricks",
	76: 'Win the last trick',
	77: 'Win the first three tricks',
	78: 'Win the first two tricks',
	79: 'Win the first trick',
	80: 'Win the first and the last trick',
	81: 'Win only the last trick',
	82: 'Win only the first trick',
	83: 'Win exactly one trick',
	84: 'Win exactly two tricks',
	85: 'Win two consecutive tricks',
	86: 'Win three consecutive tricks',
	87: 'Win exactly four tricks',
	88: 'Win exactly three tricks, all in a row',
	89: 'Win exactly two tricks, both in a row',
	90: 'Predict how many tricks you will win (0 to 13)',
	91: 'Predict how many tricks you will win (0 to 13) and keep it secret',
	92: 'Win the same number of pink and yellow cards (more than 0)',
	93: 'Win the same number of green and yellow cards in one trick (more than 0)',
	94: 'Win the same number of pink and blue cards in one trick (more than 0)',
	95: 'Win more yellow cards than blue cards',
	96: 'Win more pink cards than green cards'
};

/** Cards 48 and 49 by crew size: the thresholds of rules.md §9. */
const SUM_AT_LEAST = { 3: 23, 4: 28, 5: 31 } as const;
const SUM_BELOW = { 3: 8, 4: 12, 5: 16 } as const;

/** The objective as the server sends it to a crew of `players`. */
function sent(row: (typeof TASK_CARDS)[number], players: number): Objective {
	const o = structuredClone(row.objective);
	if (row.countBy) o.count = row.countBy[players - 3];
	return o;
}

describe('the task card fixture', () => {
	it('has the 96 cards in order, each with a difficulty for 3, 4 and 5 members', () => {
		expect(TASK_CARDS).toHaveLength(96);
		expect(TASK_CARDS.map((c) => c.id)).toEqual(Array.from({ length: 96 }, (_, i) => i + 1));
		for (const c of TASK_CARDS) expect(c.difficulty.every((d) => d >= 1 && d <= 6)).toBe(true);
		expect(Object.keys(EXPECTED)).toHaveLength(96);
	});
});

describe('taskText over all 96 cards', () => {
	it.each(TASK_CARDS.map((c) => [c.id, c] as const))('card %i', (id, row) => {
		expect(taskText(sent(row, 3), 3)).toBe(EXPECTED[id]);
	});

	it('gives every card its own text', () => {
		const texts = TASK_CARDS.map((c) => taskText(sent(c, 3), 3));
		expect(new Set(texts).size).toBe(96);
	});

	it.each([3, 4, 5] as const)('uses the thresholds of cards 48 and 49 for %i members', (n) => {
		const row = (id: number) => TASK_CARDS[id - 1];
		expect(taskText(sent(row(48), n), n)).toBe(
			`Win a trick with a total value of at least ${SUM_AT_LEAST[n]}`
		);
		expect(taskText(sent(row(49), n), n)).toBe(
			`Win a trick with a total value lower than ${SUM_BELOW[n]}`
		);
	});

	it('starts every text with a capital and ends it without a full stop', () => {
		for (const c of TASK_CARDS) {
			const text = taskText(sent(c, 3), 3);
			expect(text[0]).toBe(text[0].toUpperCase());
			expect(text.endsWith('.')).toBe(false);
		}
	});
});

describe('predictions', () => {
	const open = TASK_CARDS[89].objective;
	const hidden = TASK_CARDS[90].objective;
	const made = (o: Objective, count: number): Objective => ({ ...o, count });

	it('asks for the number before it is made, with the trick count of the crew', () => {
		expect(taskText(open, 3)).toBe('Predict how many tricks you will win (0 to 13)');
		expect(taskText(open, 4)).toBe('Predict how many tricks you will win (0 to 10)');
		expect(taskText(open, 5)).toBe('Predict how many tricks you will win (0 to 8)');
		expect(taskText(hidden, 5)).toBe(
			'Predict how many tricks you will win (0 to 8) and keep it secret'
		);
	});

	it('shows an open prediction', () => {
		expect(taskText(made(open, 3), 3)).toBe('Win exactly 3 tricks');
		expect(taskText(made(open, 1), 3)).toBe('Win exactly one trick');
		expect(taskText(made(open, 0), 3)).toBe('Win exactly 0 tricks');
	});

	it('shows a hidden prediction to its owner and as a question mark to the others', () => {
		expect(taskText(made(hidden, 2), 3)).toBe('Win exactly 2 tricks (prediction hidden)');
		// Another member's hidden prediction arrives with count -1 and predicted true.
		expect(taskText(hidden, 3, true)).toBe('Win exactly ? tricks (prediction hidden)');
	});

	it('reads the prediction flag of a task', () => {
		const task = (over: Partial<Task>): Task => ({
			id: 0,
			card: null,
			token: '',
			owner: 1,
			status: 'open',
			faceDown: false,
			cardId: 91,
			difficulty: 4,
			objective: hidden,
			predicted: false,
			...over
		});
		expect(taskTextOf(task({}), 3)).toBe(
			'Predict how many tricks you will win (0 to 13) and keep it secret'
		);
		expect(taskTextOf(task({ predicted: true }), 3)).toBe(
			'Win exactly ? tricks (prediction hidden)'
		);
		expect(taskTextOf(task({ objective: made(hidden, 4), predicted: true }), 3)).toBe(
			'Win exactly 4 tricks (prediction hidden)'
		);
	});

	it('shows a face-down task by its difficulty', () => {
		const faceDown: Task = {
			id: 0,
			card: null,
			token: '',
			owner: 0,
			status: 'open',
			faceDown: true,
			cardId: 0,
			difficulty: 3,
			objective: null,
			predicted: false
		};
		expect(taskTextOf(faceDown, 3)).toBe('Face-down task (difficulty 3)');
	});
});
