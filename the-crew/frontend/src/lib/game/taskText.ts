// What a Deep Sea task card asks for, in words: one case per objective kind. Pure functions,
// unit-tested over all 96 cards. Planet Nine's own kinds are phrased in format.ts.
import { cap, terms } from '#lib/edition.js';
import { cardLabel, SUIT_NAMES } from './cards.js';
import {
	LAST_TRICK,
	type Card,
	type CardMatch,
	type Objective,
	type Suit,
	type Task
} from '#lib/types.js';
import { listText, numberWord, plural, withArticle } from './words.js';

/** Tricks per attempt by crew members (the dummy included). */
const TOTAL_TRICKS: Record<number, number> = { 3: 13, 4: 10, 5: 8 };

const trickWord = (n: number) => (n === 1 ? 'trick' : 'tricks');

/** "the pink 3"; a submarine has no article: "submarine 3". */
function cardNoun(c: Card): string {
	return c.suit === 'trump' ? cardLabel(c) : `the ${cardLabel(c)}`;
}

/** "the pink 9 and the yellow 8", "the blue 1, 2 and 3". */
function cardsPhrase(cards: Card[]): string {
	const first = cards[0];
	if (cards.length > 1 && first.suit !== 'trump' && cards.every((c) => c.suit === first.suit)) {
		return `the ${SUIT_NAMES[first.suit]} ${listText(cards.map((c) => String(c.rank)))}`;
	}
	return listText(cards.map(cardNoun));
}

/** "a 6", "a submarine", "the pink 7", "submarine 3". */
function matchNoun(m: CardMatch): string {
	if (m.suit === 'trump')
		return m.rank ? cardLabel({ suit: 'trump', rank: m.rank }) : `a ${terms.trump}`;
	if (m.suit) return `the ${SUIT_NAMES[m.suit]} ${m.rank}`;
	return withArticle(m.rank);
}

/** The tricks a trick list names: "the first two tricks", "the last trick", "trick 3 and 5". */
function tricksPhrase(tricks: number[]): string {
	const last = tricks.includes(LAST_TRICK);
	const numbered = tricks.filter((t) => t !== LAST_TRICK);
	const firstK = numbered.length > 0 && numbered.every((t, i) => t === i + 1);
	if (firstK && last && numbered.length === 1) return 'the first and the last trick';
	const parts: string[] = [];
	if (firstK) {
		parts.push(
			numbered.length === 1 ? 'the first trick' : `the first ${numberWord(numbered.length)} tricks`
		);
	} else if (numbered.length > 0) {
		parts.push(`${numbered.length === 1 ? 'trick' : 'tricks'} ${listText(numbered.map(String))}`);
	}
	if (last) parts.push('the last trick');
	return listText(parts);
}

/** The cards a winCount counts, as a noun: "5s", "pink cards", "8s or 9s", "submarines". */
function countNoun(o: Objective, single: boolean): string {
	const suits = o.suits;
	if (o.ranks.length > 0) {
		const ranks = listText(
			o.ranks.map((r) => (single ? String(r) : `${r}s`)),
			'or'
		);
		const colours = suits.map((s) => SUIT_NAMES[s]);
		return colours.length ? `${listText(colours, 'or')} ${ranks}` : ranks;
	}
	const colours = suits.filter((s) => s !== 'trump').map((s) => SUIT_NAMES[s]);
	const parts: string[] = [];
	if (colours.length) parts.push(`${listText(colours, 'or')} ${single ? 'card' : 'cards'}`);
	if (suits.includes('trump')) parts.push(single ? terms.trump : terms.trumps);
	return listText(parts, 'or');
}

/** How many cards of a winCount's selection the deck has. */
function deckSize(o: Objective): number {
	if (o.ranks.length > 0) return (o.suits.length || 4) * o.ranks.length;
	return o.suits.reduce((n, s) => n + (s === 'trump' ? 4 : 9), 0);
}

