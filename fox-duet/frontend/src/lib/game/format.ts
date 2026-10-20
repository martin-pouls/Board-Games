// Display helpers: names, the viewer-relative board orientation, labels and status texts.
// Each player sees their own end of the path on the left.
import type {
	Ability,
	Card,
	CardRef,
	End,
	MoveOption,
	PlayerSummary,
	PlayerView,
	Result,
	Suit
} from '#lib/types.js';

export const SUITS: Suit[] = ['doves', 'roses', 'stars'];

export const SUIT_NAME: Record<Suit, string> = { doves: 'Doves', roses: 'Roses', stars: 'Stars' };

/** Singular, for "the 7 of Roses" style labels. */
const SUIT_SINGULAR: Record<Suit, string> = { doves: 'Dove', roses: 'Rose', stars: 'Star' };

export const ABILITY: Record<Exclude<Ability, ''>, { name: string; text: string }> = {
	musician: { name: 'Musician', text: 'The trick winner may move the traveller the other way.' },
	foxes: {
		name: 'Foxes',
		text: 'Choose a player who may swap the decree card for a card from their hand.'
	},
	gazelle: { name: 'Gazelle', text: 'The trick winner may ignore the movement of one card.' },
	gift: { name: 'Gift', text: 'Both players give each other one hand card at the same time.' },
	royalHeir: { name: 'Royal Heir', text: 'Led: the other player need not follow suit.' }
};

/** Movement per rank (index = rank), as on the reference card. */
export const MOVEMENT = [0, 0, 3, 2, 1, 1, 2, 0, 2, 0, 3];

export function plural(n: number, word: string, many = `${word}s`): string {
	return `${n} ${n === 1 ? word : many}`;
}

/** A key that identifies a card everywhere it appears (hand, trick, decree). */
export function cardKey(c: CardRef): string {
	return `${c.suit}-${c.rank}`;
}

export function sameCard(a: CardRef | null | undefined, b: CardRef | null | undefined): boolean {
	return !!a && !!b && a.suit === b.suit && a.rank === b.rank;
}

export function cardLabel(c: CardRef): string {
	return `${c.rank} ${SUIT_SINGULAR[c.suit]}`;
}

export function abilityName(c: Card): string {
	return c.ability ? ABILITY[c.ability].name : '';
}

/** The viewer's position on the board: their own end is shown on the left. */
export function displayIndex(i: number, n: number, myEnd: End): number {
	return myEnd === 'left' ? i : n - 1 - i;
}

/** Spaces moved toward the viewer's end (negative: toward the teammate's end). */
export function towardMe(delta: number, myEnd: End): number {
	return myEnd === 'left' ? -delta : delta;
}

/** The viewer and the teammate (null before the teammate joins). */
export function seats(view: Pick<PlayerView, 'players'>): {
	me: PlayerSummary;
	mate: PlayerSummary | null;
} {
	const me = view.players.find((p) => p.isYou) ?? view.players[0];
	return { me, mate: view.players.find((p) => !p.isYou) ?? null };
}

/** "you" for the viewer, else the player's name. */
export function who(view: Pick<PlayerView, 'players'>, id: string | undefined): string {
	const p = view.players.find((x) => x.id === id);
	if (!p) return 'someone';
	return p.isYou ? 'you' : p.name;
}

function capital(s: string): string {
	return s.charAt(0).toUpperCase() + s.slice(1);
}

/** Whose end it is, from the viewer's side: "your end" or "Bob's end". */
export function endLabel(view: Pick<PlayerView, 'players'>, end: End): string {
	const p = view.players.find((x) => x.end === end);
	if (!p) return `the ${end} end`;
	return p.isYou ? 'your end' : `${p.name}'s end`;
}

