// Texts for objectives, mission briefings and results. Pure functions, unit-tested.
import { cap, edition, terms } from '#lib/edition.js';
import { cardLabel, SUIT_NAMES, tokenLabel } from './cards.js';
import { taskPhrase } from './taskText.js';
import { listText, numberWord, plural } from './words.js';
import {
	LAST_TRICK,
	OWNER_CREW,
	OWNER_UNASSIGNED,
	type AttemptResult,
	type Card,
	type CommMode,
	type MissionDef,
	type Member,
	type Objective,
	type ObjectiveSpec,
	type Task,
	type OwnerRule,
	type PickKind,
	type Position,
	type Step
} from '#lib/types.js';

/** The name of the dummy crew member: JARVIS (Planet Nine) or Tonoja (Deep Sea). */
export const DUMMY = terms.dummy;

/** "You", the player's name, or the dummy's name. */
export function seatName(members: Member[], seat: number, you: number): string {
	if (seat === you) return 'You';
	const m = members.find((m) => m.seat === seat);
	if (!m) return `Seat ${seat + 1}`;
	return m.dummy ? DUMMY : m.name;
}

/** "trick 3", "the last trick". */
export function trickText(n: number): string {
	return n === LAST_TRICK ? 'the last trick' : `trick ${n}`;
}

export { listText };

function cardsText(cards: Card[], and = 'and'): string {
	return listText(cards.map(cardLabel), and);
}

function tricksText(tricks: number[]): string {
	const named = tricks.map(trickText);
	// "tricks 1, 2, 3 and 4", "trick 1 and the last trick"
	if (tricks.every((t) => t !== LAST_TRICK) && tricks.length > 1) {
		return `tricks ${listText(tricks.map(String))}`;
	}
	return listText(named);
}

/**
 * What an objective asks for, without its owner: "win pink 7", "win no trick". Phrased for the
 * owner as subject ("You: win …", "The crew: …").
 */
export function objectiveText(o: Objective, players = 0): string {
	const noTrump = o.notWithTrump ? `, none of them with a ${terms.trump}` : '';
	switch (o.kind) {
		case 'winCards':
			return `win ${cardsText(o.cards)}`;
		case 'winNoTricks':
			return 'win no trick';
		case 'winExactlyTricks':
			// Deep Sea's side quest Q1 adds the ranks no counted trick may contain.
			if (o.without.length > 0) return taskPhrase(o, players);
			return `win exactly ${plural(o.count, 'trick')}${noTrump}`;
		case 'winOnlyTricks':
			return `win ${tricksText(o.tricks)} and no other trick${noTrump}`;
		case 'winTricks':
			return `win ${tricksText(o.tricks)}`;
		case 'winAllOfSuit':
			return `win all ${SUIT_NAMES[o.suit || 'pink']} cards`;
		case 'crewWinWithAny':
			return `win ${plural(o.count, 'trick')} with a colour ${o.cards[0]?.rank ?? 1}`;
		case 'crewWinWithEach':
			return o.ordered
				? `win a trick with each of ${cardsText(o.cards)}, in this order`
				: `win a trick with each of ${cardsText(o.cards)}`;
		case 'crewNeverWinWith':
			return `no trick may be won by a colour ${o.rank}`;
		case 'crewBalance':
			// Deep Sea's missions 8 and 21 count the won cards of a rank instead of tricks.
			if (o.rank > 0) return taskPhrase(o, players);
			return `nobody may ever have ${plural(o.count + 1, 'trick')} more than another member`;
		case 'cardInLastTrick':
			return o.cards.length
				? `win ${cardsText(o.cards)} in the last trick`
				: 'win the Ω task in the last trick';
		default:
			// Deep Sea's kinds.
			return taskPhrase(o, players);
	}
}

/** Who owns an objective in the running attempt. */
export function ownerText(owner: number, members: Member[], you: number): string {
	if (owner === OWNER_CREW) return 'The crew';
	if (owner === OWNER_UNASSIGNED) return 'Not chosen yet';
	return seatName(members, owner, you);
}

const OWNER_RULES: Record<OwnerRule, string> = {
	crew: 'The crew',
	commander: `The ${terms.commander}`,
	chosen: 'A chosen crew member',
	roleA: 'One chosen crew member',
	roleB: 'Another chosen crew member',
	leftOfHolder: ''
};

