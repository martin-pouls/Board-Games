import { describe, expect, it } from 'vitest';
import { cap, terms } from '#lib/edition.js';
import { TASK_CARDS } from './taskCards.fixture.js';
import { chipText, taskChip } from './taskChip.js';

// The chips in words (chipText), by card number. Run with `--mode deepSea` for the Deep Sea terms.
const T = terms.trump;
const C = terms.commander;

const EXPECTED: Record<number, string> = {
	1: 'Most tricks',
	2: 'More than all others',
	3: 'Fewest tricks',
	4: `More than the ${C}`,
	5: `Fewer than the ${C}`,
	6: `As many as the ${C}`,
	7: 'All under 7',
	8: 'All over 5',
	9: 'Win with 6',
	10: 'Win with 5',
	11: 'Win with 3',
	12: '7 takes 5',
	13: '4 takes 8',
	14: '6 takes 6',
	15: 'Win with 2',
	16: 'Pink 3',
	17: 'Yellow 1',
	18: 'Blue 4',
	19: 'Green 6',
	20: '4× 3',
	21: 'At least 3× 5',
	22: 'At least 3× 9',
	23: 'At least 2× 7',
	24: '4× 9',
	25: 'Exactly 3× 6',
	26: 'Exactly 2× 9',
	27: 'Blue 1 blue 2 blue 3',
	28: 'Blue 6 yellow 7',
	29: 'Pink 5 yellow 6',
	30: 'Green 5 blue 8',
	31: 'Blue 5 pink 8',
	32: 'Pink 9 yellow 8',
	33: 'Pink 1 green 7',
	34: 'Yellow 9 blue 7',
	35: 'Green 3 yellow 4 yellow 5',
	36: 'Green 2 in last',
	37: 'Exactly 1× pink 1× green',
	38: 'At least 7× yellow',
	39: 'At least 5× pink',
	40: 'Exactly 2× green',
	41: 'Exactly 2× blue',
	42: 'Exactly 1× pink',
	43: 'No pink',
	44: 'Each colour',
	45: 'A whole colour',
	46: 'All even',
	47: 'All odd',
	48: 'Total 23 or more',
	49: 'Total under 8',
	50: 'Total 22 or 23',
	51: `Exactly 1× ${T}`,
	52: `Only ${T} 1`,
	53: `Only ${T} 2`,
	54: `${cap(T)} 3`,
	55: `Exactly 2× ${T}`,
	56: `Exactly 3× ${T}`,
	57: `No ${T}`,
	58: `${cap(T)} takes pink 7`,
	59: `${cap(T)} takes green 9`,
	60: 'Never lead pink yellow blue',
	61: 'Never lead pink green',
	62: 'No green',
	63: 'No yellow',
	64: 'No pink blue',
	65: 'No yellow green',
	66: 'No 8 9',
	67: 'No 9',
	68: 'No 5',
	69: 'No 1',
	70: 'No 1 2 3',
	71: 'None of 1–4',
	72: 'None of 1–3',
	73: 'None of 1–5',
	74: '0 tricks',
	75: 'Never 2 in a row',
	76: 'Last trick',
	77: 'Tricks 1–3',
	78: 'Tricks 1–2',
	79: 'Trick 1',
	80: 'First + last',
	81: 'Only the last',
	82: 'Only the first',
	83: 'Exactly 1 trick',
	84: 'Exactly 2 tricks',
	85: '2 in a row',
	86: '3 in a row',
	87: 'Exactly 4 tricks',
	88: 'Exactly 3, in a row',
	89: 'Exactly 2, in a row',
	90: 'Predict tricks',
	91: 'Predict, secret',
	92: 'Pink = yellow',
	93: 'Green = yellow in a trick',
	94: 'Pink = blue in a trick',
	95: 'Yellow > blue',
	96: 'Pink > green'
};

describe('taskChip', () => {
	it.each(TASK_CARDS.map((c) => c.id))('card %i', (id) => {
		const row = TASK_CARDS.find((c) => c.id === id)!;
		const o = { ...row.objective, count: row.countBy ? row.countBy[0] : row.objective.count };
		expect(chipText(taskChip(o))).toBe(EXPECTED[id]);
	});

	it('has a chip for every card', () => {
		expect(TASK_CARDS).toHaveLength(96);
		expect(Object.keys(EXPECTED)).toHaveLength(96);
	});

	it('shows a noted prediction, or a hidden one as "?"', () => {
		const said = TASK_CARDS.find((c) => c.id === 90)!.objective;
		const secret = TASK_CARDS.find((c) => c.id === 91)!.objective;
		expect(chipText(taskChip({ ...said, count: 3 }))).toBe('3 tricks, said');
		expect(chipText(taskChip({ ...secret, count: 1 }))).toBe('1 trick, secret');
		expect(chipText(taskChip({ ...secret, count: -1 }, true))).toBe('? tricks, secret');
	});

	it('keeps the counts by crew size (card 48 with 5 members)', () => {
		const row = TASK_CARDS.find((c) => c.id === 48)!;
		expect(chipText(taskChip({ ...row.objective, count: 31 }))).toBe('Total 31 or more');
	});
});