/** A move option from the winner's view, e.g. "3 toward you", plus what it leads to. */
export function moveLabel(
	view: Pick<PlayerView, 'players'>,
	o: MoveOption
): { title: string; result: string } {
	const { me, mate } = seats(view);
	const steps = towardMe(o.delta, me.end);
	const title =
		steps === 0
			? 'Stay'
			: `${plural(Math.abs(steps), 'space')} toward ${steps > 0 ? 'you' : (mate?.name ?? 'your teammate')}`;
	let result = 'no gem';
	if (o.lost) result = 'off the path: lost in the forest!';
	else if (o.beyond) result = 'off the path: forest token, back to start';
	else if (o.gem) result = 'collect a gem';
	return { title, result };
}

export function resultHeadline(result: Result): string {
	if (result.outcome === 'won') return `Victory! ${plural(result.score?.total ?? 0, 'point')}`;
	return `Defeat: ${lossReason(result)}`;
}

function lossReason(result: Result): string {
	switch (result.reason) {
		case 'lostInForest':
			return 'lost in the forest';
		case 'outOfTime':
			return 'out of time';
		default:
			return 'gave up';
	}
}

/** One sentence on why the game was lost. */
export function lossText(view: Pick<PlayerView, 'players' | 'result'>): string {
	const r = view.result;
	if (!r) return '';
	switch (r.reason) {
		case 'lostInForest':
			return 'The traveller left the path with no forest token left.';
		case 'outOfTime':
			return `Round 3 ended with ${plural(r.gemsLeft, 'gem')} still in the forest.`;
		case 'gaveUp':
			return `${capital(who(view, r.playerId))} gave up.`;
		default:
			return '';
	}
}

export interface TurnStatus {
	mine: boolean;
	/** "Your turn" or "Bob's turn". */
	title: string;
	/** On your turn: what to do, e.g. "Follow Stars if you can". */
	hint: string;
	/** On the other player's turn: "Waiting for Bob to lead." */
	waiting: string;
}

/** What the game waits for, from the viewer's side (the turn bar and the line below it). */
export function turnStatus(view: PlayerView): TurnStatus | null {
	const p = view.pending;
	if (!p || view.phase !== 'playing') return null;
	const mine = p.playerIds.includes(view.you);
	const name = capital(who(view, p.playerIds[0]));
	const mate = seats(view).mate?.name ?? 'your teammate';
	const status = (hint: string, waiting: string): TurnStatus => ({
		mine,
		title: mine ? 'Your turn' : `${name}'s turn`,
		hint: mine ? hint : '',
		waiting: mine ? '' : waiting
	});
	switch (p.step) {
		case 'lead':
			return status('Lead any card', `Waiting for ${name} to lead.`);
		case 'follow': {
			const lead = view.trick?.lead;
			const hint =
				lead?.ability === 'royalHeir'
					? 'Royal Heir led: play any card'
					: lead
						? `Follow ${SUIT_NAME[lead.suit]} if you can`
						: 'Play a card';
			return status(hint, `Waiting for ${name} to follow.`);
		}
		case 'foxesChoose':
			return status(
				'Foxes: who may swap the decree?',
				`Waiting for ${name} to choose for the Foxes.`
			);
		case 'foxesSwap':
			return status(
				'Swap a card for the decree, or keep it',
				`Waiting for ${name} to swap or keep the decree.`
			);
		case 'gift':
			return status(`Gift: pick a card for ${mate}`, `Waiting for ${name}'s gift.`);
		case 'move': {
			const abilities = [view.trick?.lead?.ability, view.trick?.follow?.ability];
			const musician = abilities.includes('musician');
			const gazelle = abilities.includes('gazelle');
			const hint =
				musician && gazelle
					? 'Musician and Gazelle: pick a move'
					: musician
						? 'Musician: pick the way'
						: gazelle
							? 'Gazelle: pick how far'
							: 'Move the traveller';
			return status(hint, `Waiting for ${name} to move the traveller.`);
		}
		case 'forestEnd':
			return status('Round over: cover an end', `Waiting for ${name} to place a forest token.`);
	}
}