/** A special objective in the briefing: "A chosen crew member: win no trick". */
export function specText(s: ObjectiveSpec): string {
	if (s.ownerRule === 'leftOfHolder') {
		const holder = s.holder ? cardLabel(s.holder) : 'the card';
		return `The player left of whoever holds ${holder}: ${objectiveText(s)}`;
	}
	return `${OWNER_RULES[s.ownerRule]}: ${objectiveText(s)}`;
}

const DISTRIBUTION: Partial<Record<MissionDef['distribution'], string>> = {
	open: `Tasks are drafted, the ${terms.commander} first.`,
	commanderDecision: `The ${terms.commander} asks who takes the tasks, then decides; one member takes them all.`,
	commanderDistribution: `The ${terms.commander} reveals the tasks one by one and assigns each after asking the crew; the tasks must be shared out as evenly as possible.`
};

/** "7", or "Q1" for the side quests: a mission's name in headings (`Mission ${missionLabel(def)}`). */
export function missionLabel(def: Pick<MissionDef, 'number' | 'label'>): string {
	return def.label || String(def.number);
}

/** The label of mission `n` when only its number is at hand; the table may not be loaded yet. */
export function missionLabelOf(n: number, defs: readonly MissionDef[] | null | undefined): string {
	const def = defs?.find((d) => d.number === n);
	return def ? missionLabel(def) : String(n);
}

/** The briefing of a mission, in the words of this build's edition. */
export function briefing(m: MissionDef): string[] {
	return edition === 'deepSea' ? deepSeaBriefing(m) : planetNineBriefing(m);
}

/** The briefing of a Planet Nine mission: its tasks, tokens and special rules, one line each. */
export function planetNineBriefing(m: MissionDef): string[] {
	const lines: string[] = [];
	if (m.tasks > 0) {
		lines.push(`${plural(m.tasks, 'task')}. ${DISTRIBUTION[m.distribution] ?? ''}`.trim());
	}
	if (m.tokens.length) {
		lines.push(`Tokens: ${m.tokens.map(tokenLabel).join(' ')}.`);
	}
	for (const s of m.objectives) lines.push(`${specText(s)}.`);
	if (m.deadZone) {
		lines.push(
			'Dead zone: communicate a card without saying whether it is highest, only or lowest.'
		);
	}
	if (m.fromTrick > 1) {
		lines.push(`Disruption: nobody may communicate before trick ${m.fromTrick}.`);
	}
	if (m.setup.includes('swapTokens')) {
		lines.push('Before play, the crew may swap the tokens of two tasks.');
	}
	if (m.setup.includes('moveToken')) {
		lines.push('Before play, the crew may move one token to a task without a token.');
	}
	if (m.setup.includes('silenceMember')) {
		lines.push(`The ${terms.commander} picks a crew member who may not communicate.`);
	}
	if (m.setup.includes('drawAfterFirstTrick')) {
		lines.push('After the first trick, everyone draws a card from their right neighbour.');
	}
	if (m.golden) {
		lines.push(
			'Golden frame: with 5 players, one task may be handed to another member before the first trick.'
		);
	}
	return lines;
}

const COMM_NAMES: Record<CommMode, string> = {
	'': '',
	currents: 'Currents',
	rapture: 'Rapture of the deep',
	unfamiliar: 'Unfamiliar terrain',
	none: 'No communication'
};

const COMM_RULES: Record<CommMode, string> = {
	'': '',
	currents: 'Currents: communicate a card without saying whether it is highest, only or lowest.',
	rapture: `Rapture of the deep: the crew shares ${terms.commToken}s, two fewer than there are crew members. Whoever wants to communicate takes one without discussion, and may communicate again while tokens are left.`,
	unfamiliar:
		'Unfamiliar terrain: before each attempt a random colour card picks the mode. 1–3: normal, 4–6: currents, 7–9: rapture of the deep.',
	none: 'Nobody may communicate.'
};

/** The fixed tasks of mission 32 (task card numbers), as their cards read. */
const FIXED_TASKS: Record<number, string> = {
	74: "don't win any tricks",
	88: 'win exactly three tricks, all in a row',
	85: 'win two consecutive tricks',
	80: 'win the first and the last trick'
};

