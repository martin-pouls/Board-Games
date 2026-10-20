import { describe, expect, it } from 'vitest';
import { BoardMarks } from './boardMarks.svelte.js';
import { playingView } from './fixtures.js';
import type { Board } from '#lib/types.js';

function board(change: (b: Board) => void = () => {}): Board {
	const b = structuredClone(playingView().board!);
	change(b);
	return b;
}

describe('BoardMarks', () => {
	it('marks nothing on the first board', () => {
		const m = new BoardMarks();
		m.update(board(), false, 'g');
		expect(m.current).toEqual({ added: {}, covered: [] });
	});

	it('marks added gems and new forest tokens and adds them up across views', () => {
		const m = new BoardMarks();
		m.update(board(), false, 'g');
		const withGems = board((b) => b.spaces.forEach((s) => (s.gems += s.plus ? 1 : 0)));
		m.update(withGems, false, 'g');
		expect(m.current.added).toEqual({ 1: 1, 3: 1, 5: 1, 7: 1, 9: 1 });

		const covered = structuredClone(withGems);
		covered.spaces[10].covered = true;
		covered.spaces[9].gems += covered.spaces[10].gems;
		covered.spaces[10].gems = 0;
		m.update(covered, false, 'g');
		expect(m.current.added).toEqual({ 1: 1, 3: 1, 5: 1, 7: 1, 9: 2 });
		expect(m.current.covered).toEqual([10]);
	});

	it('ignores collected gems', () => {
		const m = new BoardMarks();
		m.update(board(), false, 'g');
		m.update(
			board((b) => b.spaces[3].gems--),
			false,
			'g'
		);
		expect(m.current).toEqual({ added: {}, covered: [] });
	});

	it('ends the marks when a card is led', () => {
		const m = new BoardMarks();
		m.update(board(), false, 'g');
		m.update(
			board((b) => b.spaces[1].gems++),
			false,
			'g'
		);
		m.update(
			board((b) => b.spaces[1].gems++),
			true,
			'g'
		);
		expect(m.current).toEqual({ added: {}, covered: [] });
	});

	it('starts over for another game', () => {
		const m = new BoardMarks();
		m.update(board(), false, 'g');
		m.update(
			board((b) => b.spaces[1].gems++),
			false,
			'h'
		);
		expect(m.current).toEqual({ added: {}, covered: [] });
	});
});
