import { describe, expect, it } from 'vitest';
import { parseLobbyOptions } from './lobbyOptions.js';

describe('parseLobbyOptions', () => {
	it('keeps valid options and drops unknown fields', () => {
		expect(parseLobbyOptions({ distress: true, old: 1 })).toEqual({ distress: true });
	});

	it('rejects unknown shapes', () => {
		for (const v of [null, [], { distress: 'yes' }, {}]) expect(parseLobbyOptions(v)).toBeNull();
	});
});
