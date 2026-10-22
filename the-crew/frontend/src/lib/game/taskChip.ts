// The short form of a Deep Sea task card, for the chips in the crew's rows and the draft: a few
// words with the small cards drawn as printed ("No [pink]", "[7] takes [5]", "[pink] > [green]").
// Pure functions, tested over all 96 cards; taskText.ts holds the full sentence (the chip's title,
// the sheet, screen readers) and taskFace.ts the printed card.
import { cap, terms } from '#lib/edition.js';
import { LAST_TRICK, type CardMatch, type Objective, type Suit, type Task } from '#lib/types.js';
import { SUIT_NAMES } from './cards.js';
import type { Icon, Mini } from './taskFace.js';

export type ChipSeg =
	{ t: 'text'; text: string } | { t: 'card'; card: Mini } | { t: 'icon'; icon: Icon };

const W = (text: string): ChipSeg => ({ t: 'text', text });
const C = (card: Mini): ChipSeg => ({ t: 'card', card });
const whole = (suit: Suit): Mini => ({ suit, rank: 0 });
const match = (m: CardMatch): Mini => ({ suit: m.suit || 'any', rank: m.rank });

/** A Deep Sea task card (face up or down), as opposed to a Planet Nine task (a playing card). */
export function isTaskCard(t: Task): boolean {
	return !t.card && (!!t.objective || t.difficulty > 0);
}

/** The first `k` tricks, if the list is exactly those. */
function firstK(tricks: number[]): number {
	return tricks.length > 0 && tricks.every((t, i) => t === i + 1) ? tricks.length : 0;
}

/** "Trick 1", "Tricks 1–3", "Last trick", "First + last". */
function tricksChip(tricks: number[]): string {
	const last = tricks.includes(LAST_TRICK);
	const n = tricks.filter((t) => t !== LAST_TRICK);
	if (last && n.length === 1 && n[0] === 1) return 'First + last';
	if (last && !n.length) return 'Last trick';
	const k = firstK(n);
	const nums = k > 1 ? `1–${k}` : n.join(', ');
	return `${n.length === 1 ? 'Trick' : 'Tricks'} ${nums}${last ? ' + last' : ''}`;
}

/** The small cards a winCount counts, the first with the count badge. */
function countCards(o: Objective, count?: number): Mini[] {
	const minis: Mini[] =
		o.ranks.length > 0
			? (o.suits.length ? o.suits : (['any'] as const)).flatMap((s) =>
					o.ranks.map((r) => ({ suit: s, rank: r }))
				)
			: o.suits.map(whole);
	return count === undefined ? minis : minis.map((m, i) => (i === 0 ? { ...m, count } : m));
}

function deckSize(o: Objective): number {
	if (o.ranks.length > 0) return (o.suits.length || 4) * o.ranks.length;
	return o.suits.reduce((n, s) => n + (s === 'trump' ? 4 : 9), 0);
}

function winCountChip(o: Objective): ChipSeg[] {
	if (o.cmp === 'exactly' && o.count === 0) return [W('No'), ...countCards(o).map(C)];
	if (o.cmp === 'exactly' && o.count === deckSize(o)) return countCards(o, o.count).map(C);
	return [W(o.cmp === 'atLeast' ? 'At least' : 'Exactly'), ...countCards(o, o.count).map(C)];
}

function trickWhereChip(o: Objective): ChipSeg[] {
	switch (o.where) {
		case 'below':
			return [W(`All under ${o.count}`)];
		case 'above':
			return [W(`All over ${o.count}`)];
		case 'even':
			return [W('All even')];
		case 'odd':
			return [W('All odd')];
		case 'sumAtLeast':
			return [W(`Total ${o.count} or more`)];
		case 'sumBelow':
			return [W(`Total under ${o.count}`)];
		case 'sumBetween':
			return [
				W(o.max === o.count + 1 ? `Total ${o.count} or ${o.max}` : `Total ${o.count}–${o.max}`)
			];
		case '':
			return [W('A trick')];
	}
}

