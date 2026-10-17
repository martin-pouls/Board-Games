import { describe, expect, it } from 'vitest';
import { taskFace, type Seg } from './taskFace.js';
import { TASK_CARDS } from './taskCards.fixture.js';

const card = (id: number) => TASK_CARDS.find((c) => c.id === id)!.objective;
const segs = (id: number, predicted?: boolean): Seg[] =>
	taskFace(card(id), 3, predicted).rows.flat();

describe('taskFace', () => {
	it('draws every one of the 96 cards with a head and at least one row', () => {
		for (const c of TASK_CARDS) {
			const f = taskFace(c.objective, 3);
			expect(f.head, `card ${c.id}`).not.toBe('');
			expect(f.rows.length, `card ${c.id}`).toBeGreaterThan(0);
		}
	});

	it('puts the count badge on the counted card', () => {
		expect(taskFace(card(21), 3)).toEqual({
			head: 'I win at least',
			rows: [[{ t: 'cards', cards: [{ suit: 'any', rank: 5, count: 3 }], join: undefined }]]
		});
		expect(taskFace(card(43), 3).head).toBe('I win no');
	});

	it('shows all four 3s for card 20', () => {
		const f = taskFace(card(20), 3);
		expect(f.head).toBe('I win all');
		expect(f.rows[0][0]).toMatchObject({
			cards: [{ rank: 3 }, { rank: 3 }, { rank: 3 }, { rank: 3 }]
		});
	});

	it('badges both colours of card 37', () => {
		expect(segs(37)).toEqual([
			{
				t: 'cards',
				cards: [
					{ suit: 'pink', rank: 0, count: 1 },
					{ suit: 'green', rank: 0, count: 1 }
				],
				join: undefined
			}
		]);
	});

	it('shows X before a prediction, the number after, ? when hidden from the viewer', () => {
		expect(segs(90, false)[0]).toMatchObject({ text: 'X', big: true });
		expect(segs(90, true)).toEqual([
			{ t: 'text', text: '?', big: true },
			{ t: 'text', text: 'tricks', big: false }
		]);
		const shown = { ...card(90), count: 1 };
		expect(taskFace(shown, 3, true).rows[0]).toMatchObject([{ text: '1' }, { text: 'trick' }]);
	});

	it('joins two lead-ban colours with "or", overlaps three', () => {
		expect(segs(61)).toMatchObject([{ t: 'cards', join: 'or' }]);
		expect(taskFace(card(60), 3).note).toBe('any of these');
	});

	it('lists both sums of card 50', () => {
		expect(taskFace(card(50), 3).rows[1]).toMatchObject([
			{ value: 22 },
			{ text: 'or' },
			{ value: 23 }
		]);
	});
});
