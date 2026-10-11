// Display helpers: names, the viewer-relative board orientation, labels and status texts.
// Each player sees their own end of the path on the left.
import type {
	Ability,
	Card,
	CardRef,
	End,
	GameEvent,
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

/** What the game waits for, from the viewer's side. */
export function statusText(view: PlayerView): string {
	const p = view.pending;
	if (!p || view.phase !== 'playing') return '';
	const mine = p.playerIds.includes(view.you);
	const actor = who(view, p.playerIds[0]);
	switch (p.step) {
		case 'lead':
			return mine ? 'Your lead: play any card.' : `${capital(actor)} leads…`;
		case 'follow': {
			const lead = view.trick?.lead;
			if (!mine) return `${capital(actor)} follows…`;
			if (lead?.ability === 'royalHeir') return 'Royal Heir led: play any card.';
			return lead ? `Follow ${SUIT_NAME[lead.suit]} if you can.` : 'Your turn.';
		}
		case 'foxesChoose':
			return mine ? 'Foxes: who may swap the decree?' : `${capital(actor)} plays the Foxes…`;
		case 'foxesSwap':
			return mine
				? 'Swap a card for the decree, or keep it.'
				: `${capital(actor)} may swap the decree…`;
		case 'gift':
			if (mine) return `Gift: choose a card for ${seats(view).mate?.name ?? 'your teammate'}.`;
			return view.pending?.yourGift
				? `Waiting for ${actor}'s gift…`
				: `${capital(actor)} chooses a gift…`;
		case 'move':
			return mine
				? 'You won the trick. Move the traveller:'
				: `${capital(actor)} moves the traveller…`;
		case 'forestEnd':
			return mine
				? 'Round over. Which end gets a forest token?'
				: `${capital(actor)} places a forest token…`;
	}
}

/** A short line about what just happened, built from the last events. */
export function describeEvents(events: GameEvent[], view: PlayerView): string {
	const parts: string[] = [];
	for (const e of events) {
		switch (e.type) {
			case 'roundStarted':
				parts.push(`Round ${e.round}: ${who(view, e.dealerId)} dealt`);
				break;
			case 'foxesChosen':
				parts.push(
					e.skipped
						? `Foxes: ${who(view, e.playerId)} had no card to swap`
						: `Foxes: ${who(view, e.byPlayerId)} chose ${who(view, e.playerId)}`
				);
				break;
			case 'decreeSwapped':
				parts.push(`New decree: ${cardLabel(e.new)}`);
				break;
			case 'decreeKept':
				parts.push(`${capital(who(view, e.playerId))} kept the decree`);
				break;
			case 'giftExchanged':
				parts.push('Gifts exchanged');
				break;
			case 'trickWon':
				parts.push(`${capital(who(view, e.playerId))} won trick ${e.number}`);
				break;
			case 'forestTokenPlaced':
				parts.push(
					e.roundEnd
						? `Forest token on ${endLabel(view, e.end)}`
						: `Off the path! Forest token on ${endLabel(view, e.end)}`
				);
				break;
			case 'gemCollected':
				parts.push(`Gem collected, ${e.left} left`);
				break;
			case 'gemsAdded':
				parts.push(`${plural(e.count, 'gem')} added`);
				break;
		}
	}
	return parts.join(' · ');
}