function winCountPhrase(o: Objective): string {
	if (o.cmp === 'exactly' && o.count === 0) return `don't win any ${countNoun(o, false)}`;
	if (o.cmp === 'exactly' && o.count === deckSize(o)) {
		return `win all ${numberWord(o.count)} ${countNoun(o, false)}`;
	}
	const rel = o.cmp === 'atLeast' ? 'at least' : 'exactly';
	return `win ${rel} ${numberWord(o.count)} ${countNoun(o, o.count === 1)}`;
}

function winWithPhrase(o: Objective): string {
	const w = o.with;
	const t = o.target;
	if (!w) return 'win a trick';
	if (!t) return `win a trick with ${matchNoun(w)}`;
	if (t.rank && !t.suit && !w.suit && w.rank === t.rank) {
		return `win any ${t.rank} with another ${t.rank}`;
	}
	return `win ${matchNoun(t)} with ${matchNoun(w)}`;
}

function trickWherePhrase(o: Objective): string {
	switch (o.where) {
		case 'below':
			return `win a trick where all cards are lower than ${o.count}`;
		case 'above':
			return `win a trick where all cards are greater than ${o.count}`;
		case 'even':
			return 'win a trick that has only even numbers';
		case 'odd':
			return 'win a trick that has only odd numbers';
		case 'sumAtLeast':
			return `win a trick with a total value of at least ${o.count}`;
		case 'sumBelow':
			return `win a trick with a total value lower than ${o.count}`;
		case 'sumBetween':
			return o.max === o.count + 1
				? `win a trick with a total value of ${o.count} or ${o.max}`
				: `win a trick with a total value from ${o.count} to ${o.max}`;
		case '':
			return 'win a trick';
	}
}

function predictPhrase(o: Objective, players: number, predicted: boolean): string {
	if (!predicted) {
		const total = TOTAL_TRICKS[Math.min(Math.max(players, 3), 5)];
		const range = players > 0 ? ` (0 to ${total})` : '';
		return o.hidden
			? `predict how many tricks you will win${range}, kept hidden`
			: `predict how many tricks you will win${range}`;
	}
	const n = o.count >= 0 ? plural(o.count, 'trick') : '? tricks';
	return `win exactly ${n}${o.hidden ? ' (prediction hidden)' : ''}`;
}

/** Both of a task's parts as one sentence, where the card reads better that way. */
function allPhrase(o: Objective, players: number): string {
	const [a, b] = o.parts;
	if (o.parts.length === 2 && a.kind === 'winCards' && b.kind === 'cardInLastTrick') {
		return `win ${cardsPhrase(a.cards)} in the last trick`;
	}
	if (
		o.parts.length === 2 &&
		a.kind === 'winCount' &&
		a.suits.length === 1 &&
		a.suits[0] === 'trump' &&
		b.kind === 'winCards' &&
		b.cards.length === 1
	) {
		return `win ${cardLabel(b.cards[0])} and no other ${terms.trump}`;
	}
	const counts = o.parts.every(
		(p) =>
			p.kind === 'winCount' &&
			p.ranks.length === 0 &&
			p.suits.length === 1 &&
			p.cmp === a.cmp &&
			p.count === a.count &&
			p.count > 0
	);
	if (counts) {
		const rel = a.cmp === 'atLeast' ? 'at least' : 'exactly';
		const what = o.parts.map((p) => `${numberWord(p.count)} ${SUIT_NAMES[p.suits[0]]}`);
		return `win ${rel} ${listText(what)} ${a.count === 1 ? 'card' : 'cards'}`;
	}
	return listText(o.parts.map((p) => taskPhrase(p, players)));
}

/** What a neverLead forbids to lead: "a pink, yellow or blue card", "a pink card or a submarine". */
function leadPhrase(suits: Suit[]): string {
	const colours = suits.filter((s) => s !== 'trump').map((s) => SUIT_NAMES[s]);
	const parts: string[] = [];
	if (colours.length) parts.push(`a ${listText(colours, 'or')} card`);
	if (suits.includes('trump')) parts.push(`a ${terms.trump}`);
	return listText(parts, 'or');
}

/**
 * What an objective asks for, starting lower case ("win more tricks than everyone else", "the
 * yellow 5 must be the last card played in the last trick"). `players` is the number of crew
 * members, the dummy included; `predicted` says whether a prediction has been noted.
 */
