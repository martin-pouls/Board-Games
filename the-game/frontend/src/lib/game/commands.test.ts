import { describe, expect, it } from 'vitest';
import { COMMANDS, commandInfo, targetCount, turnHints } from './commands.js';
import {
	extremeEightCommands,
	extremeMyTurn,
	extremeOthersNoComm,
	extremeSkull,
	finishedLostSkull,
	fixtures,
	noConstraints,
	playingMyTurn,
	playingOthersTurn
} from './fixtures.js';
import type { CommandName, PlayerView } from '#lib/types.js';

const names: CommandName[] = [
	'stop',
	'skull',
	'exactlyThree',
	'noCommunication',
	'noBackwardsTrick',
	'singlePile',
	'drawOnlyOne'
];

describe('COMMANDS', () => {
	it('covers every command and nothing else', () => {
		expect(Object.keys(COMMANDS).sort()).toEqual([...names].sort());
	});
	it.each(names)('has non-empty texts for %s', (name) => {
		const info = commandInfo(name);
		expect(info).toBe(COMMANDS[name]);
		expect(info.label.length).toBeGreaterThan(0);
		expect(info.short.length).toBeGreaterThan(0);
		expect(info.description.length).toBeGreaterThan(0);
	});
	it('classifies lightning and infinity commands', () => {
		const lightning = names.filter((n) => COMMANDS[n].kind === 'lightning');
		const infinity = names.filter((n) => COMMANDS[n].kind === 'infinity');
		expect(lightning).toEqual(['stop', 'skull', 'exactlyThree']);
		expect(infinity).toEqual(['noCommunication', 'noBackwardsTrick', 'singlePile', 'drawOnlyOne']);
	});
});

describe('targetCount', () => {
	it('is the required minimum without 3!', () => {
		expect(targetCount(playingMyTurn.turn!)).toBe(2);
		expect(targetCount({ ...playingMyTurn.turn!, requiredMinimum: 3 })).toBe(3);
	});
	it('is the exact count after 3!', () => {
		expect(targetCount(extremeSkull.turn!)).toBe(3);
		const turn = {
			...playingMyTurn.turn!,
			requiredMinimum: 1,
			constraints: { ...noConstraints, exactCount: 3 }
		};
		expect(targetCount(turn)).toBe(3);
	});
});

const texts = (v: PlayerView, mine: boolean) => turnHints(v, mine).map((h) => h.text);

describe('turnHints', () => {
	it('is empty for every classic fixture', () => {
		for (const v of Object.values(fixtures)) {
			if (v.options.variant !== 'classic') continue;
			expect(turnHints(v, true)).toEqual([]);
			expect(turnHints(v, false)).toEqual([]);
		}
	});
	it('is empty without a turn', () => {
		expect(turnHints(fixtures.lobbyHostExtreme, true)).toEqual([]);
	});
	it('asks to cover the skull first, then 3!', () => {
		expect(turnHints(extremeSkull, true)).toEqual([
			{ command: 'skull', text: 'Cover the skull before ending your turn' },
			{ command: 'exactlyThree', text: '3!: exactly 3 cards this turn' }
		]);
	});
	it("names the active player on others' turns", () => {
		const v = structuredClone(extremeSkull);
		v.turn!.activePlayerId = 'p_ben';
		expect(texts(v, false)).toEqual(['Ben must cover the skull', '3!: exactly 3 cards this turn']);
		expect(texts(finishedLostSkull, false)).toEqual(['Dan must cover the skull']);
	});
	it('falls back to "Someone" for an unknown active player', () => {
		const v = structuredClone(extremeSkull);
		v.turn!.activePlayerId = 'p_nobody';
		expect(texts(v, false)[0]).toBe('Someone must cover the skull');
	});
	it('explains a bound single pile and a visible No backwards trick', () => {
		expect(turnHints(extremeMyTurn, true)).toEqual([
			{ command: 'singlePile', text: 'Single pile: only the marked pile' },
			{ command: 'noBackwardsTrick', text: 'No backwards trick' }
		]);
	});
	it('says the first card picks an unbound single pile', () => {
		expect(texts(extremeEightCommands, true)).toEqual([
			'Single pile: the first card picks the pile',
			'No backwards trick'
		]);
	});
	it('mentions Draw only one, but not No communication (that has its own banner)', () => {
		expect(turnHints(extremeOthersNoComm, false)).toEqual([
			{ command: 'drawOnlyOne', text: 'Draw only 1 at turn end' }
		]);
	});
	it('ignores commands that are not visible', () => {
		const v = structuredClone(playingOthersTurn);
		v.activeCommands = [{ command: 'noCommunication', pile: 0 }];
		expect(turnHints(v, false)).toEqual([]);
	});
});