function allChip(o: Objective, predicted: boolean): ChipSeg[] {
	const [a, b] = o.parts;
	if (o.parts.length === 2 && a.kind === 'winCards' && b.kind === 'cardInLastTrick') {
		return [...a.cards.map(C), W('in last')];
	}
	if (
		o.parts.length === 2 &&
		a.kind === 'winCount' &&
		a.suits.length === 1 &&
		a.suits[0] === 'trump' &&
		b.kind === 'winCards' &&
		b.cards.length === 1
	) {
		return [W('Only'), C(b.cards[0])];
	}
	if (o.parts.every((p) => p.kind === 'winCount' && p.cmp === a.cmp && p.count > 0)) {
		return [
			W(a.cmp === 'atLeast' ? 'At least' : 'Exactly'),
			...o.parts.flatMap((p) => countCards(p, p.count).map(C))
		];
	}
	return o.parts.flatMap((p, i) => [...(i ? [W('+')] : []), ...taskChip(p, predicted)]);
}

/**
 * The chip of an objective. `predicted` says whether a prediction has been noted (its count is -1
 * when it is hidden from the viewer).
 */
export function taskChip(o: Objective, predicted = o.count >= 0): ChipSeg[] {
	switch (o.kind) {
		case 'cmpTricks':
			return [
				W(
					o.cmp === 'fewer'
						? 'Fewest tricks'
						: o.cmp === 'moreThanAll'
							? 'More than all others'
							: 'Most tricks'
				)
			];
		case 'cmpCaptain':
			return [
				W(o.cmp === 'equal' ? 'As many as' : o.cmp === 'fewer' ? 'Fewer than' : 'More than'),
				{ t: 'icon', icon: 'captain' }
			];
		case 'predict': {
			if (!predicted) return [W(o.hidden ? 'Predict, secret' : 'Predict tricks')];
			const n = o.count >= 0 ? `${o.count} ${o.count === 1 ? 'trick' : 'tricks'}` : '? tricks';
			return [W(`${n}, ${o.hidden ? 'secret' : 'said'}`)];
		}
		case 'winNoTricks':
			return [W('0 tricks')];
		case 'winExactlyTricks':
			return [W(`Exactly ${o.count} ${o.count === 1 ? 'trick' : 'tricks'}`)];
		case 'avoidTricks': {
			const k = firstK(o.tricks);
			if (k > 1) return [W(`None of 1–${k}`)];
			return [W(`Not ${tricksChip(o.tricks).toLowerCase()}`)];
		}
		case 'noConsecutive':
			return [W('Never 2 in a row')];
		case 'winTricks':
			return [W(tricksChip(o.tricks))];
		case 'winOnlyTricks':
			return [
				W(
					o.tricks.length === 1 && o.tricks[0] === LAST_TRICK
						? 'Only the last'
						: o.tricks.length === 1 && o.tricks[0] === 1
							? 'Only the first'
							: `Only ${tricksChip(o.tricks).toLowerCase()}`
				)
			];
		case 'consecutive':
			return [W(o.exact ? `Exactly ${o.count}, in a row` : `${o.count} in a row`)];
		case 'winWith': {
			if (!o.with) return [W('A trick')];
			if (!o.target) return [W('Win with'), C(match(o.with))];
			return [C(match(o.with)), W('takes'), C(match(o.target))];
		}
		case 'winTrickWhere':
			return trickWhereChip(o);
		case 'winCards':
			return o.cards.map(C);
		case 'cardInLastTrick':
			return [...o.cards.map(C), W('in last')];
		case 'winCount':
			return winCountChip(o);
		case 'winEachColour':
			return [W('Each colour')];
		case 'winAllOfAnyColour':
			return [W('A whole colour')];
		case 'neverLead':
			return [W('Never lead'), ...o.suits.map((s) => C(whole(s)))];
		case 'equalSuits':
			return [
				C(whole(o.suits[0])),
				W('='),
				C(whole(o.suits[1])),
				...(o.inOneTrick ? [W('in a trick')] : [])
			];
		case 'moreSuit':
			return [C(whole(o.suits[0])), W('>'), C(whole(o.suits[1]))];
		case 'all':
			return allChip(o, predicted);
		default:
			return [W(cap(o.kind))];
	}
}

/** A small card in words: "pink 3", "pink", "6", "submarine 2", "3× 5". */
function miniText(m: Mini): string {
	const name = m.suit === 'any' ? '' : SUIT_NAMES[m.suit];
	const card = [name, m.rank ? String(m.rank) : ''].filter(Boolean).join(' ');
	return m.count ? `${m.count}× ${card}` : card;
}

/** The chip in words, for buttons: "No pink", "7 takes 5", "As many as the captain". */
export function chipText(segs: ChipSeg[]): string {
	return cap(
		segs
			.map((s) =>
				s.t === 'text' ? s.text : s.t === 'card' ? miniText(s.card) : `the ${terms.commander}`
			)
			.join(' ')
	);
}