/** "Stage A", "Stage B". */
const stageName = (i: number) => `Stage ${String.fromCharCode(65 + i)}`;

function distributionLines(m: MissionDef): string[] {
	const boss = terms.commander;
	switch (m.distribution) {
		case 'draft':
			return [
				`Tasks are drafted, the ${boss} first, then clockwise. A member may pass only while there are fewer tasks than members; every task must be taken after one round.`
			];
		case 'oneMember':
			return [
				`The ${boss} gives all tasks to one crew member. Decide together, without talking about your hands.`
			];
		case 'captainOrVolunteer':
			return [
				`The ${boss} takes all tasks or passes them to a willing crew member; the crew answers yes or no first. If they are passed on, all communication must happen before the first trick.`
			];
		case 'volunteers':
			return [
				m.volunteers === 1
					? `One volunteer takes all tasks: everyone answers yes or no, then the ${boss} picks the volunteer. Played untimed.`
					: `${cap(numberWord(m.volunteers))} volunteers share the tasks, each taking at least one: everyone answers yes or no, then the ${boss} picks them and assigns the tasks. Played untimed.`
			];
		case 'free':
			return [
				`Free selection: decide together who takes which task, without talking about your hands. The ${boss} assigns the tasks in the app.`
			];
		default:
			return [];
	}
}

/**
 * The briefing of a Deep Sea mission, built from the mission row (rules.md §8): difficulty and
 * communication mode first, then how the tasks are taken, then the special rules.
 */
export function deepSeaBriefing(m: MissionDef): string[] {
	const boss = terms.commander;
	const lines: string[] = [];
	const perSize = m.difficultyBy.some((d) => d > 0);
	const difficulty = perSize
		? `Difficulty ${m.difficultyBy.join(' / ')} (crew of 3 / 4 / 5)`
		: m.difficulty > 0
			? `Difficulty ${m.difficulty}`
			: '';
	const head = [difficulty, COMM_NAMES[m.commMode]].filter(Boolean).join(' · ');
	if (head) lines.push(head);
	if (!perSize && m.difficulty === 0 && m.fixedTasks.length === 0 && m.stages === 0) {
		lines.push('No task cards.');
	}
	if (m.fixedTasks.length > 0) {
		const tasks = m.fixedTasks.map((id) => FIXED_TASKS[id] ?? `task card ${id}`);
		lines.push(
			`${cap(numberWord(tasks.length))} fixed tasks, taken as usual: ${tasks.join('; ')}.`
		);
	}
	if (m.faceDownTasks) {
		lines.push('The task cards stay face down: only their difficulty is visible.');
	}
	if (m.distribution !== '') lines.push(...distributionLines(m));
	if (m.captainSkipped) lines.push(`The ${boss} is skipped in the selection and takes no task.`);
	if (m.hardestFirst) {
		lines.push(
			`The ${boss}'s first task must be one of the displayed tasks with the highest difficulty.`
		);
	}
	const defaultCaptain = m.captainCard.suit === 'trump' && m.captainCard.rank === 4;
	if (!defaultCaptain) {
		lines.push(
			`The holder of ${cardLabel(m.captainCard)} is the ${boss}, not the holder of ${terms.trump} 4.`
		);
	}
	m.stageObjectives.forEach((specs, i) => {
		const when = i === 0 ? '' : ' (a new deal once the stage before is won)';
		lines.push(`${stageName(i)}${when}. ${specs.map(specText).join('; ')}.`);
	});
	if (m.stages > 0) {
		lines.push(
			`If one hand holds all four ${terms.trumps}, the cards are dealt again (not an attempt).`
		);
	}
	for (const s of m.objectives) lines.push(`${specText(s)}.`);
	if (COMM_RULES[m.commMode]) lines.push(COMM_RULES[m.commMode]);
	if (m.fromTrick > 1) lines.push(`Nobody may communicate before trick ${m.fromTrick}.`);
	if (m.deadZone) {
		lines.push(
			'Dead zone: communicate a card without saying whether it is highest, only or lowest.'
		);
	}
	if (m.uncounted) {
		lines.push('Epilogue: attempts are not counted in the logbook; the distress signal still is.');
	}
	return lines;
}

