import { describe, expect, it } from 'vitest';
import {
	cardLabel,
	cardName,
	cardsLabel,
	discardProblemText,
	eventText,
	gradeLabel,
	isRed,
	lossText,
	plural,
	playKindLabel,
	playProblemText,
	rankLabel,
	resultHeadline,
	suitGlyph,
	suitPowerName,
	who
} from './format.js';
import { checkDiscard } from './rules.js';
import {
	card,
	cards,
	chooseNextOtherView,
	lostView,
	makeView,
	player,
	playView,
	wonSoloView
} from './fixtures.js';
import type { GameEvent } from '#lib/types.js';

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
		expect(cardsLabel(cards('5H 5C'))).toBe('5♥ + 5♣');
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

	it('pluralises', () => {
		expect(plural(1, 'card')).toBe('1 card');
		expect(plural(0, 'card')).toBe('0 cards');
		expect(plural(5, 'card')).toBe('5 cards');
	});
});

describe('labels and problems', () => {
	it('labels kinds', () => {
		expect(playKindLabel('single')).toBe('Single card');
		expect(playKindLabel('combo')).toBe('Combo');
		expect(playKindLabel('companion')).toBe('Animal Companion');
		expect(playKindLabel('jester')).toBe('Jester');
	});

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

describe('eventText', () => {
	const view = chooseNextOtherView; // I am Cara (p3); Anna p1, Ben p2
	const text = (e: GameEvent) => eventText(e, view);

	it('describes plays', () => {
		expect(
			text({
				type: 'cardsPlayed',
				playerId: 'p1',
				cards: cards('5H 5C'),
				value: 10,
				healed: 10,
				drawn: 0,
				damage: 20
			})
		).toBe('Anna played 5♥ + 5♣ (10): healed 10, 20 damage');
		expect(
			text({
				type: 'cardsPlayed',
				playerId: 'p3',
				cards: cards('3D'),
				value: 3,
				healed: 0,
				drawn: 3,
				damage: 3,
				blockedSuits: []
			})
		).toBe('You played 3♦ (3): drew 3, 3 damage');
		expect(
			text({
				type: 'cardsPlayed',
				playerId: 'p2',
				cards: cards('5S'),
				value: 5,
				healed: 0,
				drawn: 0,
				damage: 5,
				blockedSuits: ['spades']
			})
		).toBe('Ben played 5♠ (5): 5 damage, immune to ♠');
	});

	it('describes everything else', () => {
		expect(text({ type: 'jesterPlayed', playerId: 'p2' })).toBe(
			'Ben played the Jester — immunity cancelled'
		);
		expect(text({ type: 'yielded', playerId: 'p2' })).toBe('Ben yielded');
		expect(text({ type: 'enemyDefeated', card: card('JS'), exact: true })).toBe(
			'J♠ defeated — exact! It joins the Tavern'
		);
		expect(text({ type: 'enemyDefeated', card: card('JS'), exact: false })).toBe('J♠ defeated');
		expect(text({ type: 'cardsDiscarded', playerId: 'p3', cards: cards('KD'), attack: 15 })).toBe(
			'You took 15: discarded K♦'
		);
		expect(text({ type: 'attackBlocked', playerId: 'p2' })).toBe('Shield holds: no damage');
		expect(text({ type: 'nextPlayerChosen', byPlayerId: 'p1', playerId: 'p2' })).toBe(
			'Anna chose Ben to go next'
		);
		expect(text({ type: 'nextPlayerChosen', byPlayerId: 'p1', playerId: 'p1' })).toBe(
			'Anna will go next'
		);
		expect(text({ type: 'soloJesterUsed', playerId: 'p1', drawn: 8, left: 1 })).toBe(
			'Jester power used: drew 8 (1 left)'
		);
		expect(text({ type: 'gameWon', grade: 'gold' })).toBe('Gold victory!');
		expect(text({ type: 'gameWon' })).toBe('The last King has fallen!');
		expect(text({ type: 'gameLost', reason: 'gaveUp', playerId: 'p1' })).toBe(
			'The realm has fallen'
		);
	});

	it('is silent for noise events', () => {
		expect(text({ type: 'turnStarted', playerId: 'p1', number: 2 })).toBeNull();
		expect(text({ type: 'enemyRevealed', card: card('QH') })).toBeNull();
		expect(text({ type: 'signalSet', playerId: 'p1', signal: 'goodPlay' })).toBeNull();
		expect(text({ type: 'playerJoined', playerId: 'p4', name: 'Dan' })).toBeNull();
		expect(text({ type: 'rematchCreated', gameId: 'ZZZZ' })).toBeNull();
	});

	it('handles unknown players gracefully', () => {
		const v = makeView({ players: [player('p1', 'Anna', 0, { isYou: true })] });
		expect(eventText({ type: 'yielded', playerId: 'gone' }, v)).toBe('Someone yielded');
	});
});
