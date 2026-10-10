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

/** Target dots for TurnBar: exactCount ?? requiredMinimum. */
export function targetCount(turn: Turn): number {
	return turn.constraints.exactCount ?? turn.requiredMinimum;
}

export interface TurnHint {
	command: CommandName;
	text: string;
}

/** Hints for TurnBar, from server facts only (constraints + activeCommands). */
export function turnHints(view: PlayerView, isMyTurn: boolean): TurnHint[] {
	const turn = view.turn;
	if (!turn) return [];
	const c = turn.constraints;
	const visible = (cmd: CommandName) => view.activeCommands.some((a) => a.command === cmd);
	const hints: TurnHint[] = [];
	const name = view.players.find((p) => p.id === turn.activePlayerId)?.name ?? 'Someone';

	if (c.stopPlayed) {
		hints.push({
			command: 'stop',
			text: isMyTurn ? 'STOP: no more cards, end your turn' : `${name} played STOP`
		});
	}
	if (c.skullPiles.length > 0) {
		hints.push({
			command: 'skull',
			text: isMyTurn ? 'Cover the skull before ending your turn' : `${name} must cover the skull`
		});
	}
	if (c.exactCount !== null) {
		hints.push({ command: 'exactlyThree', text: `3!: exactly ${c.exactCount} cards this turn` });
	}
	if (c.singlePile) {
		hints.push({
			command: 'singlePile',
			text:
				c.boundPile === null
					? 'Single pile: the first card picks the pile'
					: 'Single pile: only the marked pile'
		});
	}
	if (visible('noBackwardsTrick')) {
		hints.push({ command: 'noBackwardsTrick', text: 'No backwards trick' });
	}
	if (visible('drawOnlyOne')) {
		hints.push({ command: 'drawOnlyOne', text: 'Draw only 1 at turn end' });
	}
	return hints;
}
