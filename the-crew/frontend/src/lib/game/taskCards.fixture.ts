// The 96 Deep Sea task cards as the server sends their objectives (generated once from the
// domain's tasks_deepsea.go through the wire mapping; rules.md §9 is the source). Counts that depend
// on the crew size (cards 48 and 49) are in `countBy`; a prediction has count -1 until it is made.
import { OWNER_UNASSIGNED, type Card, type Objective } from '#lib/types.js';

export interface TaskCardRow {
	id: number;
	/** For 3, 4 and 5 crew members. */
	difficulty: [number, number, number];
	/** When set, the objective's count by crew size. */
	countBy?: [number, number, number];
	/** Re-deal if one hand holds all cards of one entry. */
	redeal?: Card[][];
	objective: Objective;
}

function o(kind: Objective['kind'], over: Partial<Objective> = {}): Objective {
	return {
		kind,
		owner: OWNER_UNASSIGNED,
		cards: [],
		count: 0,
		tricks: [],
		suit: '',
		rank: 0,
		notWithTrump: false,
		ordered: false,
		cmp: '',
		suits: [],
		ranks: [],
		with: null,
		target: null,
		where: '',
		max: 0,
		exact: false,
		inOneTrick: false,
		hidden: false,
		parts: [],
		targets: [],
		without: [],
		...over
	};
}

