import { describe, expect, it } from 'vitest';
import {
	briefing,
	listText,
	objectiveText,
	ownerText,
	resultText,
	seatName,
	specText,
	trickText
} from './format.js';
import type { Card, Member, MissionDef, Objective, ObjectiveSpec } from '#lib/types.js';

const card = (suit: Card['suit'], rank: number): Card => ({ suit, rank });
const obj = (o: Partial<Objective>): Objective => ({
	kind: 'winCards',
	owner: -1,
	cards: [],
	count: 0,
	tricks: [],
	suit: '',
	rank: 0,
	notWithTrump: false,
	ordered: false,
	...o
});
const member = (seat: number, name: string, dummy = false): Member => ({
	seat,
	playerId: dummy ? '' : name,
	name: dummy ? '' : name,
	dummy,
	cardCount: 0,
	tricksWon: 0,
	communication: null,
	communicated: false,
	stacks: [],
	mustAct: false,
	cards: null
});
const members = [member(0, 'Alice'), member(1, 'Bob'), member(2, '', true)];
const mission = (m: Partial<MissionDef>): MissionDef => ({
	number: 1,
	tasks: 0,
	tokens: [],
	distribution: '',
	deadZone: false,
	fromTrick: 0,
	golden: false,
	setup: [],
	poll: 'yesNo',
	objectives: [],
	...m
});

describe('seatName', () => {
	it('names you, players and JARVIS', () => {
		expect(seatName(members, 0, 0)).toBe('You');
		expect(seatName(members, 1, 0)).toBe('Bob');
		expect(seatName(members, 2, 0)).toBe('JARVIS');
	});
});

describe('listText and trickText', () => {
	it('joins lists', () => {
		expect(listText([])).toBe('');
		expect(listText(['a'])).toBe('a');
		expect(listText(['a', 'b'])).toBe('a and b');
		expect(listText(['a', 'b', 'c'])).toBe('a, b and c');
	});
	it('names the last trick', () => {
		expect(trickText(3)).toBe('trick 3');
		expect(trickText(-1)).toBe('the last trick');
	});
});

describe('objectiveText', () => {
	it.each<[string, Partial<Objective>, string]>([
		['task', { kind: 'winCards', cards: [card('pink', 7)] }, 'win pink 7'],
		['no trick (5)', { kind: 'winNoTricks' }, 'win no trick'],
		[
			'exactly one (33)',
			{ kind: 'winExactlyTricks', count: 1, notWithTrump: true },
			'win exactly one trick, none of them with a rocket'
		],
		[
			'only first and last (41)',
			{ kind: 'winOnlyTricks', tricks: [1, -1], notWithTrump: true },
			'win trick 1 and the last trick and no other trick, none of them with a rocket'
		],
		[
			'only first four (50)',
			{ kind: 'winOnlyTricks', tricks: [1, 2, 3, 4] },
			'win tricks 1, 2, 3 and 4 and no other trick'
		],
		[
			'first and last (34)',
			{ kind: 'winTricks', tricks: [1, -1] },
			'win trick 1 and the last trick'
		],
		['all pink (46)', { kind: 'winAllOfSuit', suit: 'pink' }, 'win all pink cards'],
		[
			'a colour 1 (9)',
			{ kind: 'crewWinWithAny', count: 1, cards: [card('pink', 1), card('blue', 1)] },
			'win one trick with a colour 1'
		],
		[
			'two colour 1s (26)',
			{ kind: 'crewWinWithAny', count: 2, cards: [card('pink', 1)] },
			'win 2 tricks with a colour 1'
		],
		[
			'each rocket in order (44)',
			{ kind: 'crewWinWithEach', ordered: true, cards: [card('trump', 1), card('trump', 2)] },
			'win a trick with each of rocket 1 and rocket 2, in this order'
		],
		['no 9 (16)', { kind: 'crewNeverWinWith', rank: 9 }, 'no trick may be won by a colour 9'],
		[
			'balance (29)',
			{ kind: 'crewBalance', count: 1 },
			'nobody may ever have 2 tricks more than another member'
		],
		[
			'omega last (48)',
			{ kind: 'cardInLastTrick', cards: [card('green', 3)] },
			'win green 3 in the last trick'
		]
	])('%s', (_, o, want) => {
		expect(objectiveText(obj(o))).toBe(want);
	});
});

describe('owners', () => {
	it('names the owner of an objective', () => {
		expect(ownerText(-1, members, 0)).toBe('The crew');
		expect(ownerText(-2, members, 0)).toBe('Not chosen yet');
		expect(ownerText(2, members, 0)).toBe('JARVIS');
	});
	it('names the owner rule in the briefing', () => {
		const s = (o: Partial<ObjectiveSpec>): ObjectiveSpec => ({
			...obj({}),
			ownerRule: 'crew',
			holder: null,
			...o
		});
		expect(specText(s({ kind: 'winNoTricks', ownerRule: 'chosen' }))).toBe(
			'A chosen crew member: win no trick'
		);
		expect(
			specText(
				s({
					kind: 'winAllOfSuit',
					suit: 'pink',
					ownerRule: 'leftOfHolder',
					holder: card('pink', 9)
				})
			)
		).toBe('The player left of whoever holds pink 9: win all pink cards');
	});
});

describe('briefing', () => {
	it('describes tasks, tokens and special rules', () => {
		expect(
			briefing(
				mission({ tasks: 3, distribution: 'open', tokens: ['arrow1', 'arrow2'], deadZone: true })
			)
		).toEqual([
			'3 tasks. Tasks are drafted, the commander first.',
			'Tokens: › ››.',
			'Dead zone: communicate a card without saying whether it is highest, only or lowest.'
		]);
		expect(briefing(mission({ tasks: 5, distribution: 'open', fromTrick: 3 }))[1]).toBe(
			'Disruption: nobody may communicate before trick 3.'
		);
	});
	it('lists objectives of missions without tasks', () => {
		const lines = briefing(
			mission({
				objectives: [
					{ ...obj({ kind: 'crewNeverWinWith', rank: 9 }), ownerRule: 'crew', holder: null }
				]
			})
		);
		expect(lines).toEqual(['The crew: no trick may be won by a colour 9.']);
	});
});

describe('resultText', () => {
	const tasks = [{ card: card('pink', 7) }, { card: card('blue', 2) }];
	const r = (o: object) => ({
		success: false,
		reason: 'completed' as const,
		task: -1,
		objective: -1,
		trick: 0,
		seat: -1,
		...o
	});
	it('explains each reason', () => {
		expect(resultText(r({ success: true }), members, 0, tasks)).toBe(
			'Mission accomplished. Every task is fulfilled.'
		);
		expect(
			resultText(r({ reason: 'taskFailed', task: 1, seat: 1, trick: 4 }), members, 0, tasks)
		).toBe('Task blue 2 failed: Bob won it in trick 4.');
		expect(
			resultText(r({ reason: 'taskFailed', task: 0, seat: 0, trick: 2 }), members, 0, tasks)
		).toBe('Task pink 7 failed: you won it in trick 2.');
		expect(resultText(r({ reason: 'tokenOrder', task: 0, trick: 3 }), members, 0, tasks)).toBe(
			'Task pink 7 was fulfilled out of token order in trick 3.'
		);
		expect(
			resultText(r({ reason: 'objectiveFailed', objective: 0, trick: 5 }), members, 0, tasks)
		).toBe('A special objective failed in trick 5.');
		expect(resultText(r({ reason: 'gaveUp' }), members, 0, tasks)).toBe(
			'The crew gave up this attempt.'
		);
	});
});