export function taskPhrase(o: Objective, players: number, predicted = o.count >= 0): string {
	switch (o.kind) {
		case 'cmpTricks':
			return o.cmp === 'fewer'
				? 'win fewer tricks than everyone else'
				: o.cmp === 'moreThanAll'
					? 'win more tricks than everyone else together'
					: 'win more tricks than everyone else';
		case 'cmpCaptain':
			return o.cmp === 'equal'
				? `win the same number of tricks as the ${terms.commander}`
				: `win ${o.cmp} tricks than the ${terms.commander}`;
		case 'predict':
			return predictPhrase(o, players, predicted);
		case 'winNoTricks':
			return "don't win any tricks";
		case 'winExactlyTricks': {
			const base = `win exactly ${numberWord(o.count)} ${trickWord(o.count)}`;
			if (o.without.length === 0) return base;
			const article = o.without[0] === 8 ? 'an' : 'a';
			return `${base}, none containing ${article} ${listText(o.without.map(String), 'or')}`;
		}
		case 'avoidTricks':
			return o.tricks.length === 1
				? `don't win ${tricksPhrase(o.tricks)}`
				: `don't win any of ${tricksPhrase(o.tricks)}`;
		case 'noConsecutive':
			return "don't win two consecutive tricks";
		case 'winTricks':
			return `win ${tricksPhrase(o.tricks)}`;
		case 'winOnlyTricks':
			return `win only ${tricksPhrase(o.tricks)}`;
		case 'consecutive':
			return o.exact
				? `win exactly ${numberWord(o.count)} tricks, all in a row`
				: `win ${numberWord(o.count)} consecutive tricks`;
		case 'winWith':
			return winWithPhrase(o);
		case 'winTrickWhere':
			return trickWherePhrase(o);
		case 'winCards':
			return `win ${cardsPhrase(o.cards)}`;
		case 'cardInLastTrick':
			return `win ${cardsPhrase(o.cards)} in the last trick`;
		case 'winCount':
			return winCountPhrase(o);
		case 'winEachColour':
			return 'win at least one card of each colour';
		case 'winAllOfAnyColour':
			return 'win all cards of at least one colour';
		case 'neverLead':
			return `don't open a trick with ${leadPhrase(o.suits)}`;
		case 'equalSuits': {
			const [a, b] = o.suits.map((s) => SUIT_NAMES[s]);
			return `win the same number of ${a} and ${b} cards${o.inOneTrick ? ' in one trick' : ''} (more than 0)`;
		}
		case 'moreSuit': {
			const [a, b] = o.suits.map((s) => SUIT_NAMES[s]);
			return `win more ${a} cards than ${b} cards`;
		}
		case 'all':
			return allPhrase(o, players);
		case 'crewBalance':
			return o.rank
				? `nobody may ever have won ${numberWord(o.count + 1)} more ${o.rank}s than another member`
				: `nobody may ever have ${plural(o.count + 1, 'trick')} more than another member`;
		case 'leaderStaysAhead':
			return 'whoever wins the first trick must always have won more tricks than every other member';
		case 'lastCardPlayed':
			return `${o.cards.length ? cardNoun(o.cards[0]) : 'the card'} must be the last card played in the last trick`;
		case 'trickQuota':
			return 'every member must win exactly as many tricks as the difficulties of their tasks add up to';
		case 'none':
			return 'its difficulty adds to the tricks you must win';
		default:
			// Planet Nine's own kinds are phrased by objectiveText (format.ts).
			return o.kind;
	}
}

/** A task card's text: "Win more tricks than everyone else". */
export function taskText(o: Objective, players: number, predicted = o.count >= 0): string {
	return cap(taskPhrase(o, players, predicted));
}

/** The text of a task as the viewer sees it; a face-down task shows only its difficulty. */
export function taskTextOf(t: Task, players: number): string {
	if (!t.objective)
		return t.difficulty > 0 ? `Face-down task (difficulty ${t.difficulty})` : 'Face-down task';
	return taskText(t.objective, players, t.predicted);
}
