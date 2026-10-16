/**
 * Display metadata and turn hints for The Game Extreme command cards.
 * Pure TS (no Svelte imports): everything here renders server facts
 * (turn.constraints, activeCommands) and never re-derives rules.
 */
import type { CommandName, PlayerView, Turn } from '#lib/types.js';

/** Lightning: in effect for the rest of the turn it was played in. Infinity: while visible. */
export type CommandClass = 'lightning' | 'infinity';

export interface CommandInfo {
	kind: CommandClass;
	/** Full name, e.g. "No backwards trick". */
	label: string;
	/** Chip text, e.g. "No trick". */
	short: string;
	description: string;
}

export const COMMANDS: Record<CommandName, CommandInfo> = {
	stop: {
		kind: 'lightning',
		label: 'STOP',
		short: 'STOP',
		description: 'No more cards this turn: only End turn.'
	},
	skull: {
		kind: 'lightning',
		label: 'Skull',
		short: 'Skull',
		description: 'Must be covered in the same turn.'
	},
	exactlyThree: {
		kind: 'lightning',
		label: '3!',
		short: '3!',
		description: 'Exactly 3 cards this turn.'
	},
	noCommunication: {
		kind: 'infinity',
		label: 'No communication',
		short: 'Silence',
		description: 'No talking about the game while it is visible.'
	},
	noBackwardsTrick: {
		kind: 'infinity',
		label: 'No backwards trick',
		short: 'No trick',
		description: 'No backwards trick while it is visible.'
	},
	singlePile: {
		kind: 'infinity',
		label: 'Single pile',
		short: 'Single pile',
		description: 'All cards of a turn on one pile while it is visible.'
	},
	drawOnlyOne: {
		kind: 'infinity',
		label: 'Draw only one',
		short: 'Draw 1',
		description: 'Draw only 1 card at the end of the turn while it is visible.'
	}
};

export function commandInfo(c: CommandName): CommandInfo {
	return COMMANDS[c];
}

/** Cards a turn must reach: exactCount ?? requiredMinimum. */
export function targetCount(turn: Turn): number {
	return turn.constraints.exactCount ?? turn.requiredMinimum;
}

/**
 * What TurnBar shows besides the buttons, from server facts only. Each rule appears once on the
 * board: Infinity commands and the skull on their pile, the rest here. `line` is one line of text,
 * so the turn bar keeps its height and the piles don't move.
 */
export interface TurnStatus {
	/** Single pile is in force: a tag below the status line (the bound pile is also marked). */
	singlePile: { bound: boolean; label: string; title: string } | null;
	/** 3! is in force (and no STOP): a badge beside the Single pile tag; `line` counts toward it. */
	exactly: number | null;
	line: { text: string; tone: 'muted' | 'warn'; dots: boolean };
	/** Draw only one is visible: shown on the End turn button, where it takes effect. */
	drawOnlyOne: boolean;
}

export function turnStatus(view: PlayerView, isMyTurn: boolean): TurnStatus | null {
	const turn = view.turn;
	if (!turn) return null;
	const c = turn.constraints;
	const singlePile = c.singlePile
		? c.boundPile !== null
			? {
					bound: true,
					label: 'Only the marked pile',
					title: 'Single pile: all cards this turn on the marked pile'
				}
			: {
					bound: false,
					label: 'All cards on one pile',
					title: 'Single pile: the first card picks the pile for the whole turn'
				}
		: null;

	const played = turn.cardsPlayed;
	const target = targetCount(turn);
	let line: TurnStatus['line'];
	if (isMyTurn && turn.stuck) {
		line = { text: 'Stuck: undo or give up', tone: 'warn', dots: false };
	} else if (c.stopPlayed) {
		line = {
			text: isMyTurn ? 'STOP: end your turn' : `STOP: ${played} played`,
			tone: 'warn',
			dots: false
		};
	} else if (c.exactCount !== null) {
		line = { text: `${played} of ${c.exactCount} played`, tone: 'muted', dots: isMyTurn };
	} else if (isMyTurn && played < target) {
		line = { text: `${played} of ${target} played`, tone: 'muted', dots: true };
	} else {
		line = { text: `${played} played`, tone: 'muted', dots: isMyTurn };
	}

	return {
		singlePile,
		exactly: c.stopPlayed ? null : c.exactCount,
		line,
		drawOnlyOne: view.activeCommands.some((a) => a.command === 'drawOnlyOne')
	};
}
