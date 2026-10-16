import { describe, expect, it } from 'vitest';
import { parseLobbyOptions } from './lobbyOptions.js';

describe('parseLobbyOptions', () => {
	it('keeps valid options', () => {
		const o = { variant: 'extreme', pro: true, reducedHandSize: true };
		expect(parseLobbyOptions(o)).toEqual(o);
	});

	it('drops reduced hand size without Pro', () => {
		expect(parseLobbyOptions({ variant: 'classic', pro: false, reducedHandSize: true })).toEqual({
			variant: 'classic',
			pro: false,
			reducedHandSize: false
		});
	});

	it('rejects unknown shapes', () => {
		expect(parseLobbyOptions(null)).toBeNull();
		expect(parseLobbyOptions({ variant: 'hard', pro: true, reducedHandSize: false })).toBeNull();
		expect(
			parseLobbyOptions({ variant: 'classic', pro: 'yes', reducedHandSize: false })
		).toBeNull();
		expect(parseLobbyOptions({ variant: 'classic' })).toBeNull();
	});
});
