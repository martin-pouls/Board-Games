import { describe, expect, it } from 'vitest';
import { COMMANDS, commandInfo, targetCount, turnStatus } from './commands.js';
import {
	extremeEightCommands,
	extremeMyTurn,
	extremeOthersNoComm,
	extremeSkull,
	fixtures,
	noConstraints,
	playingMyTurn,
	playingOthersTurn
} from './fixtures.js';
import type { CommandName } from '#lib/types.js';

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

describe('turnStatus', () => {
	it('is null without a turn', () => {
		expect(turnStatus(fixtures.lobbyHostExtreme(), true)).toBeNull();
	});
	it('has no Single pile tag and no Draw only one in any classic fixture', () => {
		for (const v of Object.values(fixtures).map((make) => make())) {
			if (v.options.variant !== 'classic' || !v.turn) continue;
			const s = turnStatus(v, true)!;
			expect(s.singlePile).toBeNull();
			expect(s.exactly).toBeNull();
			expect(s.drawOnlyOne).toBe(false);
		}
	});
	it('counts toward the minimum on your turn, plain count on others', () => {
		expect(turnStatus(playingMyTurn, true)!.line).toEqual({
			text: '1 of 2 played',
			tone: 'muted',
			dots: true
		});
		const done = structuredClone(playingMyTurn);
		done.turn!.cardsPlayed = 3;
		expect(turnStatus(done, true)!.line.text).toBe('3 played');
		expect(turnStatus(playingOthersTurn, false)!.line).toMatchObject({ dots: false });
	});
	it('badges 3! and counts toward it', () => {
		expect(turnStatus(extremeSkull, true)!.exactly).toBe(3);
		expect(turnStatus(extremeSkull, true)!.line).toEqual({
			text: '2 of 3 played',
			tone: 'muted',
			dots: true
		});
		const v = structuredClone(extremeSkull);
		v.turn!.activePlayerId = 'p_ben';
		expect(turnStatus(v, false)!.line).toEqual({
			text: '2 of 3 played',
			tone: 'muted',
			dots: false
		});
	});
	it('tags a single pile, bound or still to be picked', () => {
		expect(turnStatus(extremeMyTurn, true)!.singlePile).toEqual({
			bound: true,
			label: 'Only the marked pile',
			title: 'Single pile: all cards this turn on the marked pile'
		});
		expect(turnStatus(extremeEightCommands, true)!.singlePile).toEqual({
			bound: false,
			label: 'All cards on one pile',
			title: 'Single pile: the first card picks the pile for the whole turn'
		});
	});
	it('puts Draw only one on the End turn button, nothing else for Infinity commands', () => {
		const s = turnStatus(extremeOthersNoComm, false)!;
		expect(s.drawOnlyOne).toBe(true);
		expect(s.singlePile).toBeNull();
		expect(turnStatus(extremeMyTurn, true)!.drawOnlyOne).toBe(false);
	});
	it('replaces the count with STOP', () => {
		const v = structuredClone(extremeMyTurn);
		v.turn!.constraints.stopPlayed = true;
		expect(turnStatus(v, true)!.line).toEqual({
			text: 'STOP: end your turn',
			tone: 'warn',
			dots: false
		});
		v.turn!.activePlayerId = 'p_ben';
		expect(turnStatus(v, false)!.line.text).toBe('STOP: 1 played');
		v.turn!.constraints.exactCount = 3;
		expect(turnStatus(v, false)!.exactly).toBeNull();
	});
	it('says stuck in the same line, only to the stuck player', () => {
		const v = structuredClone(extremeSkull);
		v.turn!.stuck = true;
		expect(turnStatus(v, true)!.line).toEqual({
			text: 'Stuck: undo or give up',
			tone: 'warn',
			dots: false
		});
		v.turn!.activePlayerId = 'p_ben';
		expect(turnStatus(v, false)!.line.tone).toBe('muted');
	});
});
