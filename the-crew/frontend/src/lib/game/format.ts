// Texts for objectives, mission briefings and results. Pure functions, unit-tested.
import { cardLabel, SUIT_NAMES, tokenLabel } from './cards.js';
import {
	LAST_TRICK,
	OWNER_CREW,
	OWNER_UNASSIGNED,
	type AttemptResult,
	type Card,
	type GameEvent,
	type MissionDef,
	type Member,
	type Objective,
	type ObjectiveSpec,
	type OwnerRule,
	type PickKind,
	type Position,
	type Step
} from '#lib/types.js';

export const JARVIS = 'JARVIS';

/** "You", the player's name, or JARVIS. */
export function seatName(members: Member[], seat: number, you: number): string {
	if (seat === you) return 'You';
	const m = members.find((m) => m.seat === seat);
	if (!m) return `Seat ${seat + 1}`;
	return m.dummy ? JARVIS : m.name;
}

/** "trick 3", "the last trick". */
export function trickText(n: number): string {
	return n === LAST_TRICK ? 'the last trick' : `trick ${n}`;
}

/** "a", "a and b", "a, b and c". */
export function listText(items: string[], and = 'and'): string {
	if (items.length <= 1) return items.join('');
	return `${items.slice(0, -1).join(', ')} ${and} ${items[items.length - 1]}`;
}

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

function plural(n: number, word: string): string {
	return `${n === 1 ? 'one' : n} ${word}${n === 1 ? '' : 's'}`;
}

/**
 * What an objective asks for, without its owner: "win pink 7", "win no trick". Phrased for the
 * owner as subject ("You: win …", "The crew: …").
 */
export function objectiveText(o: Objective): string {
	const noTrump = o.notWithTrump ? ', none of them with a rocket' : '';
	switch (o.kind) {
		case 'winCards':
			return `win ${cardsText(o.cards)}`;
		case 'winNoTricks':
			return 'win no trick';
		case 'winExactlyTricks':
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
			return `nobody may ever have ${plural(o.count + 1, 'trick')} more than another member`;
		case 'cardInLastTrick':
			return o.cards.length
				? `win ${cardsText(o.cards)} in the last trick`
				: 'win the Ω task in the last trick';
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
	commander: 'The commander',
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

const DISTRIBUTION: Record<MissionDef['distribution'], string> = {
	'': '',
	open: 'Tasks are drafted, the commander first.',
	commanderDecision:
		'The commander asks who takes the tasks, then decides; one member takes them all.',
	commanderDistribution:
		'The commander reveals the tasks one by one and assigns each after asking the crew; the tasks must be shared out as evenly as possible.'
};

/** The briefing of a mission: its tasks, tokens and special rules, one line each. */
export function briefing(m: MissionDef): string[] {
	const lines: string[] = [];
	if (m.tasks > 0) {
		lines.push(`${plural(m.tasks, 'task')}. ${DISTRIBUTION[m.distribution]}`.trim());
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
		lines.push('The commander picks a crew member who may not communicate.');
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

const PICK_PROMPT: Record<Exclude<PickKind, ''>, string> = {
	silence: 'The commander picks a crew member who may not communicate.',
	roles: 'The commander picks the two crew members for the special objectives.',
	objective: 'The commander picks who takes the special objective.',
	decision: 'The commander picks who takes all the tasks.',
	distribute: 'The commander assigns the revealed task.'
};

/** A sentence about the current step, for the status line. */
export function stepText(step: Step, pick: PickKind): string {
	switch (step) {
		case 'adjustTokens':
			return 'The crew may adjust the tokens before play.';
		case 'poll':
			return 'The commander asks the crew.';
		case 'pick':
			return pick ? PICK_PROMPT[pick] : '';
		case 'draft':
			return 'The crew drafts the tasks.';
		case 'pass':
			return 'Distress signal: everyone passes a card.';
		case 'play':
		case '':
			return '';
	}
}

/** Why an attempt ended, in one sentence. `r.task` indexes `tasks`. */
export function resultText(
	r: AttemptResult,
	members: Member[],
	you: number,
	tasks: { card: Card | null }[]
): string {
	switch (r.reason) {
		case 'completed':
			return 'Mission accomplished. Every task is fulfilled.';
		case 'gaveUp':
			return 'The crew gave up this attempt.';
		case 'taskFailed': {
			const t = tasks[r.task];
			const card = t?.card ? cardLabel(t.card) : 'a task';
			const who = r.seat >= 0 ? seatName(members, r.seat, you) : 'the wrong member';
			return r.trick > 0
				? `Task ${card} failed: ${who === 'You' ? 'you' : who} won it in trick ${r.trick}.`
				: `Task ${card} failed.`;
		}
		case 'tokenOrder': {
			const t = tasks[r.task];
			const card = t?.card ? cardLabel(t.card) : 'a task';
			return r.trick > 0
				? `Task ${card} was fulfilled out of token order in trick ${r.trick}.`
				: `Task ${card} was fulfilled out of token order.`;
		}
		case 'objectiveFailed':
			return r.trick > 0
				? `A special objective failed in trick ${r.trick}.`
				: 'A special objective failed.';
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

function subject(members: Member[], seat: number, you: number): string {
	return seatName(members, seat, you);
}

/** One line about the latest events, for the chronicle under the trick ("" when nothing fits). */
export function describeEvents(events: GameEvent[], members: Member[], you: number): string {
	const parts: string[] = [];
	for (const e of events) {
		switch (e.type) {
			case 'cardPlayed':
				parts.push(`${subject(members, e.seat, you)} played ${cardLabel(e.card)}.`);
				break;
			case 'trickWon':
				parts.push(
					e.seat === you
						? `You won trick ${e.number}.`
						: `${subject(members, e.seat, you)} won trick ${e.number}.`
				);
				break;
			case 'communicated':
				parts.push(
					`${subject(members, e.seat, you)} communicated ${cardLabel(e.card)}${
						e.position ? ` (${e.position})` : ''
					}.`
				);
				break;
			case 'taskPicked':
				parts.push(
					e.seat === you ? 'You took a task.' : `${subject(members, e.seat, you)} took a task.`
				);
				break;
			case 'distressSignal':
				parts.push(
					e.skipped ? 'No distress signal.' : `Distress signal: pass a card to the ${e.direction}.`
				);
				break;
			case 'cardsPassed':
				parts.push('The cards were passed.');
				break;
			case 'cardsDrawn':
				parts.push('Everyone drew a card from their right neighbour.');
				break;
			case 'taskHandedOver':
				parts.push(
					`A task went from ${subject(members, e.from, you)} to ${subject(members, e.to, you)}.`
				);
				break;
			case 'tokensAdjusted':
				parts.push(e.from < 0 ? 'The tokens stay as they are.' : 'The tokens were adjusted.');
				break;
		}
	}
	return parts.slice(-2).join(' ');
}
