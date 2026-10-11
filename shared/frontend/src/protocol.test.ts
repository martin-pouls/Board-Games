import { describe, expect, it } from 'vitest';
import {
	normalizeCode,
	nextBackoff,
	parseServerMessage,
	shouldAcceptView,
	wsUrl
} from './protocol.js';
import type { BaseView } from './wire.js';

const view = (revision: number) => ({ revision }) as BaseView;

describe('wsUrl', () => {
	it('uses ws for http and wss for https', () => {
		expect(wsUrl({ protocol: 'http:', host: 'localhost:5173' }, '', 'K7MQ2', 'tok')).toBe(
			'ws://localhost:5173/ws/games/K7MQ2?token=tok'
		);
		expect(wsUrl({ protocol: 'https:', host: 'x.example' }, '/regicide', 'K7MQ2', 'a b')).toBe(
			'wss://x.example/regicide/ws/games/K7MQ2?token=a%20b'
		);
	});
});

describe('nextBackoff', () => {
	it('grows exponentially within +-20% and caps at 10s', () => {
		expect(nextBackoff(0, () => 0.5)).toBe(500);
		expect(nextBackoff(1, () => 0.5)).toBe(1000);
		expect(nextBackoff(3, () => 0.5)).toBe(4000);
		expect(nextBackoff(10, () => 0.5)).toBe(10_000);
		expect(nextBackoff(0, () => 0)).toBeCloseTo(400);
		expect(nextBackoff(0, () => 1)).toBeCloseTo(600);
		expect(nextBackoff(50, () => 1)).toBeCloseTo(12_000);
	});
	it('stays in bounds with real randomness', () => {
		for (let a = 0; a < 12; a++) {
			const base = Math.min(500 * 2 ** a, 10_000);
			for (let i = 0; i < 20; i++) {
				const d = nextBackoff(a);
				expect(d).toBeGreaterThanOrEqual(base * 0.8);
				expect(d).toBeLessThanOrEqual(base * 1.2);
			}
		}
	});
});

describe('shouldAcceptView', () => {
	it('accepts first, equal and newer; rejects stale', () => {
		expect(shouldAcceptView(null, view(1))).toBe(true);
		expect(shouldAcceptView(view(3), view(3))).toBe(true);
		expect(shouldAcceptView(view(3), view(4))).toBe(true);
		expect(shouldAcceptView(view(3), view(2))).toBe(false);
	});
});

describe('parseServerMessage', () => {
	it('parses valid messages', () => {
		expect(parseServerMessage('{"type":"ack","requestId":"7"}')).toEqual({
			type: 'ack',
			requestId: '7'
		});
		expect(parseServerMessage('{"type":"presence","online":["p_1"]}')).toEqual({
			type: 'presence',
			online: ['p_1']
		});
		expect(
			parseServerMessage(
				'{"type":"error","requestId":"2","error":{"code":"illegal_placement","message":"x"}}'
			)
		).toEqual({
			type: 'error',
			requestId: '2',
			error: { code: 'illegal_placement', message: 'x' }
		});
		const v = parseServerMessage('{"type":"view","view":{"revision":2}}');
		expect(v).toEqual({ type: 'view', view: { revision: 2 }, events: [] });
	});
	it('returns null for garbage', () => {
		for (const s of [
			'',
			'nope',
			'null',
			'42',
			'[]',
			'{}',
			'{"type":"wat"}',
			'{"type":"view"}',
			'{"type":"ack"}',
			'{"type":"error"}'
		]) {
			expect(parseServerMessage(s)).toBeNull();
		}
	});
});

describe('normalizeCode', () => {
	it('uppercases and filters to the alphabet', () => {
		expect(normalizeCode('k7mq2')).toBe('K7MQ2');
		expect(normalizeCode('o0i1-a b')).toBe('AB');
		expect(normalizeCode('ABCDEFGH')).toBe('ABCDE');
	});
});