export const TASK_CARDS: TaskCardRow[] = [
	{ id: 1, difficulty: [2, 3, 3], objective: o('cmpTricks', { cmp: 'more' }) },
	{ id: 2, difficulty: [3, 4, 5], objective: o('cmpTricks', { cmp: 'moreThanAll' }) },
	{ id: 3, difficulty: [2, 2, 3], objective: o('cmpTricks', { cmp: 'fewer' }) },
	{ id: 4, difficulty: [2, 2, 3], objective: o('cmpCaptain', { cmp: 'more' }) },
	{ id: 5, difficulty: [2, 2, 2], objective: o('cmpCaptain', { cmp: 'fewer' }) },
	{ id: 6, difficulty: [4, 3, 3], objective: o('cmpCaptain', { cmp: 'equal' }) },
	{ id: 7, difficulty: [2, 3, 3], objective: o('winTrickWhere', { count: 7, where: 'below' }) },
	{ id: 8, difficulty: [2, 3, 4], objective: o('winTrickWhere', { count: 5, where: 'above' }) },
	{ id: 9, difficulty: [2, 3, 3], objective: o('winWith', { with: { suit: '', rank: 6 } }) },
	{ id: 10, difficulty: [2, 3, 4], objective: o('winWith', { with: { suit: '', rank: 5 } }) },
	{ id: 11, difficulty: [3, 4, 5], objective: o('winWith', { with: { suit: '', rank: 3 } }) },
	{
		id: 12,
		difficulty: [1, 2, 2],
		objective: o('winWith', { with: { suit: '', rank: 7 }, target: { suit: '', rank: 5 } })
	},
	{
		id: 13,
		difficulty: [3, 4, 5],
		objective: o('winWith', { with: { suit: '', rank: 4 }, target: { suit: '', rank: 8 } })
	},
	{
		id: 14,
		difficulty: [2, 3, 4],
		objective: o('winWith', { with: { suit: '', rank: 6 }, target: { suit: '', rank: 6 } })
	},
	{ id: 15, difficulty: [3, 4, 5], objective: o('winWith', { with: { suit: '', rank: 2 } }) },
	{
		id: 16,
		difficulty: [1, 1, 1],
		objective: o('winCards', { cards: [{ suit: 'pink', rank: 3 }] })
	},
	{
		id: 17,
		difficulty: [1, 1, 1],
		objective: o('winCards', { cards: [{ suit: 'yellow', rank: 1 }] })
	},
	{
		id: 18,
		difficulty: [1, 1, 1],
		objective: o('winCards', { cards: [{ suit: 'blue', rank: 4 }] })
	},
	{
		id: 19,
		difficulty: [1, 1, 1],
		objective: o('winCards', { cards: [{ suit: 'green', rank: 6 }] })
	},
	{
		id: 20,
		difficulty: [3, 4, 5],
		objective: o('winCount', { count: 4, cmp: 'exactly', ranks: [3] })
	},
	{
		id: 21,
		difficulty: [3, 4, 5],
		objective: o('winCount', { count: 3, cmp: 'atLeast', ranks: [5] })
	},
	{
		id: 22,
		difficulty: [3, 4, 5],
		objective: o('winCount', { count: 3, cmp: 'atLeast', ranks: [9] })
	},
	{
		id: 23,
		difficulty: [2, 2, 2],
		objective: o('winCount', { count: 2, cmp: 'atLeast', ranks: [7] })
	},
	{
		id: 24,
		difficulty: [4, 5, 6],
		objective: o('winCount', { count: 4, cmp: 'exactly', ranks: [9] })
	},
	{
		id: 25,
		difficulty: [3, 4, 4],
		objective: o('winCount', { count: 3, cmp: 'exactly', ranks: [6] })
	},
	{
		id: 26,
		difficulty: [2, 3, 3],
		objective: o('winCount', { count: 2, cmp: 'exactly', ranks: [9] })
	},
	{
		id: 27,
		difficulty: [2, 3, 3],
		objective: o('winCards', {
			cards: [
				{ suit: 'blue', rank: 1 },
				{ suit: 'blue', rank: 2 },
				{ suit: 'blue', rank: 3 }
			]
		})
	},
	{
		id: 28,
		difficulty: [2, 2, 3],
		objective: o('winCards', {
			cards: [
				{ suit: 'blue', rank: 6 },
				{ suit: 'yellow', rank: 7 }
			]
		})
	},
	{
		id: 29,
		difficulty: [2, 2, 3],
		objective: o('winCards', {
			cards: [
				{ suit: 'pink', rank: 5 },
				{ suit: 'yellow', rank: 6 }
			]
		})
	},
	{
		id: 30,
		difficulty: [2, 2, 3],
		objective: o('winCards', {
			cards: [
				{ suit: 'green', rank: 5 },
				{ suit: 'blue', rank: 8 }
			]
		})
	},
	{
		id: 31,
		difficulty: [2, 2, 3],
		objective: o('winCards', {
			cards: [
				{ suit: 'blue', rank: 5 },
				{ suit: 'pink', rank: 8 }
			]
		})
	},
	{
		id: 32,
		difficulty: [2, 2, 3],
		objective: o('winCards', {
			cards: [
				{ suit: 'pink', rank: 9 },
				{ suit: 'yellow', rank: 8 }
			]
		})
	},
	{
		id: 33,
		difficulty: [2, 2, 2],
		objective: o('winCards', {
			cards: [
				{ suit: 'pink', rank: 1 },
				{ suit: 'green', rank: 7 }
			]
		})
	},
	{
		id: 34,
		difficulty: [2, 3, 3],
		objective: o('winCards', {
			cards: [
				{ suit: 'yellow', rank: 9 },
				{ suit: 'blue', rank: 7 }
			]
		})
	},
	{
		id: 35,
		difficulty: [3, 4, 4],
		objective: o('winCards', {
			cards: [
				{ suit: 'green', rank: 3 },
				{ suit: 'yellow', rank: 4 },
				{ suit: 'yellow', rank: 5 }
			]
		})
	},
	{
		id: 36,
		difficulty: [3, 4, 5],
		objective: o('all', {
			parts: [
				o('winCards', { owner: 0, cards: [{ suit: 'green', rank: 2 }] }),
				o('cardInLastTrick', { owner: 0, cards: [{ suit: 'green', rank: 2 }] })
			]
		})
	},
	{
		id: 37,
		difficulty: [4, 4, 4],
		objective: o('all', {
			parts: [
				o('winCount', { owner: 0, count: 1, cmp: 'exactly', suits: ['pink'] }),
				o('winCount', { owner: 0, count: 1, cmp: 'exactly', suits: ['green'] })
			]
		})
	},
	{
		id: 38,
		difficulty: [3, 3, 3],
		objective: o('winCount', { count: 7, cmp: 'atLeast', suits: ['yellow'] })
	},
	{
		id: 39,
		difficulty: [2, 3, 3],
		objective: o('winCount', { count: 5, cmp: 'atLeast', suits: ['pink'] })
	},
	{
		id: 40,
		difficulty: [3, 4, 4],
		objective: o('winCount', { count: 2, cmp: 'exactly', suits: ['green'] })
	},
	{
		id: 41,
		difficulty: [3, 4, 4],
		objective: o('winCount', { count: 2, cmp: 'exactly', suits: ['blue'] })
	},
	{
		id: 42,
		difficulty: [3, 3, 4],
		objective: o('winCount', { count: 1, cmp: 'exactly', suits: ['pink'] })
	},
	{ id: 43, difficulty: [2, 2, 2], objective: o('winCount', { cmp: 'exactly', suits: ['pink'] }) },
	{ id: 44, difficulty: [2, 3, 4], objective: o('winEachColour') },
	{ id: 45, difficulty: [3, 4, 5], objective: o('winAllOfAnyColour') },
	{ id: 46, difficulty: [2, 5, 6], objective: o('winTrickWhere', { where: 'even' }) },
	{ id: 47, difficulty: [2, 4, 5], objective: o('winTrickWhere', { where: 'odd' }) },
	{
		id: 48,
		difficulty: [3, 3, 4],
		countBy: [23, 28, 31],
		objective: o('winTrickWhere', { where: 'sumAtLeast' })
	},
	{
		id: 49,
		difficulty: [3, 3, 4],
		countBy: [8, 12, 16],
		objective: o('winTrickWhere', { where: 'sumBelow' })
	},
	{
		id: 50,
		difficulty: [3, 3, 4],
		objective: o('winTrickWhere', { count: 22, where: 'sumBetween', max: 23 })
	},
	{
		id: 51,
		difficulty: [3, 3, 3],
		redeal: [
			[
				{ suit: 'trump', rank: 1 },
				{ suit: 'trump', rank: 2 },
				{ suit: 'trump', rank: 3 },
				{ suit: 'trump', rank: 4 }
			]
		],
		objective: o('winCount', { count: 1, cmp: 'exactly', suits: ['trump'] })
	},
	{
		id: 52,
		difficulty: [3, 3, 3],
		redeal: [
			[
				{ suit: 'trump', rank: 1 },
				{ suit: 'trump', rank: 4 }
			],
			[
				{ suit: 'trump', rank: 1 },
				{ suit: 'trump', rank: 2 },
				{ suit: 'trump', rank: 3 }
			]
		],
		objective: o('all', {
			parts: [
				o('winCount', { owner: 0, count: 1, cmp: 'exactly', suits: ['trump'] }),
				o('winCards', { owner: 0, cards: [{ suit: 'trump', rank: 1 }] })
			]
		})
	},
	{
		id: 53,
		difficulty: [3, 3, 3],
		redeal: [
			[
				{ suit: 'trump', rank: 2 },
				{ suit: 'trump', rank: 4 }
			],
			[
				{ suit: 'trump', rank: 1 },
				{ suit: 'trump', rank: 2 },
				{ suit: 'trump', rank: 3 }
			]
		],
		objective: o('all', {
			parts: [
				o('winCount', { owner: 0, count: 1, cmp: 'exactly', suits: ['trump'] }),
				o('winCards', { owner: 0, cards: [{ suit: 'trump', rank: 2 }] })
			]
		})
	},
	{
		id: 54,
		difficulty: [1, 1, 1],
		objective: o('winCards', { cards: [{ suit: 'trump', rank: 3 }] })
	},
	{
		id: 55,
		difficulty: [3, 3, 4],
		redeal: [
			[
				{ suit: 'trump', rank: 2 },
				{ suit: 'trump', rank: 3 },
				{ suit: 'trump', rank: 4 }
			]
		],
		objective: o('winCount', { count: 2, cmp: 'exactly', suits: ['trump'] })
	},
	{
		id: 56,
		difficulty: [3, 4, 4],
		redeal: [
			[
				{ suit: 'trump', rank: 1 },
				{ suit: 'trump', rank: 2 },
				{ suit: 'trump', rank: 3 },
				{ suit: 'trump', rank: 4 }
			]
		],
		objective: o('winCount', { count: 3, cmp: 'exactly', suits: ['trump'] })
	},
	{ id: 57, difficulty: [1, 1, 1], objective: o('winCount', { cmp: 'exactly', suits: ['trump'] }) },
	{
		id: 58,
		difficulty: [3, 3, 3],
		objective: o('winWith', { with: { suit: 'trump', rank: 0 }, target: { suit: 'pink', rank: 7 } })
	},
	{
		id: 59,
		difficulty: [3, 3, 3],
		objective: o('winWith', {
			with: { suit: 'trump', rank: 0 },
			target: { suit: 'green', rank: 9 }
		})
	},
	{
		id: 60,
		difficulty: [4, 3, 3],
		objective: o('neverLead', { suits: ['pink', 'yellow', 'blue'] })
	},
	{ id: 61, difficulty: [2, 1, 1], objective: o('neverLead', { suits: ['pink', 'green'] }) },
	{ id: 62, difficulty: [2, 2, 2], objective: o('winCount', { cmp: 'exactly', suits: ['green'] }) },
	{
		id: 63,
		difficulty: [2, 2, 2],
		objective: o('winCount', { cmp: 'exactly', suits: ['yellow'] })
	},
	{
		id: 64,
		difficulty: [3, 3, 3],
		objective: o('winCount', { cmp: 'exactly', suits: ['pink', 'blue'] })
	},
	{
		id: 65,
		difficulty: [3, 3, 3],
		objective: o('winCount', { cmp: 'exactly', suits: ['yellow', 'green'] })
	},
	{ id: 66, difficulty: [3, 3, 2], objective: o('winCount', { cmp: 'exactly', ranks: [8, 9] }) },
	{ id: 67, difficulty: [1, 1, 1], objective: o('winCount', { cmp: 'exactly', ranks: [9] }) },
	{ id: 68, difficulty: [1, 2, 2], objective: o('winCount', { cmp: 'exactly', ranks: [5] }) },
	{ id: 69, difficulty: [2, 2, 2], objective: o('winCount', { cmp: 'exactly', ranks: [1] }) },
	{ id: 70, difficulty: [3, 3, 3], objective: o('winCount', { cmp: 'exactly', ranks: [1, 2, 3] }) },
	{ id: 71, difficulty: [1, 2, 3], objective: o('avoidTricks', { tricks: [1, 2, 3, 4] }) },
	{ id: 72, difficulty: [1, 2, 2], objective: o('avoidTricks', { tricks: [1, 2, 3] }) },
	{ id: 73, difficulty: [2, 3, 3], objective: o('avoidTricks', { tricks: [1, 2, 3, 4, 5] }) },
	{ id: 74, difficulty: [4, 3, 3], objective: o('winNoTricks') },
	{ id: 75, difficulty: [3, 2, 2], objective: o('noConsecutive') },
	{ id: 76, difficulty: [2, 3, 3], objective: o('winTricks', { tricks: [-1] }) },
	{ id: 77, difficulty: [2, 3, 4], objective: o('winTricks', { tricks: [1, 2, 3] }) },
	{ id: 78, difficulty: [1, 1, 2], objective: o('winTricks', { tricks: [1, 2] }) },
	{ id: 79, difficulty: [1, 1, 1], objective: o('winTricks', { tricks: [1] }) },
	{ id: 80, difficulty: [3, 4, 4], objective: o('winTricks', { tricks: [1, -1] }) },
	{ id: 81, difficulty: [4, 4, 4], objective: o('winOnlyTricks', { tricks: [-1] }) },
	{ id: 82, difficulty: [4, 3, 3], objective: o('winOnlyTricks', { tricks: [1] }) },
	{ id: 83, difficulty: [3, 2, 2], objective: o('winExactlyTricks', { count: 1 }) },
	{ id: 84, difficulty: [2, 2, 2], objective: o('winExactlyTricks', { count: 2 }) },
	{ id: 85, difficulty: [1, 1, 1], objective: o('consecutive', { count: 2 }) },
	{ id: 86, difficulty: [2, 3, 4], objective: o('consecutive', { count: 3 }) },
	{ id: 87, difficulty: [2, 3, 5], objective: o('winExactlyTricks', { count: 4 }) },
	{ id: 88, difficulty: [3, 3, 4], objective: o('consecutive', { count: 3, exact: true }) },
	{ id: 89, difficulty: [3, 3, 3], objective: o('consecutive', { count: 2, exact: true }) },
	{ id: 90, difficulty: [3, 2, 2], objective: o('predict', { count: -1 }) },
	{ id: 91, difficulty: [4, 3, 3], objective: o('predict', { count: -1, hidden: true }) },
	{ id: 92, difficulty: [4, 4, 4], objective: o('equalSuits', { suits: ['pink', 'yellow'] }) },
	{
		id: 93,
		difficulty: [2, 3, 3],
		objective: o('equalSuits', { suits: ['green', 'yellow'], inOneTrick: true })
	},
	{
		id: 94,
		difficulty: [2, 3, 3],
		objective: o('equalSuits', { suits: ['pink', 'blue'], inOneTrick: true })
	},
	{ id: 95, difficulty: [1, 1, 1], objective: o('moreSuit', { suits: ['yellow', 'blue'] }) },
	{ id: 96, difficulty: [1, 1, 1], objective: o('moreSuit', { suits: ['pink', 'green'] }) }
];
