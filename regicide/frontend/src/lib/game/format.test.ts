import { describe, expect, it } from 'vitest';
import {
	cardLabel,
	cardName,
	discardProblemText,
	gradeLabel,
	isRed,
	lossText,
	plural,
	playStacks,
	playProblemText,
	rankLabel,
	resultHeadline,
	suitGlyph,
	suitPowerName,
	who
} from './format.js';
import { checkDiscard } from './rules.js';
import { card, cards, lostView, makeView, playView, wonSoloView } from './fixtures.js';

describe('cards', () => {
	it.each([
		[0, 'Jester'],
		[1, 'A'],
		[2, '2'],
		[10, '10'],
		[11, 'J'],
		[12, 'Q'],
		[13, 'K']
	])('rank %i is %s', (r, want) => {
		expect(rankLabel(r)).toBe(want);
	});

	it('labels and names cards', () => {
		expect(cardLabel(card('5H'))).toBe('5♥');
		expect(cardLabel(card('AS'))).toBe('A♠');
		expect(cardLabel(card('10D'))).toBe('10♦');
		expect(cardLabel(card('KC'))).toBe('K♣');
		expect(cardLabel(card('X'))).toBe('Jester');
		expect(cardName(card('5H'))).toBe('5 of Hearts');
		expect(cardName(card('QS'))).toBe('Queen of Spades');
		expect(cardName(card('AD'))).toBe('Ace of Diamonds');
		expect(cardName(card('X'))).toBe('Jester');
	});

	it('knows glyphs, colours and powers', () => {
		expect(suitGlyph('spades')).toBe('♠');
		expect(isRed('hearts')).toBe(true);
		expect(isRed('diamonds')).toBe(true);
		expect(isRed('clubs')).toBe(false);
		expect(isRed('')).toBe(false);
		expect(suitPowerName('hearts')).toBe('Heal');
		expect(suitPowerName('diamonds')).toBe('Draw');
		expect(suitPowerName('clubs')).toBe('Double damage');
		expect(suitPowerName('spades')).toBe('Shield');
	});

	it('stacks the plays per player', () => {
		const play = (playerId: string, cs: string) => ({
			playerId,
			cards: cards(cs),
			value: 0,
			damage: 0,
			doubled: false,
			blockedSuits: [],
			jester: false
		});
		expect(playStacks([play('p2', '2S'), play('p1', '5H 5C'), play('p2', 'X')])).toEqual([
			{ playerId: 'p2', cards: cards('2S X') },
			{ playerId: 'p1', cards: cards('5H 5C') }
		]);
		expect(playStacks([])).toEqual([]);
	});

	it('pluralises', () => {
		expect(plural(1, 'card')).toBe('1 card');
		expect(plural(0, 'card')).toBe('0 cards');
		expect(plural(5, 'card')).toBe('5 cards');
	});
});

describe('labels and problems', () => {
	it('explains every play problem', () => {
		for (const p of [
			'empty',
			'jesterNotAlone',
			'companionPair',
			'royalCombo',
			'mixedRanks',
			'tooManyCards',
			'comboTooHigh'
		] as const) {
			expect(playProblemText(p).length).toBeGreaterThan(0);
		}
	});

	it('explains discard selections', () => {
		const t = (s: string, req: number) => {
			const cs = cards(s);
			return discardProblemText(checkDiscard(cs, req), req, cs);
		};
		expect(t('', 15)).toBe('Select cards worth at least 15.');
		expect(t('JS', 15)).toBe('Need 5 more (10 of 15).');
		expect(t('QS', 15)).toBe('15 covers 15 damage.');
		expect(t('KS 2H', 15)).toBe("Too many cards: 2♥ isn't needed.");
	});

	it('labels grades and headlines', () => {
		expect(gradeLabel('gold')).toBe('Gold victory');
		expect(resultHeadline(wonSoloView.result!)).toBe('Silver victory! The realm is saved.');
		expect(resultHeadline({ outcome: 'won', enemiesDefeated: 12, soloJestersUsed: 0 })).toBe(
			'Victory! The realm is saved.'
		);
		expect(resultHeadline(lostView.result!)).toBe('Defeat: 5 of 12 enemies slain.');
	});
});

describe('who and lossText', () => {
	it('names players', () => {
		expect(who(playView, 'p1')).toBe('You');
		expect(who(playView, 'p2')).toBe('Ben');
		expect(who(playView, 'zzz')).toBe('Someone');
	});

	it('explains losses', () => {
		const lost = (reason: 'cannotCoverDamage' | 'cannotAct' | 'gaveUp', by: string) =>
			makeView({
				...lostView,
				result: { outcome: 'lost', reason, playerId: by, enemiesDefeated: 3, soloJestersUsed: 0 }
			});
		expect(lossText(lost('cannotCoverDamage', 'p2'))).toBe("Ben couldn't take 15 damage.");
		expect(lossText(lost('cannotCoverDamage', 'p1'))).toBe("You couldn't take 15 damage.");
		expect(lossText(lost('cannotAct', 'p2'))).toBe("Ben had no cards and couldn't yield.");
		expect(lossText(lost('gaveUp', 'p2'))).toBe('Ben gave up.');
		expect(lossText(playView)).toBe('');
	});
});
