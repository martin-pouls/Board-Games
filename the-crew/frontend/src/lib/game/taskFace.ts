// The face of a Deep Sea task card, laid out like the printed card: a small head ("I win",
// "I win exactly"), then rows of small cards with "3×" badges, orange numbers, crew icons and short
// words. Pure functions; taskText.ts holds the full sentence (title, screen readers).
import { terms } from '#lib/edition.js';
import { LAST_TRICK, type CardMatch, type Objective, type Suit } from '#lib/types.js';

/** A small card on a task face: a colour (with or without a value), "any" colour with a value, or
 * a submarine. `count` puts a "3×" badge on it. */
export interface Mini {
	suit: Suit | 'any';
	/** 0 = the whole colour (no value). */
	rank: number;
	count?: number;
}
export type Icon = 'me' | 'others' | 'crew' | 'captain';
export type Seg =
	| { t: 'text'; text: string; big?: boolean }
	| { t: 'cards'; cards: Mini[]; join?: string }
	| { t: 'num'; value: number }
	| { t: 'icon'; icon: Icon };
export interface Face {
	head: string;
	rows: Seg[][];
	/** Small print at the bottom ("no submarines in the trick"). */
	note?: string;
}

const COLOURS: Suit[] = ['pink', 'yellow', 'blue', 'green'];

const text = (s: string, big = false): Seg => ({ t: 'text', text: s, big });
const cards = (c: Mini[], join?: string): Seg => ({ t: 'cards', cards: c, join });
const whole = (s: Suit): Mini => ({ suit: s, rank: 0 });

function matchMini(m: CardMatch): Mini {
	return { suit: m.suit || 'any', rank: m.rank };
}

/** The cards a winCount counts. */
function countMinis(o: Objective): Mini[] {
	if (o.ranks.length > 0) {
		const suits: (Suit | 'any')[] = o.suits.length ? o.suits : ['any'];
		return suits.flatMap((s) => o.ranks.map((r) => ({ suit: s, rank: r })));
	}
	return o.suits.map(whole);
}

function deckSize(o: Objective): number {
	if (o.ranks.length > 0) return (o.suits.length || 4) * o.ranks.length;
	return o.suits.reduce((n, s) => n + (s === 'trump' ? 4 : 9), 0);
}

function winCountFace(o: Objective): Face {
	const minis = countMinis(o);
	if (o.cmp === 'exactly' && o.count === 0) return { head: 'I win no', rows: [[cards(minis)]] };
	if (o.cmp === 'exactly' && o.count === deckSize(o) && o.ranks.length === 1 && !o.suits.length) {
		return {
			head: 'I win all',
			rows: [[cards(COLOURS.map((s) => ({ suit: s, rank: o.ranks[0] })))]]
		};
	}
	const head = o.cmp === 'atLeast' ? 'I win at least' : 'I win exactly';
	return { head, rows: [[cards(minis.map((m, i) => (i === 0 ? { ...m, count: o.count } : m)))]] };
}

/** "the first 3 tricks", "the last trick", "the first and the last trick". */
function tricksWords(tricks: number[]): string {
	const last = tricks.includes(LAST_TRICK);
	const n = tricks.filter((t) => t !== LAST_TRICK);
	const firstK = n.length > 0 && n.every((t, i) => t === i + 1);
	if (firstK && last && n.length === 1) return 'the first and the last trick';
	if (firstK && !last) return n.length === 1 ? 'the first trick' : `the first ${n.length} tricks`;
	if (!n.length && last) return 'the last trick';
	return `tricks ${n.join(', ')}${last ? ' and the last' : ''}`;
}

function trickWhereFace(o: Objective): Face {
	const note = `no ${terms.trumps}`;
	const head = 'I win a trick';
	switch (o.where) {
		case 'below':
			return { head, rows: [[text('all cards lower than')], [{ t: 'num', value: o.count }]], note };
		case 'above':
			return {
				head,
				rows: [[text('all cards greater than')], [{ t: 'num', value: o.count }]],
				note
			};
		case 'even':
		case 'odd': {
			const ranks = o.where === 'even' ? [2, 4, 6, 8] : [1, 3, 5, 7, 9];
			return {
				head,
				rows: [
					[text(`only ${o.where} values`)],
					[cards(ranks.map((r) => ({ suit: 'any', rank: r })))]
				],
				note
			};
		}
		case 'sumAtLeast':
			return { head, rows: [[text('total value at least')], [{ t: 'num', value: o.count }]], note };
		case 'sumBelow':
			return {
				head,
				rows: [[text('total value lower than')], [{ t: 'num', value: o.count }]],
				note
			};
		case 'sumBetween': {
			const nums: Seg[] = [];
			for (let v = o.count; v <= o.max; v++) {
				if (nums.length) nums.push(text('or'));
				nums.push({ t: 'num', value: v });
			}
			return { head, rows: [[text('total value')], nums], note };
		}
		case '':
			return { head, rows: [] };
	}
}

function allFace(o: Objective, players: number, predicted: boolean): Face {
	const [a, b] = o.parts;
	if (o.parts.length === 2 && a.kind === 'winCards' && b.kind === 'cardInLastTrick') {
		return { head: 'I win', rows: [[cards(a.cards)], [text('in the last trick')]] };
	}
	if (
		o.parts.length === 2 &&
		a.kind === 'winCount' &&
		a.suits.length === 1 &&
		a.suits[0] === 'trump' &&
		b.kind === 'winCards' &&
		b.cards.length === 1
	) {
		return { head: 'I win', rows: [[cards(b.cards)], [text(`no other ${terms.trump}`)]] };
	}
	if (o.parts.every((p) => p.kind === 'winCount' && p.cmp === a.cmp && p.count > 0)) {
		const head = a.cmp === 'atLeast' ? 'I win at least' : 'I win exactly';
		return {
			head,
			rows: [[cards(o.parts.flatMap((p) => countMinis(p).map((m) => ({ ...m, count: p.count }))))]]
		};
	}
	const faces = o.parts.map((p) => taskFace(p, players, predicted));
	return { head: faces[0]?.head ?? '', rows: faces.flatMap((f) => f.rows) };
}