const PICK_PROMPT: Record<Exclude<PickKind, ''>, string> = {
	silence: `The ${terms.commander} picks a crew member who may not communicate.`,
	roles: `The ${terms.commander} picks the two crew members for the special objectives.`,
	objective: `The ${terms.commander} picks who takes the special objective.`,
	decision: `The ${terms.commander} picks who takes all the tasks.`,
	distribute: `The ${terms.commander} assigns the revealed task.`,
	oneMember: `The ${terms.commander} picks who takes all the tasks.`,
	volunteers: `The ${terms.commander} picks the volunteers.`
};

/**
 * A sentence about the current step, for the status line. `def` is the mission row: Deep Sea's
 * polls, picks and assignments read differently per mission.
 */
export function stepText(step: Step, pick: PickKind, def: MissionDef | null = null): string {
	const asksVolunteers =
		def?.distribution === 'volunteers' || def?.distribution === 'captainOrVolunteer';
	switch (step) {
		case 'adjustTokens':
			return 'The crew may adjust the tokens before play.';
		case 'poll':
			return asksVolunteers
				? 'Who is willing to take the tasks? Everyone answers yes or no.'
				: `The ${terms.commander} asks the crew.`;
		case 'pick':
			if (pick === 'volunteers') {
				return def && def.volunteers > 1
					? `The ${terms.commander} picks the ${numberWord(def.volunteers)} volunteers who share the tasks.`
					: `The ${terms.commander} picks the volunteer who takes all the tasks.`;
			}
			return pick ? PICK_PROMPT[pick] : '';
		case 'draft':
			return 'The crew drafts the tasks.';
		case 'assign':
			return def?.distribution === 'free'
				? `The ${terms.commander} assigns every task to a crew member.`
				: `The ${terms.commander} splits the tasks between the volunteers, each taking at least one.`;
		case 'predict':
			return 'The owners of the prediction tasks note how many tricks they will win.';
		case 'pass':
			return 'Distress signal: everyone passes a card.';
		case 'play':
		case '':
			return '';
	}
}

/** What resultText needs of a task; a Deep Sea task has an objective instead of a card. */
type ResultTask = Pick<Task, 'card'> & Partial<Pick<Task, 'objective' | 'owner' | 'predicted'>>;

/** Why an attempt ended, in one sentence. `r.task` indexes `tasks`. */
export function resultText(
	r: AttemptResult,
	members: Member[],
	you: number,
	tasks: ResultTask[],
	objectives: Objective[] = []
): string {
	switch (r.reason) {
		case 'completed':
			return 'Every task and objective is fulfilled.';
		case 'gaveUp':
			return 'The crew gave up this attempt.';
		case 'taskFailed': {
			const t = tasks[r.task];
			if (t?.objective) {
				// Deep Sea: no card to be won by someone else; the task simply could not be met any more.
				const what = `${ownerText(t.owner ?? OWNER_UNASSIGNED, members, you)}: ${taskPhrase(t.objective, members.length, t.predicted)}`;
				return r.trick > 0 ? `Failed in trick ${r.trick}: ${what}.` : `Failed: ${what}.`;
			}
			const card = t?.card ? cardLabel(t.card) : 'a task';
			if (r.seat < 0) return `Task ${card} failed: nobody won it.`;
			const who = seatName(members, r.seat, you);
			return `Task ${card} failed: ${who === 'You' ? 'you' : who} won it in trick ${r.trick}.`;
		}
		case 'tokenOrder': {
			const t = tasks[r.task];
			const card = t?.card ? cardLabel(t.card) : 'a task';
			return `Token order broken in trick ${r.trick}: task ${card} can no longer come in its turn.`;
		}
		case 'objectiveFailed': {
			const o = objectives[r.objective];
			const what = o
				? `${ownerText(o.owner, members, you)}: ${objectiveText(o)}`
				: 'a special objective';
			return r.trick > 0 ? `Failed in trick ${r.trick}: ${what}.` : `Failed: ${what}.`;
		}
	}
}

/** The statement of a communicated card. */
export function positionText(p: Position): string {
	switch (p) {
		case 'highest':
			return 'my highest of this colour';
		case 'only':
			return 'my only card of this colour';
		case 'lowest':
			return 'my lowest of this colour';
		case '':
			return 'position not told (dead zone)';
	}
}
