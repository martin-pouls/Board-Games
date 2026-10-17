// The 41 Deep Sea mission rows as the server sends them (generated once from the domain's
// missions_deepsea.go through the wire mapping; rules.md §8 is the source).
import { OWNER_UNASSIGNED, type MissionDef } from '#lib/types.js';
import { def, spec } from './fixtures.js';

export const DEEP_SEA_MISSIONS: MissionDef[] = [
	def(1, { poll: 'yesNo', distribution: 'draft', difficulty: 1 }),
	def(2, { poll: 'yesNo', distribution: 'draft', difficulty: 2 }),
	def(3, { poll: 'yesNo', distribution: 'draft', difficulty: 3 }),
	def(4, { poll: 'yesNo', distribution: 'draft', difficulty: 4 }),
	def(5, { poll: 'yesNo', distribution: 'draft', difficulty: 5 }),
	def(6, { poll: 'yesNo', distribution: 'oneMember', difficulty: 5 }),
	def(7, { poll: 'yesNo', distribution: 'draft', difficulty: 6 }),
	def(8, {
		poll: 'yesNo',
		objectives: [spec('crewBalance', 'crew', { owner: OWNER_UNASSIGNED, count: 1, rank: 9 })]
	}),
	def(9, { poll: 'yesNo', distribution: 'draft', commMode: 'currents', difficulty: 7 }),
	def(10, { poll: 'yesNo', distribution: 'captainOrVolunteer', difficulty: 4 }),
	def(11, { poll: 'yesNo', distribution: 'draft', commMode: 'rapture', difficulty: 8 }),
	def(12, {
		poll: 'yesNo',
		objectives: [spec('neverLead', 'crew', { owner: OWNER_UNASSIGNED, suits: ['pink', 'trump'] })]
	}),
	def(13, { poll: 'yesNo', distribution: 'captainOrVolunteer', difficulty: 5 }),
	def(14, {
		poll: 'yesNo',
		distribution: 'volunteers',
		commMode: 'currents',
		difficulty: 6,
		volunteers: 1
	}),
	def(15, {
		poll: 'yesNo',
		distribution: 'volunteers',
		commMode: 'rapture',
		difficulty: 6,
		volunteers: 1
	}),
	def(16, {
		poll: 'yesNo',
		distribution: 'volunteers',
		commMode: 'none',
		difficulty: 6,
		volunteers: 1
	}),
	def(17, { poll: 'yesNo', distribution: 'free', difficulty: 9 }),
	def(18, { poll: 'yesNo', distribution: 'draft', difficulty: 9 }),
	def(19, { poll: 'yesNo', distribution: 'draft', difficulty: 9, hardestFirst: true }),
	def(20, { poll: 'yesNo', distribution: 'draft', commMode: 'unfamiliar', difficulty: 10 }),
	def(21, {
		poll: 'yesNo',
		objectives: [spec('crewBalance', 'crew', { owner: OWNER_UNASSIGNED, count: 1, rank: 1 })],
		commMode: 'unfamiliar'
	}),
	def(22, { poll: 'yesNo', distribution: 'draft', commMode: 'unfamiliar', difficulty: 11 }),
	def(23, {
		poll: 'yesNo',
		fromTrick: 2,
		objectives: [spec('leaderStaysAhead', 'crew', { owner: OWNER_UNASSIGNED })],
		commMode: 'unfamiliar'
	}),
	def(24, { poll: 'yesNo', distribution: 'draft', commMode: 'unfamiliar', difficulty: 12 }),
	def(25, {
		poll: 'yesNo',
		distribution: 'draft',
		commMode: 'unfamiliar',
		difficulty: 12,
		captainSkipped: true
	}),
	def(26, {
		poll: 'yesNo',
		distribution: 'volunteers',
		commMode: 'unfamiliar',
		difficulty: 12,
		volunteers: 2
	}),
	def(27, {
		poll: 'yesNo',
		distribution: 'draft',
		objectives: [
			spec('lastCardPlayed', 'crew', {
				owner: OWNER_UNASSIGNED,
				cards: [{ suit: 'yellow', rank: 5 }]
			})
		],
		commMode: 'unfamiliar',
		difficulty: 5
	}),
	def(28, { poll: 'yesNo', distribution: 'free', difficulty: 14 }),
	def(29, { poll: 'yesNo', distribution: 'free', difficulty: 15 }),
	def(30, { poll: 'yesNo', distribution: 'free', difficulty: 16 }),
	def(31, { poll: 'yesNo', distribution: 'free', difficulty: 17 }),
	def(32, { poll: 'yesNo', distribution: 'draft', fixedTasks: [74, 88, 85, 80] }),
	def(33, { poll: 'yesNo', distribution: 'free', difficulty: 18, uncounted: true }),
	def(34, { poll: 'yesNo', distribution: 'free', difficulty: 19, uncounted: true }),
	def(35, { poll: 'yesNo', distribution: 'free', difficulty: 20, uncounted: true }),
	def(36, { poll: 'yesNo', distribution: 'free', difficulty: 21, uncounted: true }),
	def(37, { poll: 'yesNo', distribution: 'free', difficulty: 22, uncounted: true }),
	def(38, { poll: 'yesNo', distribution: 'free', difficulty: 23, uncounted: true }),
	def(39, { poll: 'yesNo', distribution: 'free', difficulty: 24, uncounted: true }),
	def(40, { poll: 'yesNo', distribution: 'free', difficulty: 25, uncounted: true }),
	def(41, { poll: 'yesNo', distribution: 'free', difficulty: 26, uncounted: true })
];