/**
 * The face of a task card. `players` is the number of crew members (the dummy included);
 * `predicted` says whether a prediction has been noted (its count is -1 when hidden from the viewer).
 */
export function taskFace(o: Objective, players: number, predicted = o.count >= 0): Face {
	switch (o.kind) {
		case 'cmpTricks': {
			const words =
				o.cmp === 'fewer'
					? 'fewer tricks than each other member'
					: o.cmp === 'moreThanAll'
						? 'more tricks than all others together'
						: 'more tricks than each other member';
			const sym = o.cmp === 'fewer' ? '<' : '>';
			return {
				head: 'I win',
				rows: [
					[text(words)],
					[
						{ t: 'icon', icon: 'me' },
						text(sym, true),
						{ t: 'icon', icon: o.cmp === 'moreThanAll' ? 'crew' : 'others' }
					]
				]
			};
		}
		case 'cmpCaptain': {
			const words =
				o.cmp === 'equal'
					? 'as many tricks as'
					: o.cmp === 'fewer'
						? 'fewer tricks than'
						: 'more tricks than';
			return {
				head: 'I win',
				rows: [[text(words)], [{ t: 'icon', icon: 'captain' }, text(`the ${terms.commander}`)]],
				note: `not for the ${terms.commander}`
			};
		}
		case 'predict': {
			const n = !predicted ? 'X' : o.count >= 0 ? String(o.count) : '?';
			return {
				head: 'I win',
				rows: [[text(n, true), text(n === '1' ? 'trick' : 'tricks')]],
				note: !predicted
					? o.hidden
						? 'predict it, keep it secret'
						: 'predict it, tell the crew'
					: o.hidden
						? 'secret prediction'
						: undefined
			};
		}
		case 'winNoTricks':
			return { head: 'I win', rows: [[text('0', true), text('tricks')]] };
		case 'winExactlyTricks':
			return {
				head: 'I win exactly',
				rows: [
					[text(String(o.count), true), text(o.count === 1 ? 'trick' : 'tricks')],
					...(o.without.length
						? [
								[
									text('none with'),
									cards(o.without.map((r) => ({ suit: 'any' as const, rank: r })))
								]
							]
						: [])
				]
			};
		case 'avoidTricks':
			return {
				head: 'I win',
				rows: [
					[
						text(
							o.tricks.length === 1
								? `not ${tricksWords(o.tricks)}`
								: `none of ${tricksWords(o.tricks)}`
						)
					]
				]
			};
		case 'noConsecutive':
			return { head: 'I win', rows: [[text('never'), text('2', true), text('tricks in a row')]] };
		case 'winTricks':
			return { head: 'I win', rows: [[text(tricksWords(o.tricks))]] };
		case 'winOnlyTricks':
			return { head: 'I win', rows: [[text(`only ${tricksWords(o.tricks)}`)]] };
		case 'consecutive':
			return o.exact
				? {
						head: 'I win exactly',
						rows: [[text(String(o.count), true), text('tricks')], [text('and all in a row')]]
					}
				: { head: 'I win', rows: [[text(String(o.count), true), text('tricks in a row')]] };
		case 'winWith': {
			const w = o.with;
			const t = o.target;
			if (!w) return { head: 'I win a trick', rows: [] };
			if (!t) return { head: 'I win a trick', rows: [[text('with'), cards([matchMini(w)])]] };
			const same = !t.suit && !w.suit && t.rank === w.rank;
			return {
				head: 'I win',
				rows: [[cards([matchMini(t)]), text('with'), cards([matchMini(w)])]],
				note: same ? `two different ${t.rank}s` : undefined
			};
		}
		case 'winTrickWhere':
			return trickWhereFace(o);
		case 'winCards':
			return { head: 'I win', rows: [[cards(o.cards)]] };
		case 'cardInLastTrick':
			return { head: 'I win', rows: [[cards(o.cards)], [text('in the last trick')]] };
		case 'winCount':
			return winCountFace(o);
		case 'winEachColour':
			return {
				head: 'I win at least',
				rows: [[text('one card of each colour')], [cards(COLOURS.map(whole))]]
			};
		case 'winAllOfAnyColour':
			return { head: 'I win', rows: [[text('all cards of at least one colour')]] };
		case 'neverLead':
			return o.suits.length > 2
				? { head: 'I never lead with', rows: [[cards(o.suits.map(whole))]], note: 'any of these' }
				: { head: 'I never lead with', rows: [[cards(o.suits.map(whole), 'or')]] };
		case 'equalSuits':
			return {
				head: 'I win as many',
				rows: [
					[cards([whole(o.suits[0])]), text('as'), cards([whole(o.suits[1])])],
					...(o.inOneTrick ? [[text('in one trick')]] : [])
				],
				note: '0 and 0 does not count'
			};
		case 'moreSuit':
			return {
				head: 'I win more',
				rows: [[cards([whole(o.suits[0])]), text('than'), cards([whole(o.suits[1])])]],
				note: o.suits[1] ? `0 ${o.suits[1]} is fine` : undefined
			};
		case 'all':
			return allFace(o, players, predicted);
		default:
			return { head: '', rows: [] };
	}
}
